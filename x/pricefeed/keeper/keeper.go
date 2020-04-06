package keeper

import (
	"fmt"
	"sort"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"

	"github.com/kava-labs/kava/x/pricefeed/types"
)

// Keeper struct for pricefeed module
type Keeper struct {
	// key used to access the stores from Context
	key sdk.StoreKey
	// Codec for binary encoding/decoding
	cdc *codec.Codec
	// The reference to the Paramstore to get and set pricefeed specific params
	paramSubspace subspace.Subspace
	// Reserved codespace
	codespace sdk.CodespaceType
}

// NewKeeper returns a new keeper for the pricefeed module.
func NewKeeper(
	cdc *codec.Codec, key sdk.StoreKey, paramSubspace subspace.Subspace, codespace sdk.CodespaceType,
) Keeper {
	return Keeper{
		paramSubspace: paramSubspace.WithKeyTable(types.ParamKeyTable()),
		key:           key,
		cdc:           cdc,
		codespace:     codespace,
	}
}

// SetPrice updates the posted price for a specific oracle
func (k Keeper) SetPrice(
	ctx sdk.Context,
	oracle sdk.AccAddress,
	marketID string,
	price sdk.Dec,
	expiry time.Time) (types.PostedPrice, sdk.Error) {
	// If the expiry is less than or equal to the current blockheight, we consider the price valid
	if expiry.After(ctx.BlockTime()) {
		store := ctx.KVStore(k.key)
		prices := k.GetRawPrices(ctx, marketID)
		var index int
		found := false
		for i := range prices {
			if prices[i].OracleAddress.Equals(oracle) {
				index = i
				found = true
				break
			}
		}
		// set the price for that particular oracle
		if found {
			prices[index] = types.NewPostedPrice(marketID, oracle, price, expiry)
		} else {
			prices = append(prices, types.NewPostedPrice(marketID, oracle, price, expiry))
			index = len(prices) - 1
		}

		// Emit an event containing the oracle's new price
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeOracleUpdatedPrice,
				sdk.NewAttribute(types.AttributeMarketID, marketID),
				sdk.NewAttribute(types.AttributeOracle, oracle.String()),
				sdk.NewAttribute(types.AttributeMarketPrice, price.String()),
				sdk.NewAttribute(types.AttributeExpiry, fmt.Sprintf("%d", expiry.Unix())),
			),
		)
		store.Set(
			types.RawPriceKey(marketID), k.cdc.MustMarshalBinaryBare(prices),
		)
		return prices[index], nil
	}
	return types.PostedPrice{}, types.ErrExpired(k.codespace)

}

// SetCurrentPrices updates the price of an asset to the median of all valid oracle inputs
func (k Keeper) SetCurrentPrices(ctx sdk.Context, marketID string) sdk.Error {
	_, ok := k.GetMarket(ctx, marketID)
	if !ok {
		return types.ErrInvalidMarket(k.codespace, marketID)
	}
	// store current price
	validPrevPrice := true
	prevPrice, err := k.GetCurrentPrice(ctx, marketID)
	if err != nil {
		validPrevPrice = false
	}

	prices := k.GetRawPrices(ctx, marketID)
	var notExpiredPrices types.CurrentPrices
	// filter out expired prices
	for _, v := range prices {
		if v.Expiry.After(ctx.BlockTime()) {
			notExpiredPrices = append(notExpiredPrices, types.NewCurrentPrice(v.MarketID, v.Price))
		}
	}
	if len(notExpiredPrices) == 0 {
		store := ctx.KVStore(k.key)
		store.Set(
			types.CurrentPriceKey(marketID), k.cdc.MustMarshalBinaryBare(types.CurrentPrice{}),
		)
		return types.ErrNoValidPrice(k.codespace)
	}
	medianPrice := k.CalculateMedianPrice(ctx, notExpiredPrices)

	// check case that market price was not set in genesis
	if validPrevPrice {
		// only emit event if price has changed
		if !medianPrice.Equal(prevPrice.Price) {
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeMarketPriceUpdated,
					sdk.NewAttribute(types.AttributeMarketID, fmt.Sprintf("%s", marketID)),
					sdk.NewAttribute(types.AttributeMarketPrice, fmt.Sprintf("%s", medianPrice.String())),
				),
			)
		}
	}

	store := ctx.KVStore(k.key)
	currentPrice := types.NewCurrentPrice(marketID, medianPrice)

	store.Set(
		types.CurrentPriceKey(marketID), k.cdc.MustMarshalBinaryBare(currentPrice),
	)

	return nil
}

// CalculateMedianPrice calculates the median prices for the input prices.
func (k Keeper) CalculateMedianPrice(ctx sdk.Context, prices types.CurrentPrices) sdk.Dec {
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
		median := k.calculateMeanPrice(ctx, prices[l/2-1:l/2+1])
		return median
	}
	// for odd numbers of prices, return the middle element
	return prices[l/2].Price

}

func (k Keeper) calculateMeanPrice(ctx sdk.Context, prices types.CurrentPrices) sdk.Dec {
	sum := prices[0].Price.Add(prices[1].Price)
	mean := sum.Quo(sdk.NewDec(2))
	return mean
}

// GetCurrentPrice fetches the current median price of all oracles for a specific market
func (k Keeper) GetCurrentPrice(ctx sdk.Context, marketID string) (types.CurrentPrice, sdk.Error) {
	store := ctx.KVStore(k.key)
	bz := store.Get(types.CurrentPriceKey(marketID))

	if bz == nil {
		return types.CurrentPrice{}, types.ErrNoValidPrice(k.codespace)
	}
	var price types.CurrentPrice
	k.cdc.MustUnmarshalBinaryBare(bz, &price)
	if price.Price.Equal(sdk.ZeroDec()) {
		return types.CurrentPrice{}, types.ErrNoValidPrice(k.codespace)
	}
	return price, nil
}

// GetRawPrices fetches the set of all prices posted by oracles for an asset
func (k Keeper) GetRawPrices(ctx sdk.Context, marketID string) types.PostedPrices {
	store := ctx.KVStore(k.key)
	bz := store.Get(types.RawPriceKey(marketID))
	var prices types.PostedPrices
	k.cdc.MustUnmarshalBinaryBare(bz, &prices)
	return prices
}

// Codespace return the codespace for the keeper
func (k Keeper) Codespace() sdk.CodespaceType {
	return k.codespace
}
