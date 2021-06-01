package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/swap/types"
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

// GetPairsWithDenom all pairs with the given token denom
func (k Keeper) GetPairsWithDenom(ctx sdk.Context, denom string) types.Pairs {
	params := k.GetParams(ctx)
	pairs := types.Pairs{}
	for _, p := range params.Pairs {
		if p.TokenA == denom || p.TokenB == denom {
			pairs = append(pairs, p)
		}
	}
	return pairs
}
