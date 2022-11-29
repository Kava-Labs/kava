package accumulators

import (
	"errors"
	"fmt"
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	earntypes "github.com/kava-labs/kava/x/earn/types"
	"github.com/kava-labs/kava/x/incentive/keeper/adapters"
	"github.com/kava-labs/kava/x/incentive/keeper/store"
	"github.com/kava-labs/kava/x/incentive/types"
)

// EarnAccumulator is an accumulator for Earn claim types. This includes
// claiming staking rewards and reward distribution for liquid kava.
type EarnAccumulator struct {
	store        store.IncentiveStore
	liquidKeeper types.LiquidKeeper
	earnKeeper   types.EarnKeeper
	adapters     adapters.SourceAdapters
}

var _ types.RewardAccumulator = EarnAccumulator{}

// NewEarnAccumulator returns a new EarnAccumulator.
func NewEarnAccumulator(
	store store.IncentiveStore,
	liquidKeeper types.LiquidKeeper,
	earnKeeper types.EarnKeeper,
	adapters adapters.SourceAdapters,
) EarnAccumulator {
	return EarnAccumulator{
		store:        store,
		liquidKeeper: liquidKeeper,
		earnKeeper:   earnKeeper,
		adapters:     adapters,
	}
}

// AccumulateRewards calculates new rewards to distribute this block and updates
// the global indexes to reflect this. The provided rewardPeriod must be valid
// to avoid panics in calculating time durations.
func (a EarnAccumulator) AccumulateRewards(
	ctx sdk.Context,
	claimType types.ClaimType,
	rewardPeriod types.MultiRewardPeriod,
) error {
	if claimType != types.CLAIM_TYPE_EARN {
		panic(fmt.Sprintf(
			"invalid claim type for earn accumulator, expected %s but got %s",
			types.CLAIM_TYPE_EARN,
			claimType,
		))
	}

	if rewardPeriod.CollateralType == "bkava" {
		return a.accumulateEarnBkavaRewards(ctx, rewardPeriod)
	}

	// Non bkava vaults use the basic accumulator.
	return NewBasicAccumulator(a.store, a.adapters).AccumulateRewards(ctx, claimType, rewardPeriod)
}

// accumulateEarnBkavaRewards does the same as AccumulateEarnRewards but for
// *all* bkava vaults.
func (k EarnAccumulator) accumulateEarnBkavaRewards(ctx sdk.Context, rewardPeriod types.MultiRewardPeriod) error {
	// All bkava vault denoms
	bkavaVaultsDenoms := make(map[string]bool)

	// bkava vault denoms from earn records (non-empty vaults)
	k.earnKeeper.IterateVaultRecords(ctx, func(record earntypes.VaultRecord) (stop bool) {
		if k.liquidKeeper.IsDerivativeDenom(ctx, record.TotalShares.Denom) {
			bkavaVaultsDenoms[record.TotalShares.Denom] = true
		}

		return false
	})

	// bkava vault denoms from past incentive indexes, may include vaults
	// that were fully withdrawn.
	k.store.IterateRewardIndexesByClaimType(
		ctx,
		types.CLAIM_TYPE_EARN,
		func(reward types.TypedRewardIndexes) (stop bool) {
			if k.liquidKeeper.IsDerivativeDenom(ctx, reward.CollateralType) {
				bkavaVaultsDenoms[reward.CollateralType] = true
			}

			return false
		})

	totalBkavaValue, err := k.liquidKeeper.GetTotalDerivativeValue(ctx)
	if err != nil {
		return err
	}

	i := 0
	sortedBkavaVaultsDenoms := make([]string, len(bkavaVaultsDenoms))
	for vaultDenom := range bkavaVaultsDenoms {
		sortedBkavaVaultsDenoms[i] = vaultDenom
		i++
	}

	// Sort the vault denoms to ensure deterministic iteration order.
	sort.Strings(sortedBkavaVaultsDenoms)

	// Accumulate rewards for each bkava vault.
	for _, bkavaDenom := range sortedBkavaVaultsDenoms {
		derivativeValue, err := k.liquidKeeper.GetDerivativeValue(ctx, bkavaDenom)
		if err != nil {
			return err
		}

		k.accumulateBkavaEarnRewards(
			ctx,
			bkavaDenom,
			rewardPeriod.Start,
			rewardPeriod.End,
			GetProportionalRewardsPerSecond(
				rewardPeriod,
				totalBkavaValue.Amount,
				derivativeValue.Amount,
			),
		)
	}

	return nil
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

func (k EarnAccumulator) accumulateBkavaEarnRewards(
	ctx sdk.Context,
	collateralType string,
	periodStart time.Time,
	periodEnd time.Time,
	periodRewardsPerSecond sdk.DecCoins,
) {
	// Collect staking rewards for this validator, does not have any start/end
	// period time restrictions.
	stakingRewards := k.collectDerivativeStakingRewards(ctx, collateralType)

	// Collect incentive rewards
	// **Total rewards** for vault per second, NOT per share
	perSecondRewards := k.collectPerSecondRewards(
		ctx,
		collateralType,
		periodStart,
		periodEnd,
		periodRewardsPerSecond,
	)

	// **Total rewards** for vault per second, NOT per share
	rewards := stakingRewards.Add(perSecondRewards...)

	// Distribute rewards by incrementing indexes
	indexes, found := k.store.GetRewardIndexesOfClaimType(ctx, types.CLAIM_TYPE_EARN, collateralType)
	if !found {
		indexes = types.RewardIndexes{}
	}

	totalSourceShares := k.adapters.TotalSharesBySource(ctx, types.CLAIM_TYPE_EARN, collateralType)
	var increment types.RewardIndexes
	if totalSourceShares.GT(sdk.ZeroDec()) {
		// Divide total rewards by total shares to get the reward **per share**
		// Leave as nil if no source shares
		increment = types.NewRewardIndexesFromCoins(rewards).Quo(totalSourceShares)
	}
	updatedIndexes := indexes.Add(increment)

	if len(updatedIndexes) > 0 {
		// the store panics when setting empty or nil indexes
		k.store.SetRewardIndexes(ctx, types.CLAIM_TYPE_EARN, collateralType, updatedIndexes)
	}
}

func (k EarnAccumulator) collectDerivativeStakingRewards(ctx sdk.Context, collateralType string) sdk.DecCoins {
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

func (k EarnAccumulator) collectPerSecondRewards(
	ctx sdk.Context,
	collateralType string,
	periodStart time.Time,
	periodEnd time.Time,
	periodRewardsPerSecond sdk.DecCoins,
) sdk.DecCoins {
	previousAccrualTime, found := k.store.GetRewardAccrualTime(ctx, types.CLAIM_TYPE_EARN, collateralType)
	if !found {
		previousAccrualTime = ctx.BlockTime()
	}

	rewards, accumulatedTo := types.CalculatePerSecondRewards(
		periodStart,
		periodEnd,
		periodRewardsPerSecond,
		previousAccrualTime,
		ctx.BlockTime(),
	)

	k.store.SetRewardAccrualTime(ctx, types.CLAIM_TYPE_EARN, collateralType, accumulatedTo)

	// Don't need to move funds as they're assumed to be in the IncentiveMacc module account already.
	return rewards
}
