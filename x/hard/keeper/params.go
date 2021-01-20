package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/hard/types"
)

// GetParams returns the params from the store
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	var p types.Params
	k.paramSubspace.GetParamSet(ctx, &p)
	return p
}

// SetParams sets params on the store
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSubspace.SetParamSet(ctx, &params)
}

// GetMoneyMarketParam returns the corresponding Money Market param for a specific denom
func (k Keeper) GetMoneyMarketParam(ctx sdk.Context, denom string) (types.MoneyMarket, bool) {
	params := k.GetParams(ctx)
	for _, mm := range params.MoneyMarkets {
		if mm.Denom == denom {
			return mm, true
		}
	}
	return types.MoneyMarket{}, false
}
