package keeper

import (
	"fmt"
	"sort"
	"time"

	"github.com/cometbft/cometbft/libs/log"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/kava-labs/kava/x/pricefeed/types"
)

// Keeper struct for pricefeed module
type Keeper struct {
	// key used to access the stores from Context
	key storetypes.StoreKey
	// Codec for binary encoding/decoding
	cdc codec.Codec
	// The reference to the Paramstore to get and set pricefeed specific params
	paramSubspace paramtypes.Subspace
}

// NewKeeper returns a new keeper for the pricefeed module.
func NewKeeper(
	cdc codec.Codec, key storetypes.StoreKey, paramstore paramtypes.Subspace,
) Keeper {
	if !paramstore.HasKeyTable() {
		paramstore = paramstore.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:           cdc,
		key:           key,
		paramSubspace: paramstore,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetPrice updates the posted price for a specific oracle
func (k Keeper) SetPrice(
	ctx sdk.Context,
	oracle sdk.AccAddress,
	marketID string,
	price sdk.Dec,
	expiry time.Time,
) (types.PostedPrice, error) {
	// If the expiry is less than or equal to the current blockheight, we consider the price valid
	if !expiry.After(ctx.BlockTime()) {
		return types.PostedPrice{}, types.ErrExpired
	}

	store := ctx.KVStore(k.key)

	newRawPrice := types.NewPostedPrice(marketID, oracle, price, expiry)

	// Emit an event containing the oracle's new price
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeOracleUpdatedPrice,
			sdk.NewAttribute(types.AttributeMarketID, marketID),
			sdk.NewAttribute(types.AttributeOracle, oracle.String()),
			sdk.NewAttribute(types.AttributeMarketPrice, price.String()),
			sdk.NewAttribute(types.AttributeExpiry, expiry.UTC().String()),
		),
	)

	// Sets the raw price for a single oracle instead of an array of all oracle's raw prices
	store.Set(types.RawPriceKey(marketID, oracle), k.cdc.MustMarshal(&newRawPrice))
	return newRawPrice, nil
}

// SetCurrentPrices updates the price of an asset to the median of all valid oracle inputs
func (k Keeper) SetCurrentPrices(ctx sdk.Context, marketID string) error {
	_, ok := k.GetMarket(ctx, marketID)
	if !ok {
		return errorsmod.Wrap(types.ErrInvalidMarket, marketID)
	}
	// store current price
	validPrevPrice := true
	prevPrice, err := k.GetCurrentPrice(ctx, marketID)
	if err != nil {
		validPrevPrice = false
	}

	prices := k.GetRawPrices(ctx, marketID)

	var notExpiredPrices []types.CurrentPrice
	// filter out expired prices
	for _, v := range prices {
		if v.Expiry.After(ctx.BlockTime()) {
			notExpiredPrices = append(notExpiredPrices, types.NewCurrentPrice(v.MarketID, v.Price))
		}
	}

	if len(notExpiredPrices) == 0 {
		// NOTE: The current price stored will continue storing the most recent (expired)
		// price if this is not set.
		// This zero's out the current price stored value for that market and ensures
		// that CDP methods that GetCurrentPrice will return error.
		k.setCurrentPrice(ctx, marketID, types.CurrentPrice{})
		return types.ErrNoValidPrice
	}

	medianPrice := k.CalculateMedianPrice(notExpiredPrices)

	// check case that market price was not set in genesis
	if validPrevPrice && !medianPrice.Equal(prevPrice.Price) {
		// only emit event if price has changed
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeMarketPriceUpdated,
				sdk.NewAttribute(types.AttributeMarketID, marketID),
				sdk.NewAttribute(types.AttributeMarketPrice, medianPrice.String()),
			),
		)
	}

	currentPrice := types.NewCurrentPrice(marketID, medianPrice)
	k.setCurrentPrice(ctx, marketID, currentPrice)

	return nil
}

// SetCurrentPricesForAllMarkets updates the price of an asset to the median of all valid oracle inputs
func (k Keeper) SetCurrentPricesForAllMarkets(ctx sdk.Context) {
	orderedMarkets := []string{}
	marketPricesByID := make(map[string]types.CurrentPrices)

	for _, market := range k.GetMarkets(ctx) {
		if market.Active {
			orderedMarkets = append(orderedMarkets, market.MarketID)
			marketPricesByID[market.MarketID] = types.CurrentPrices{}
		}
	}

	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.key), types.RawPriceFeedPrefix)
	for ; iterator.Valid(); iterator.Next() {
		var postedPrice types.PostedPrice
		k.cdc.MustUnmarshal(iterator.Value(), &postedPrice)

		prices, found := marketPricesByID[postedPrice.MarketID]
		if !found {
			continue
		}

		// filter out expired prices
		if postedPrice.Expiry.After(ctx.BlockTime()) {
			marketPricesByID[postedPrice.MarketID] = append(prices, types.NewCurrentPrice(postedPrice.MarketID, postedPrice.Price))
		}
	}
	iterator.Close()

	for _, marketID := range orderedMarkets {
		// store current price
		validPrevPrice := true
		prevPrice, err := k.GetCurrentPrice(ctx, marketID)
		if err != nil {
			validPrevPrice = false
		}

		notExpiredPrices, _ := marketPricesByID[marketID]

		if len(notExpiredPrices) == 0 {
			// NOTE: The current price stored will continue storing the most recent (expired)
			// price if this is not set.
			// This zero's out the current price stored value for that market and ensures
			// that CDP methods that GetCurrentPrice will return error.
			k.setCurrentPrice(ctx, marketID, types.CurrentPrice{})
			continue
		}

		medianPrice := k.CalculateMedianPrice(notExpiredPrices)

		// check case that market price was not set in genesis
		//if validPrevPrice && !medianPrice.Equal(prevPrice.Price) {
		if validPrevPrice && !medianPrice.Equal(prevPrice.Price) {
			// only emit event if price has changed
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeMarketPriceUpdated,
					sdk.NewAttribute(types.AttributeMarketID, marketID),
					sdk.NewAttribute(types.AttributeMarketPrice, medianPrice.String()),
				),
			)
		}

		currentPrice := types.NewCurrentPrice(marketID, medianPrice)
		k.setCurrentPrice(ctx, marketID, currentPrice)
	}
}

func (k Keeper) setCurrentPrice(ctx sdk.Context, marketID string, currentPrice types.CurrentPrice) {
	store := ctx.KVStore(k.key)
	store.Set(types.CurrentPriceKey(marketID), k.cdc.MustMarshal(&currentPrice))
}

// CalculateMedianPrice calculates the median prices for the input prices.
func (k Keeper) CalculateMedianPrice(prices []types.CurrentPrice) sdk.Dec {
	l := len(prices)

	if l == 1 {
		// Return immediately if there's only one price
		return prices[0].Price
	}
	// sort the prices
	sort.Slice(prices, func(i, j int) bool {
		return prices[i].Price.LT(prices[j].Price)
	})
	// for even numbers of prices, the median is calculated as the mean of the two middle prices
	if l%2 == 0 {
		median := k.calculateMeanPrice(prices[l/2-1], prices[l/2])
		return median
	}
	// for odd numbers of prices, return the middle element
	return prices[l/2].Price
}

func (k Keeper) calculateMeanPrice(priceA, priceB types.CurrentPrice) sdk.Dec {
	sum := priceA.Price.Add(priceB.Price)
	mean := sum.Quo(sdk.NewDec(2))
	return mean
}

// GetCurrentPrice fetches the current median price of all oracles for a specific market
func (k Keeper) GetCurrentPrice(ctx sdk.Context, marketID string) (types.CurrentPrice, error) {
	store := ctx.KVStore(k.key)
	bz := store.Get(types.CurrentPriceKey(marketID))

	if bz == nil {
		return types.CurrentPrice{}, types.ErrNoValidPrice
	}
	var price types.CurrentPrice
	err := k.cdc.Unmarshal(bz, &price)
	if err != nil {
		return types.CurrentPrice{}, err
	}
	if price.Price.Equal(sdk.ZeroDec()) {
		return types.CurrentPrice{}, types.ErrNoValidPrice
	}
	return price, nil
}

// IterateCurrentPrices iterates over all current price objects in the store and performs a callback function
func (k Keeper) IterateCurrentPrices(ctx sdk.Context, cb func(cp types.CurrentPrice) (stop bool)) {
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.key), types.CurrentPricePrefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var cp types.CurrentPrice
		k.cdc.MustUnmarshal(iterator.Value(), &cp)
		if cb(cp) {
			break
		}
	}
}

// GetCurrentPrices returns all current price objects from the store
func (k Keeper) GetCurrentPrices(ctx sdk.Context) types.CurrentPrices {
	var cps types.CurrentPrices
	k.IterateCurrentPrices(ctx, func(cp types.CurrentPrice) (stop bool) {
		cps = append(cps, cp)
		return false
	})
	return cps
}

// GetRawPrices fetches the set of all prices posted by oracles for an asset
func (k Keeper) GetRawPrices(ctx sdk.Context, marketId string) types.PostedPrices {
	var pps types.PostedPrices
	k.IterateRawPricesByMarket(ctx, marketId, func(pp types.PostedPrice) (stop bool) {
		pps = append(pps, pp)
		return false
	})
	return pps
}

// IterateRawPrices iterates over all raw prices in the store and performs a callback function
func (k Keeper) IterateRawPricesByMarket(ctx sdk.Context, marketId string, cb func(record types.PostedPrice) (stop bool)) {
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.key), types.RawPriceIteratorKey((marketId)))
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var record types.PostedPrice
		k.cdc.MustUnmarshal(iterator.Value(), &record)
		if cb(record) {
			break
		}
	}
}
