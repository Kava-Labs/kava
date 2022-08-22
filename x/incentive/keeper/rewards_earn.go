package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	earntypes "github.com/kava-labs/kava/x/earn/types"
	"github.com/kava-labs/kava/x/incentive/types"
)

// AccumulateEarnRewards calculates new rewards to distribute this block and updates the global indexes to reflect this.
// The provided rewardPeriod must be valid to avoid panics in calculating time durations.
func (k Keeper) AccumulateEarnRewards(ctx sdk.Context, rewardPeriod types.MultiRewardPeriod) {
	previousAccrualTime, found := k.GetEarnRewardAccrualTime(ctx, rewardPeriod.CollateralType)
	if !found {
		previousAccrualTime = ctx.BlockTime()
	}

	indexes, found := k.GetEarnRewardIndexes(ctx, rewardPeriod.CollateralType)
	if !found {
		indexes = types.RewardIndexes{}
	}

	acc := types.NewAccumulator(previousAccrualTime, indexes)

	totalSource := k.getEarnTotalSourceShares(ctx, rewardPeriod.CollateralType)

	acc.Accumulate(rewardPeriod, totalSource, ctx.BlockTime())

	k.SetEarnRewardAccrualTime(ctx, rewardPeriod.CollateralType, acc.PreviousAccumulationTime)
	if len(acc.Indexes) > 0 {
		// the store panics when setting empty or nil indexes
		k.SetEarnRewardIndexes(ctx, rewardPeriod.CollateralType, acc.Indexes)
	}
}

// getEarnTotalSourceShares fetches the sum of all source shares for a earn reward.
// In the case of earn, these are the total (earn module) shares in a particular vault.
func (k Keeper) getEarnTotalSourceShares(ctx sdk.Context, vaultDenom string) sdk.Dec {
	totalShares, found := k.earnKeeper.GetVaultTotalShares(ctx, vaultDenom)
	if !found {
		return sdk.ZeroDec()
	}
	return totalShares.Amount
}

// InitializeEarnReward creates a new claim with zero rewards and indexes matching the global indexes.
// If the claim already exists it just updates the indexes.
func (k Keeper) InitializeEarnReward(ctx sdk.Context, vaultDenom string, owner sdk.AccAddress) {
	claim, found := k.GetEarnClaim(ctx, owner)
	if !found {
		claim = types.NewEarnClaim(owner, sdk.Coins{}, nil)
	}

	globalRewardIndexes, found := k.GetEarnRewardIndexes(ctx, vaultDenom)
	if !found {
		globalRewardIndexes = types.RewardIndexes{}
	}
	claim.RewardIndexes = claim.RewardIndexes.With(vaultDenom, globalRewardIndexes)

	k.SetEarnClaim(ctx, claim)
}

// SynchronizeEarnReward updates the claim object by adding any accumulated rewards
// and updating the reward index value.
func (k Keeper) SynchronizeEarnReward(
	ctx sdk.Context,
	vaultDenom string,
	owner sdk.AccAddress,
	shares sdk.Dec,
) {
	claim, found := k.GetEarnClaim(ctx, owner)
	if !found {
		return
	}
	claim = k.synchronizeEarnReward(ctx, claim, vaultDenom, owner, shares)

	k.SetEarnClaim(ctx, claim)
}

// synchronizeEarnReward updates the reward and indexes in a earn claim for one vault.
func (k *Keeper) synchronizeEarnReward(
	ctx sdk.Context,
	claim types.EarnClaim,
	vaultDenom string,
	owner sdk.AccAddress,
	shares sdk.Dec,
) types.EarnClaim {
	globalRewardIndexes, found := k.GetEarnRewardIndexes(ctx, vaultDenom)
	if !found {
		// The global factor is only not found if
		// - the vault has not started accumulating rewards yet (either there is no reward specified in params, or the reward start time hasn't been hit)
		// - OR it was wrongly deleted from state (factors should never be removed while unsynced claims exist)
		// If not found we could either skip this sync, or assume the global factor is zero.
		// Skipping will avoid storing unnecessary factors in the claim for non rewarded vaults.
		// And in the event a global factor is wrongly deleted, it will avoid this function panicking when calculating rewards.
		return claim
	}

	userRewardIndexes, found := claim.RewardIndexes.Get(vaultDenom)
	if !found {
		// Normally the reward indexes should always be found.
		// But if a vault was not rewarded then becomes rewarded (ie a reward period is added to params), then the indexes will be missing from claims for that vault.
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
	claim.RewardIndexes = claim.RewardIndexes.With(vaultDenom, globalRewardIndexes)

	return claim
}

// GetSynchronizedEarnClaim fetches a earn claim from the store and syncs rewards for all rewarded vaults.
func (k Keeper) GetSynchronizedEarnClaim(ctx sdk.Context, owner sdk.AccAddress) (types.EarnClaim, bool) {
	claim, found := k.GetEarnClaim(ctx, owner)
	if !found {
		return types.EarnClaim{}, false
	}

	shares, found := k.earnKeeper.GetVaultAccountShares(ctx, owner)
	if !found {
		shares = earntypes.NewVaultShares()
	}

	k.IterateEarnRewardIndexes(ctx, func(vaultDenom string, _ types.RewardIndexes) bool {
		vaultAmount := shares.AmountOf(vaultDenom)
		claim = k.synchronizeEarnReward(ctx, claim, vaultDenom, owner, vaultAmount)

		return false
	})

	return claim, true
}
