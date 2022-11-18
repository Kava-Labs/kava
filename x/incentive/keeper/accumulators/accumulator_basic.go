package accumulators

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/keeper/adapters"
	"github.com/kava-labs/kava/x/incentive/keeper/store"
	"github.com/kava-labs/kava/x/incentive/types"
)

// BasicAccumulator is a default implementation of the RewardAccumulator
// interface. This applies to all claim types except for those with custom
// accumulator logic e.g. Earn.
type BasicAccumulator struct {
	store    store.IncentiveStore
	adapters adapters.SourceAdapters
}

var _ types.RewardAccumulator = BasicAccumulator{}

// NewBasicAccumulator returns a new BasicAccumulator.
func NewBasicAccumulator(
	store store.IncentiveStore,
	adapters adapters.SourceAdapters,
) BasicAccumulator {
	return BasicAccumulator{
		store:    store,
		adapters: adapters,
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
	previousAccrualTime, found := k.store.GetRewardAccrualTime(ctx, claimType, rewardPeriod.CollateralType)
	if !found {
		previousAccrualTime = ctx.BlockTime()
	}

	indexes, found := k.store.GetRewardIndexesOfClaimType(ctx, claimType, rewardPeriod.CollateralType)
	if !found {
		indexes = types.RewardIndexes{}
	}

	acc := types.NewAccumulator(previousAccrualTime, indexes)

	totalSource := k.adapters.TotalSharesBySource(ctx, claimType, rewardPeriod.CollateralType)

	acc.Accumulate(rewardPeriod, totalSource, ctx.BlockTime())

	k.store.SetRewardAccrualTime(ctx, claimType, rewardPeriod.CollateralType, acc.PreviousAccumulationTime)
	if len(acc.Indexes) > 0 {
		// the store panics when setting empty or nil indexes
		k.store.SetRewardIndexes(ctx, claimType, rewardPeriod.CollateralType, acc.Indexes)
	}

	return nil
}
