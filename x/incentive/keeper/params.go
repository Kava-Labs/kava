package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/types"
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

// GetRewardPeriod returns the reward period with the specified collateral type if it's found
func (k Keeper) GetRewardPeriod(ctx sdk.Context, collateralType string) (types.RewardPeriod, bool) {
	params := k.GetParams(ctx)
	for _, rp := range params.RewardPeriods {
		if rp.CollateralType == collateralType {
			return rp, true
		}
	}
	return types.RewardPeriod{}, false
}
