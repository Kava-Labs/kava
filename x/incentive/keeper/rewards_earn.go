package keeper

import (
	"errors"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	earntypes "github.com/kava-labs/kava/x/earn/types"
	"github.com/kava-labs/kava/x/incentive/types"

	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	liquidtypes "github.com/kava-labs/kava/x/liquid/types"
)

// AccumulateEarnRewards calculates new rewards to distribute this block and updates the global indexes to reflect this.
// The provided rewardPeriod must be valid to avoid panics in calculating time durations.
func (k Keeper) AccumulateEarnRewards(ctx sdk.Context, rewardPeriod types.MultiRewardPeriod) {
	previousAccrualTime, found := k.GetEarnRewardAccrualTime(ctx)
	if !found {
		previousAccrualTime = ctx.BlockTime()
	}

	// TODO: Use DefaultDerivativeDenom or a liquid keeper method to get it?
	if rewardPeriod.CollateralType == liquidtypes.DefaultDerivativeDenom {
		k.accumulateEarnBkavaRewards(ctx, previousAccrualTime, rewardPeriod)
		return
	}

	k.accumulateEarnRewards(
		ctx,
		previousAccrualTime,
		rewardPeriod.CollateralType,
		rewardPeriod.Start,
		rewardPeriod.End,
		sdk.NewDecCoinsFromCoins(rewardPeriod.RewardsPerSecond...),
	)

	// Accrual time is always set to the current block time, even if the reward
	// period has ended.
	k.SetEarnRewardAccrualTime(ctx, ctx.BlockTime())
}

func GetProportionalRewardsPerSecond(
	rewardPeriod types.MultiRewardPeriod,
	totalBkavaSupply sdk.Int,
	singleBkavaSupply sdk.Int,
) sdk.DecCoins {
	// Rate per bkava-xxx = rewardsPerSecond * % of bkava-xxx
	//                    = rewardsPerSecond * (bkava-xxx / total bkava)
	//                    = (rewardsPerSecond * bkava-xxx) / total bkava

	newRate := sdk.NewDecCoins()

	// Prevent division by zero, if there are no total shares then there are no
	// rewards.
	if totalBkavaSupply.IsZero() {
		return newRate
	}

	for _, rewardCoin := range rewardPeriod.RewardsPerSecond {
		scaledAmount := rewardCoin.Amount.ToDec().
			Mul(singleBkavaSupply.ToDec()).
			Quo(totalBkavaSupply.ToDec())

		newRate = newRate.Add(sdk.NewDecCoinFromDec(rewardCoin.Denom, scaledAmount))
	}

	return newRate
}

// accumulateEarnBkavaRewards does the same as AccumulateEarnRewards but for
// *all* bkava vaults.
func (k Keeper) accumulateEarnBkavaRewards(
	ctx sdk.Context,
	previousAccrualTime time.Time,
	rewardPeriod types.MultiRewardPeriod,
) {
	// All bkava vault denoms
	var bkavaVaultsDenoms []string

	// bkava vault denoms from earn records (non-empty vaults)
	k.earnKeeper.IterateVaultRecords(ctx, func(record earntypes.VaultRecord) (stop bool) {
		if k.liquidKeeper.IsDerivativeDenom(ctx, record.TotalShares.Denom) {
			bkavaVaultsDenoms = append(bkavaVaultsDenoms, record.TotalShares.Denom)
		}

		return false
	})

	// Vaults that have accumulated rewards in indexes, but are empty in earn
	// will not be included in accumulateBkavaEarnRewards. This is fine, as
	// these indexes won't change.
	totalBkavaSupply := k.liquidKeeper.GetTotalDerivativeSupply(ctx)

	// Accumulate rewards for each non-empty bkava vault.
	for _, bkavaDenom := range bkavaVaultsDenoms {
		k.accumulateBkavaEarnRewards(
			ctx,
			previousAccrualTime,
			bkavaDenom,
			rewardPeriod.Start,
			rewardPeriod.End,
			GetProportionalRewardsPerSecond(
				rewardPeriod,
				totalBkavaSupply,
				k.liquidKeeper.GetDerivativeSupply(ctx, bkavaDenom),
			),
		)
	}
}

func (k Keeper) accumulateBkavaEarnRewards(
	ctx sdk.Context,
	previousAccrualTime time.Time,
	collateralType string,
	periodStart time.Time,
	periodEnd time.Time,
	periodRewardsPerSecond sdk.DecCoins,
) {
	// Collect staking rewards for this validator
	stakingRewards := k.collectDerivativeStakingRewards(ctx, collateralType)

	// Collect incentive rewards
	perSecondRewards, _ := types.CalculatePerSecondRewards(
		periodStart,
		periodEnd,
		periodRewardsPerSecond,
		previousAccrualTime,
		ctx.BlockTime(),
	)
	rewards := stakingRewards.Add(perSecondRewards...)

	// Distribute rewards by incrementing indexes
	indexes, found := k.GetEarnRewardIndexes(ctx, collateralType)
	if !found {
		indexes = types.RewardIndexes{}
	}

	totalSourceShares := k.getEarnTotalSourceShares(ctx, collateralType)
	var increment types.RewardIndexes
	if totalSourceShares.GT(sdk.ZeroDec()) {
		// leave as nil if no source shares
		increment = types.NewRewardIndexesFromCoins(rewards).Quo(totalSourceShares)
	}
	updatedIndexes := indexes.Add(increment)

	if len(updatedIndexes) > 0 {
		// the store panics when setting empty or nil indexes
		k.SetEarnRewardIndexes(ctx, collateralType, updatedIndexes)
	}
}

func (k Keeper) collectDerivativeStakingRewards(ctx sdk.Context, collateralType string) sdk.DecCoins {
	// the store panics when setting empty or nil indexes
	rewards, err := k.liquidKeeper.CollectStakingRewardsByDenom(ctx, collateralType, types.IncentiveMacc)
	if err != nil {
		if !errors.Is(err, distrtypes.ErrNoValidatorDistInfo) &&
			!errors.Is(err, distrtypes.ErrEmptyDelegationDistInfo) {
			panic(fmt.Sprintf("failed to collect staking rewards for %s: %s", collateralType, err))
		}
		// otherwise there's no validator or delegation yet
		rewards = nil
	}
	return sdk.NewDecCoinsFromCoins(rewards...)
}

func (k Keeper) accumulateEarnRewards(
	ctx sdk.Context,
	previousAccrualTime time.Time,
	collateralType string,
	periodStart time.Time,
	periodEnd time.Time,
	periodRewardsPerSecond sdk.DecCoins,
) {
	indexes, found := k.GetEarnRewardIndexes(ctx, collateralType)
	if !found {
		indexes = types.RewardIndexes{}
	}

	acc := types.NewAccumulator(previousAccrualTime, indexes)

	totalSourceShares := k.getEarnTotalSourceShares(ctx, collateralType)

	acc.AccumulateDecCoins(
		periodStart,
		periodEnd,
		periodRewardsPerSecond,
		totalSourceShares,
		ctx.BlockTime(),
	)

	if len(acc.Indexes) > 0 {
		// the store panics when setting empty or nil indexes
		k.SetEarnRewardIndexes(ctx, collateralType, acc.Indexes)
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
