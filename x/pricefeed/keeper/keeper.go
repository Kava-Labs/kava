package keeper

import (
	"sort"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"

	"github.com/kava-labs/kava/x/pricefeed/types"
)

// Keeper struct for pricefeed module
type Keeper struct {
	// The keys used to access the stores from Context
	storeKey sdk.StoreKey
	// Codec for binary encoding/decoding
	cdc *codec.Codec
	// The reference to the Paramstore to get and set pricefeed specific params
	paramstore params.Subspace
	// Reserved codespace
	codespace sdk.CodespaceType
}

// NewKeeper returns a new keeper for the pricefeed module. It handles:
// - adding oracles
// - adding/removing assets from the pricefeed
func NewKeeper(
	storeKey sdk.StoreKey, cdc *codec.Codec, paramstore params.Subspace, codespace sdk.CodespaceType,
) Keeper {
	return Keeper{
		paramstore: paramstore.WithKeyTable(types.ParamKeyTable()),
		storeKey:   storeKey,
		cdc:        cdc,
		codespace:  codespace,
	}
}

// SetPrice updates the posted price for a specific oracle
func (k Keeper) SetPrice(
	ctx sdk.Context,
	oracle sdk.AccAddress,
	assetCode string,
	price sdk.Dec,
	expiry time.Time) (types.PostedPrice, sdk.Error) {
	// If the expiry is less than or equal to the current blockheight, we consider the price valid
	if expiry.After(ctx.BlockTime()) {
		store := ctx.KVStore(k.storeKey)
		prices := k.GetRawPrices(ctx, assetCode)
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
			prices[index] = types.PostedPrice{
				AssetCode: assetCode, OracleAddress: oracle,
				Price: price, Expiry: expiry}
		} else {
			prices = append(prices, types.PostedPrice{
				AssetCode: assetCode, OracleAddress: oracle,
				Price: price, Expiry: expiry})
			index = len(prices) - 1
		}

		store.Set(
			[]byte(types.RawPriceFeedPrefix+assetCode), k.cdc.MustMarshalBinaryBare(prices),
		)
		return prices[index], nil
	}
	return types.PostedPrice{}, types.ErrExpired(k.codespace)

}

// SetCurrentPrices updates the price of an asset to the meadian of all valid oracle inputs
func (k Keeper) SetCurrentPrices(ctx sdk.Context, assetCode string) sdk.Error {
	_, ok := k.GetAsset(ctx, assetCode)
	if !ok {
		return types.ErrInvalidAsset(k.codespace)
	}
	prices := k.GetRawPrices(ctx, assetCode)
	var notExpiredPrices []types.CurrentPrice
	// filter out expired prices
	for _, v := range prices {
		if v.Expiry.After(ctx.BlockTime()) {
			notExpiredPrices = append(notExpiredPrices, types.CurrentPrice{
				AssetCode: v.AssetCode,
				Price:     v.Price,
			})
		}
	}
	l := len(notExpiredPrices)
	var medianPrice sdk.Dec
	// TODO make threshold for acceptance (ie. require 51% of oracles to have posted valid prices
	if l == 0 {
		// Error if there are no valid prices in the raw pricefeed
		return types.ErrNoValidPrice(k.codespace)
	} else if l == 1 {
		// Return immediately if there's only one price
		medianPrice = notExpiredPrices[0].Price
	} else {
		// sort the prices
		sort.Slice(notExpiredPrices, func(i, j int) bool {
			return notExpiredPrices[i].Price.LT(notExpiredPrices[j].Price)
		})
		// If there's an even number of prices
		if l%2 == 0 {
			// TODO make sure this is safe.
			// Since it's a price and not a balance, division with precision loss is OK.
			price1 := notExpiredPrices[l/2-1].Price
			price2 := notExpiredPrices[l/2].Price
			sum := price1.Add(price2)
			divsor, _ := sdk.NewDecFromStr("2")
			medianPrice = sum.Quo(divsor)
		} else {
			// integer division, so we'll get an integer back, rounded down
			medianPrice = notExpiredPrices[l/2].Price
		}
	}

	store := ctx.KVStore(k.storeKey)
	currentPrice := types.CurrentPrice{
		AssetCode: assetCode,
		Price:     medianPrice,
	}
	store.Set(
		[]byte(types.CurrentPricePrefix+assetCode), k.cdc.MustMarshalBinaryBare(currentPrice),
	)

	return nil
}

// GetCurrentPrice fetches the current median price of all oracles for a specific asset
func (k Keeper) GetCurrentPrice(ctx sdk.Context, assetCode string) types.CurrentPrice {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(types.CurrentPricePrefix + assetCode))
	// TODO panic or return error if not found
	var price types.CurrentPrice
	k.cdc.MustUnmarshalBinaryBare(bz, &price)
	return price
}

// GetRawPrices fetches the set of all prices posted by oracles for an asset
func (k Keeper) GetRawPrices(ctx sdk.Context, assetCode string) []types.PostedPrice {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(types.RawPriceFeedPrefix + assetCode))
	var prices []types.PostedPrice
	k.cdc.MustUnmarshalBinaryBare(bz, &prices)
	return prices
}

func (k Keeper) Codespace() sdk.CodespaceType {
	return k.codespace
}
