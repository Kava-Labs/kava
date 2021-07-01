package keeper

import (
	"time"

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

// GetUSDXMintingRewardPeriod returns the reward period with the specified collateral type if it's found in the params
func (k Keeper) GetUSDXMintingRewardPeriod(ctx sdk.Context, collateralType string) (types.RewardPeriod, bool) {
	params := k.GetParams(ctx)
	for _, rp := range params.USDXMintingRewardPeriods {
		if rp.CollateralType == collateralType {
			return rp, true
		}
	}
	return types.RewardPeriod{}, false
}

// GetHardSupplyRewardPeriods returns the reward period with the specified collateral type if it's found in the params
func (k Keeper) GetHardSupplyRewardPeriods(ctx sdk.Context, denom string) (types.MultiRewardPeriod, bool) {
	params := k.GetParams(ctx)
	for _, rp := range params.HardSupplyRewardPeriods {
		if rp.CollateralType == denom {
			return rp, true
		}
	}
	return types.MultiRewardPeriod{}, false
}

// GetHardBorrowRewardPeriods returns the reward period with the specified collateral type if it's found in the params
func (k Keeper) GetHardBorrowRewardPeriods(ctx sdk.Context, denom string) (types.MultiRewardPeriod, bool) {
	params := k.GetParams(ctx)
	for _, rp := range params.HardBorrowRewardPeriods {
		if rp.CollateralType == denom {
			return rp, true
		}
	}
	return types.MultiRewardPeriod{}, false
}

// GetHardDelegatorRewardPeriod returns the reward period with the specified collateral type if it's found in the params
func (k Keeper) GetHardDelegatorRewardPeriods(ctx sdk.Context, denom string) (types.MultiRewardPeriod, bool) {
	params := k.GetParams(ctx)
	for _, rp := range params.HardDelegatorRewardPeriods {
		if rp.CollateralType == denom {
			return rp, true
		}
	}
	return types.MultiRewardPeriod{}, false
}

// GetMultiplier returns the multiplier with the specified name if it's found in the params
func (k Keeper) GetMultiplier(ctx sdk.Context, name types.MultiplierName) (types.Multiplier, bool) {
	params := k.GetParams(ctx)
	for _, m := range params.ClaimMultipliers {
		if m.Name == name {
			return m, true
		}
	}
	return types.Multiplier{}, false
}

// GetClaimEnd returns the claim end time for the params
func (k Keeper) GetClaimEnd(ctx sdk.Context) time.Time {
	params := k.GetParams(ctx)
	return params.ClaimEnd
}
