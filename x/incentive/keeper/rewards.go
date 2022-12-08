package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/incentive/keeper/accumulators"
	"github.com/kava-labs/kava/x/incentive/types"
)

// AccumulateRewards calculates new rewards to distribute this block and updates the global indexes to reflect this.
// The provided rewardPeriod must be valid to avoid panics in calculating time durations.
func (k Keeper) AccumulateRewards(
	ctx sdk.Context,
	claimType types.ClaimType,
	rewardPeriod types.MultiRewardPeriod,
) error {
	var accumulator types.RewardAccumulator

	switch claimType {
	case types.CLAIM_TYPE_EARN:
		accumulator = accumulators.NewEarnAccumulator(k.Store, k.liquidKeeper, k.earnKeeper, k.Adapters)
	default:
		accumulator = accumulators.NewBasicAccumulator(k.Store, k.Adapters)
	}

	return accumulator.AccumulateRewards(ctx, claimType, rewardPeriod)
}

// InitializeClaim creates a new claim with zero rewards and indexes matching
// the global indexes. If the claim already exists it just updates the indexes.
func (k Keeper) InitializeClaim(
	ctx sdk.Context,
	claimType types.ClaimType,
	owner sdk.AccAddress,
	sourceIDs []string,
) {
	claim, found := k.Store.GetClaim(ctx, claimType, owner)
	if !found {
		claim = types.NewClaim(claimType, owner, sdk.Coins{}, nil)
	}

	claim = k.initializeClaim(ctx, claim, sourceIDs)
	k.Store.SetClaim(ctx, claim)
}

// initializeClaim updates an existing claim's specified reward indexes to match
// the global indexes.
func (k Keeper) initializeClaim(
	ctx sdk.Context,
	claim types.Claim,
	sourceIDs []string,
) types.Claim {
	for _, sourceID := range sourceIDs {
		globalRewardIndexes, found := k.Store.GetRewardIndexesOfClaimType(ctx, claim.Type, sourceID)
		if !found {
			globalRewardIndexes = types.RewardIndexes{}
		}

		claim.RewardIndexes = claim.RewardIndexes.With(sourceID, globalRewardIndexes)
	}

	return claim
}

// InitializeClaimSingleReward creates a new claim with zero rewards and indexes matching
// the global indexes. If the claim already exists it just updates the indexes.
func (k Keeper) InitializeClaimSingleReward(
	ctx sdk.Context,
	claimType types.ClaimType,
	owner sdk.AccAddress,
	sourceID string,
) {
	claim, found := k.Store.GetClaim(ctx, claimType, owner)
	if !found {
		claim = types.NewClaim(claimType, owner, sdk.Coins{}, nil)
	}

	globalRewardIndexes, found := k.Store.GetRewardIndexesOfClaimType(ctx, claimType, sourceID)
	if !found {
		globalRewardIndexes = types.RewardIndexes{}
	}

	claim.RewardIndexes = claim.RewardIndexes.With(sourceID, globalRewardIndexes)
	k.Store.SetClaim(ctx, claim)
}

// SynchronizeClaim updates the claim object same as SynchronizeClaimSingleReward,
// but with multiple share coins.
func (k Keeper) SynchronizeClaim(
	ctx sdk.Context,
	claimType types.ClaimType,
	owner sdk.AccAddress,
	shareCoins sdk.DecCoins,
	initializeSourceIDs []string,
) {
	claim, found := k.Store.GetClaim(ctx, claimType, owner)
	if !found {
		return
	}

	for _, coin := range shareCoins {
		claim = k.synchronizeClaimSingleReward(ctx, claim, coin.Denom, coin.Amount)
	}

	// Does nothing if initializeSourceIDs is empty
	claim = k.initializeClaim(ctx, claim, initializeSourceIDs)

	// Prune any rewards that are no longer active in the source.
	// e.g. withdrawn deposits.
	activeDenoms := append(getDecCoinsDenoms(shareCoins), initializeSourceIDs...)
	claim = k.PruneClaimRewards(
		ctx,
		claim,
		activeDenoms,
	)

	k.Store.SetClaim(ctx, claim)
}

// SynchronizeClaimSingleReward updates the claim object by adding any
// accumulated rewards and updating the reward index value.
func (k Keeper) SynchronizeClaimSingleReward(
	ctx sdk.Context,
	claimType types.ClaimType,
	owner sdk.AccAddress,
	sourceID string,
	shares sdk.Dec,
) {
	claim, found := k.Store.GetClaim(ctx, claimType, owner)
	if !found {
		return
	}

	claim = k.synchronizeClaimSingleReward(ctx, claim, sourceID, shares)
	k.Store.SetClaim(ctx, claim)
}

// synchronizeClaimSingleReward updates the reward and indexes in a claim for one sourceID.
func (k *Keeper) synchronizeClaimSingleReward(
	ctx sdk.Context,
	claim types.Claim,
	sourceID string,
	shares sdk.Dec,
) types.Claim {
	globalRewardIndexes, found := k.Store.GetRewardIndexesOfClaimType(ctx, claim.Type, sourceID)
	if !found {
		// The global factor is only not found if
		// - the pool has not started accumulating rewards yet (either there is no reward specified in params, or the reward start time hasn't been hit)
		// - OR it was wrongly deleted from state (factors should never be removed while unsynced claims exist)
		// If not found we could either skip this sync, or assume the global factor is zero.
		// Skipping will avoid storing unnecessary factors in the claim for non rewarded pools.
		// And in the event a global factor is wrongly deleted, it will avoid this function panicking when calculating rewards.
		return claim
	}

	userRewardIndexes, found := claim.RewardIndexes.Get(sourceID)
	if !found {
		// Normally the reward indexes should always be found.
		// But if a pool was not rewarded then becomes rewarded (ie a reward period is added to params), then the indexes will be missing from claims for that pool.
		// So given the reward period was just added, assume the starting value for any global reward indexes, which is an empty slice.
		userRewardIndexes = types.RewardIndexes{}
	}

	newRewards, err := k.CalculateRewards(userRewardIndexes, globalRewardIndexes, shares)
	if err != nil {
		// Global reward factors should never decrease, as it would lead to a negative update to claim.Rewards.
		// This panics if a global reward factor decreases or disappears between the old and new indexes.
		panic(fmt.Sprintf("corrupted global reward indexes found: %v", err))
	}

	claim.Reward = claim.Reward.Add(newRewards...)
	claim.RewardIndexes = claim.RewardIndexes.With(sourceID, globalRewardIndexes)

	return claim
}

// GetSynchronizedClaim fetches a claim from the store and syncs rewards for all
// rewarded sourceIDs.
func (k Keeper) GetSynchronizedClaim(
	ctx sdk.Context,
	claimType types.ClaimType,
	owner sdk.AccAddress,
) (types.Claim, bool) {
	claim, found := k.Store.GetClaim(ctx, claimType, owner)
	if !found {
		return types.Claim{}, false
	}

	// Fetch all source IDs from indexes
	var sourceIDs []string
	k.Store.IterateRewardIndexesByClaimType(ctx, claimType, func(rewardIndexes types.TypedRewardIndexes) bool {
		sourceIDs = append(sourceIDs, rewardIndexes.CollateralType)
		return false
	})

	accShares := k.Adapters.OwnerSharesBySource(ctx, claimType, owner, sourceIDs)

	// Synchronize claim for each source ID
	for _, share := range accShares {
		claim = k.synchronizeClaimSingleReward(ctx, claim, share.ID, share.Shares)
	}

	return claim, true
}

func (k Keeper) PruneClaimRewards(
	ctx sdk.Context,
	claim types.Claim,
	activeSourceIDs []string,
) types.Claim {
	claimIndexDenoms := claim.RewardIndexes.GetCollateralTypes()

	// claimIndexDenoms - activeSourceIDs
	inactiveSourceIDs := setDifference(claimIndexDenoms, activeSourceIDs)

	// Remove rewards that aren't contained in the active source IDs, e.g.
	// assets that are no longer deposited.
	for _, denom := range inactiveSourceIDs {
		claim.RewardIndexes = claim.RewardIndexes.RemoveRewardIndex(denom)
	}

	return claim
}
