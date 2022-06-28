package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/types"
)

// AccumulateSwapRewards calculates new rewards to distribute this block and updates the global indexes to reflect this.
// The provided rewardPeriod must be valid to avoid panics in calculating time durations.
func (k Keeper) AccumulateSwapRewards(ctx sdk.Context, rewardPeriod types.MultiRewardPeriod) {
	previousAccrualTime, found := k.GetSwapRewardAccrualTime(ctx, rewardPeriod.CollateralType)
	if !found {
		previousAccrualTime = ctx.BlockTime()
	}

	indexes, found := k.GetGlobalIndexes(ctx, types.RewardTypeSwap, rewardPeriod.CollateralType)
	if !found {
		indexes = types.RewardIndexes{}
	}

	acc := types.NewAccumulator(previousAccrualTime, indexes)

	totalSource := k.getSwapTotalSourceShares(ctx, rewardPeriod.CollateralType)

	acc.Accumulate(rewardPeriod, totalSource, ctx.BlockTime())

	k.SetSwapRewardAccrualTime(ctx, rewardPeriod.CollateralType, acc.PreviousAccumulationTime)
	if len(acc.Indexes) > 0 {
		// the store panics when setting empty or nil indexes
		k.SetGlobalIndexes(ctx, types.RewardTypeSwap, rewardPeriod.CollateralType, acc.Indexes)
	}
}

// getSwapTotalSourceShares fetches the sum of all source shares for a swap reward.
// In the case of swap, these are the total (swap module) shares in a particular pool.
func (k Keeper) getSwapTotalSourceShares(ctx sdk.Context, poolID string) sdk.Dec {
	totalShares, found := k.swapKeeper.GetPoolShares(ctx, poolID)
	if !found {
		totalShares = sdk.ZeroInt()
	}
	return totalShares.ToDec()
}

// InitializeSwapReward creates a new claim with zero rewards and indexes matching the global indexes.
// If the claim already exists it just updates the indexes.
func (k Keeper) InitializeReward(ctx sdk.Context, rewardType types.RewardType, poolID string, owner sdk.AccAddress) {
	claim, found := k.GetClaim(ctx, rewardType, owner)
	if !found {
		claim = types.NewClaim(owner, sdk.Coins{}, nil)
	}

	globalRewardIndexes, found := k.GetGlobalIndexes(ctx, rewardType, poolID)
	if !found {
		globalRewardIndexes = types.RewardIndexes{}
	}
	claim.RewardIndexes = claim.RewardIndexes.With(poolID, globalRewardIndexes)

	k.SetClaim(ctx, rewardType, claim)
}

// SynchronizeSwapReward updates the claim object by adding any accumulated rewards
// and updating the reward index value.
func (k Keeper) SynchronizeReward(ctx sdk.Context, rewardID types.RewardType, poolID string, owner sdk.AccAddress, shares sdk.Int) {
	claim, found := k.GetClaim(ctx, rewardID, owner)
	if !found {
		return
	}
	claim = k.synchronizeReward(ctx, claim, rewardID, poolID, owner, shares)

	k.SetClaim(ctx, rewardID, claim)
}

// synchronizeSwapReward updates the reward and indexes in a swap claim for one pool.
func (k *Keeper) synchronizeReward(ctx sdk.Context, claim types.Claim, rewardType types.RewardType, poolID string, owner sdk.AccAddress, shares sdk.Int) types.Claim {
	globalRewardIndexes, found := k.GetGlobalIndexes(ctx, rewardType, poolID)
	if !found {
		// The global factor is only not found if
		// - the pool has not started accumulating rewards yet (either there is no reward specified in params, or the reward start time hasn't been hit)
		// - OR it was wrongly deleted from state (factors should never be removed while unsynced claims exist)
		// If not found we could either skip this sync, or assume the global factor is zero.
		// Skipping will avoid storing unnecessary factors in the claim for non rewarded pools.
		// And in the event a global factor is wrongly deleted, it will avoid this function panicking when calculating rewards.
		return claim
	}

	userRewardIndexes, found := claim.RewardIndexes.Get(poolID)
	if !found {
		// Normally the reward indexes should always be found.
		// But if a pool was not rewarded then becomes rewarded (ie a reward period is added to params), then the indexes will be missing from claims for that pool.
		// So given the reward period was just added, assume the starting value for any global reward indexes, which is an empty slice.
		userRewardIndexes = types.RewardIndexes{}
	}

	newRewards, err := k.CalculateRewards(userRewardIndexes, globalRewardIndexes, shares.ToDec())
	if err != nil {
		// Global reward factors should never decrease, as it would lead to a negative update to claim.Rewards.
		// This panics if a global reward factor decreases or disappears between the old and new indexes.
		panic(fmt.Sprintf("corrupted global reward indexes found: %v", err))
	}

	claim.Reward = claim.Reward.Add(newRewards...)
	claim.RewardIndexes = claim.RewardIndexes.With(poolID, globalRewardIndexes)

	return claim
}

// GetSynchronizedSwapClaim fetches a swap claim from the store and syncs rewards for all rewarded pools.
func (k Keeper) GetSynchronizedClaim(ctx sdk.Context, rewardID types.RewardType, owner sdk.AccAddress) (types.Claim, bool) {
	claim, found := k.GetClaim(ctx, rewardID, owner)
	if !found {
		return types.Claim{}, false
	}

	// TODO will iterating global indexes ok to find all sources?
	k.IterateGlobalIndexes(ctx, rewardID, func(poolID string, _ types.RewardIndexes) bool {

		shares := k.GetSourceShares(ctx, rewardID, poolID, owner)

		claim = k.synchronizeReward(ctx, claim, rewardID, poolID, owner, shares)

		return false
	})

	return claim, true
}

// ---------------

// This could be functions registered on the keeper instead of a big switch

func (k Keeper) GetSourceShares(ctx sdk.Context, rewardType types.RewardType, rewardID string, owner sdk.AccAddress) sdk.Int {
	switch rewardType {
	case types.RewardTypeSwap:
		shares, found := k.swapKeeper.GetDepositorSharesAmount(ctx, owner, rewardID)
		if !found {
			shares = sdk.ZeroInt()
		}
		return shares
	// TODO add other reward types
	default:
		panic("unknown")
	}
}
