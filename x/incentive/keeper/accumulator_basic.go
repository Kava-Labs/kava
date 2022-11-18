package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/types"
)

// BasicAccumulator is a default implementation of the RewardAccumulator
// interface. This applies to all claim types except for those with custom
// accumulator logic e.g. Earn.
type BasicAccumulator struct {
	keeper Keeper
}

var _ types.RewardAccumulator = BasicAccumulator{}

// NewBasicAccumulator returns a new BasicAccumulator.
func NewBasicAccumulator(k Keeper) BasicAccumulator {
	return BasicAccumulator{
		keeper: k,
	}
}

// AccumulateRewards calculates new rewards to distribute this block and updates
// the global indexes to reflect this. The provided rewardPeriod must be valid
// to avoid panics in calculating time durations.
func (k BasicAccumulator) AccumulateRewards(
	ctx sdk.Context,
	claimType types.ClaimType,
	rewardPeriod types.MultiRewardPeriod,
) error {
	previousAccrualTime, found := k.keeper.GetRewardAccrualTime(ctx, claimType, rewardPeriod.CollateralType)
	if !found {
		previousAccrualTime = ctx.BlockTime()
	}

	indexes, found := k.keeper.GetRewardIndexesOfClaimType(ctx, claimType, rewardPeriod.CollateralType)
	if !found {
		indexes = types.RewardIndexes{}
	}

	acc := types.NewAccumulator(previousAccrualTime, indexes)

	totalSource := k.keeper.Adapters.TotalSharesBySource(ctx, claimType, rewardPeriod.CollateralType)

	acc.Accumulate(rewardPeriod, totalSource, ctx.BlockTime())

	k.keeper.SetRewardAccrualTime(ctx, claimType, rewardPeriod.CollateralType, acc.PreviousAccumulationTime)
	if len(acc.Indexes) > 0 {
		// the store panics when setting empty or nil indexes
		k.keeper.SetRewardIndexes(ctx, claimType, rewardPeriod.CollateralType, acc.Indexes)
	}

	return nil
}
