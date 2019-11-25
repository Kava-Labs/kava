package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/liquidator/types"
)

// GetParams returns the params for liquidator module
func (k Keeper) GetParams(ctx sdk.Context) types.LiquidatorParams {
	var params types.LiquidatorParams
	k.paramSubspace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets params for the liquidator module
func (k Keeper) SetParams(ctx sdk.Context, params types.LiquidatorParams) {
	k.paramSubspace.SetParamSet(ctx, &params)
}