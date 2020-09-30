package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/harvest/types"
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

func (k Keeper) GetLPSchedule(ctx sdk.Context, denom string) (types.DistributionSchedule, bool) {
	params := k.GetParams(ctx)
	for _, lps := range params.LiquidityProviderSchedules {
		if lps.DepositDenom == denom {
			return lps, true
		}
	}
	return types.DistributionSchedule{}, false
}

func (k Keeper) GetDelegatorSchedule(ctx sdk.Context, denom string) (types.DelegatorDistributionSchedule, bool) {
	params := k.GetParams(ctx)
	for _, dds := range params.DelegatorDistributionSchedules {
		if dds.DistributionSchedule.DepositDenom == denom {
			return dds, true
		}
	}
	return types.DelegatorDistributionSchedule{}, false
}
