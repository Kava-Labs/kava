package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	hardtypes "github.com/kava-labs/kava/x/hard/types"
	"github.com/kava-labs/kava/x/incentive/types"
)

// AccumulateHardBorrowRewards updates the rewards accumulated for the input reward period
func (k Keeper) AccumulateHardBorrowRewards(ctx sdk.Context, rewardPeriod types.MultiRewardPeriod) error {
	previousAccrualTime, found := k.GetPreviousHardBorrowRewardAccrualTime(ctx, rewardPeriod.CollateralType)
	if !found {
		k.SetPreviousHardBorrowRewardAccrualTime(ctx, rewardPeriod.CollateralType, ctx.BlockTime())
		return nil
	}
	timeElapsed := CalculateTimeElapsed(rewardPeriod.Start, rewardPeriod.End, ctx.BlockTime(), previousAccrualTime)
	if timeElapsed.IsZero() {
		return nil
	}
	if rewardPeriod.RewardsPerSecond.IsZero() {
		k.SetPreviousHardBorrowRewardAccrualTime(ctx, rewardPeriod.CollateralType, ctx.BlockTime())
		return nil
	}

	totalBorrowedCoins, foundTotalBorrowedCoins := k.hardKeeper.GetBorrowedCoins(ctx)
	if !foundTotalBorrowedCoins {
		k.SetPreviousHardBorrowRewardAccrualTime(ctx, rewardPeriod.CollateralType, ctx.BlockTime())
		return nil
	}

	totalBorrowed := totalBorrowedCoins.AmountOf(rewardPeriod.CollateralType).ToDec()
	if totalBorrowed.IsZero() {
		k.SetPreviousHardBorrowRewardAccrualTime(ctx, rewardPeriod.CollateralType, ctx.BlockTime())
		return nil
	}

	previousRewardIndexes, found := k.GetHardBorrowRewardIndexes(ctx, rewardPeriod.CollateralType)
	if !found {
		for _, rewardCoin := range rewardPeriod.RewardsPerSecond {
			rewardIndex := types.NewRewardIndex(rewardCoin.Denom, sdk.ZeroDec())
			previousRewardIndexes = append(previousRewardIndexes, rewardIndex)
		}
		k.SetHardBorrowRewardIndexes(ctx, rewardPeriod.CollateralType, previousRewardIndexes)
	}
	hardFactor, found := k.hardKeeper.GetBorrowInterestFactor(ctx, rewardPeriod.CollateralType)
	if !found {
		k.SetPreviousHardBorrowRewardAccrualTime(ctx, rewardPeriod.CollateralType, ctx.BlockTime())
		return nil
	}

	newRewardIndexes := previousRewardIndexes
	for _, rewardCoin := range rewardPeriod.RewardsPerSecond {
		newRewards := rewardCoin.Amount.ToDec().Mul(timeElapsed.ToDec())
		previousRewardIndex, found := previousRewardIndexes.GetRewardIndex(rewardCoin.Denom)
		if !found {
			previousRewardIndex = types.NewRewardIndex(rewardCoin.Denom, sdk.ZeroDec())
		}

		// Calculate new reward factor and update reward index
		rewardFactor := newRewards.Mul(hardFactor).Quo(totalBorrowed)
		newRewardFactorValue := previousRewardIndex.RewardFactor.Add(rewardFactor)
		newRewardIndex := types.NewRewardIndex(rewardCoin.Denom, newRewardFactorValue)
		i, found := newRewardIndexes.GetFactorIndex(rewardCoin.Denom)
		if found {
			newRewardIndexes[i] = newRewardIndex
		} else {
			newRewardIndexes = append(newRewardIndexes, newRewardIndex)
		}
	}
	k.SetHardBorrowRewardIndexes(ctx, rewardPeriod.CollateralType, newRewardIndexes)
	k.SetPreviousHardBorrowRewardAccrualTime(ctx, rewardPeriod.CollateralType, ctx.BlockTime())
	return nil
}

// InitializeHardBorrowReward initializes the borrow-side of a hard liquidity provider claim
// by creating the claim and setting the borrow reward factor index
func (k Keeper) InitializeHardBorrowReward(ctx sdk.Context, borrow hardtypes.Borrow) {
	claim, found := k.GetHardLiquidityProviderClaim(ctx, borrow.Borrower)
	if !found {
		claim = types.NewHardLiquidityProviderClaim(borrow.Borrower, sdk.Coins{}, nil, nil)
	}

	var borrowRewardIndexes types.MultiRewardIndexes
	for _, coin := range borrow.Amount {
		globalRewardIndexes, found := k.GetHardBorrowRewardIndexes(ctx, coin.Denom)
		if !found {
			globalRewardIndexes = types.RewardIndexes{}
		}
		borrowRewardIndexes = borrowRewardIndexes.With(coin.Denom, globalRewardIndexes)
	}

	claim.BorrowRewardIndexes = borrowRewardIndexes
	k.SetHardLiquidityProviderClaim(ctx, claim)
}

// SynchronizeHardBorrowReward updates the claim object by adding any accumulated rewards
// and updating the reward index value
func (k Keeper) SynchronizeHardBorrowReward(ctx sdk.Context, borrow hardtypes.Borrow) {
	claim, found := k.GetHardLiquidityProviderClaim(ctx, borrow.Borrower)
	if !found {
		return
	}

	for _, coin := range borrow.Amount {
		globalRewardIndexes, found := k.GetHardBorrowRewardIndexes(ctx, coin.Denom)
		if !found {
			// The global factor is only not found if
			// - the borrowed denom has not started accumulating rewards yet (either there is no reward specified in params, or the reward start time hasn't been hit)
			// - OR it was wrongly deleted from state (factors should never be removed while unsynced claims exist)
			// If not found we could either skip this sync, or assume the global factor is zero.
			// Skipping will avoid storing unnecessary factors in the claim for non rewarded denoms.
			// And in the event a global factor is wrongly deleted, it will avoid this function panicking when calculating rewards.
			continue
		}

		userRewardIndexes, found := claim.BorrowRewardIndexes.Get(coin.Denom)
		if !found {
			// Normally the reward indexes should always be found.
			// But if a denom was not rewarded then becomes rewarded (ie a reward period is added to params), then the indexes will be missing from claims for that borrowed denom.
			// So given the reward period was just added, assume the starting value for any global reward indexes, which is an empty slice.
			userRewardIndexes = types.RewardIndexes{}
		}

		newRewards, err := k.CalculateRewards(userRewardIndexes, globalRewardIndexes, coin.Amount.ToDec())
		if err != nil {
			// Global reward factors should never decrease, as it would lead to a negative update to claim.Rewards.
			// This panics if a global reward factor decreases or disappears between the old and new indexes.
			panic(fmt.Sprintf("corrupted global reward indexes found: %v", err))
		}

		claim.Reward = claim.Reward.Add(newRewards...)
		claim.BorrowRewardIndexes = claim.BorrowRewardIndexes.With(coin.Denom, globalRewardIndexes)
	}
	k.SetHardLiquidityProviderClaim(ctx, claim)
}

// UpdateHardBorrowIndexDenoms adds any new borrow denoms to the claim's borrow reward index
func (k Keeper) UpdateHardBorrowIndexDenoms(ctx sdk.Context, borrow hardtypes.Borrow) {
	claim, found := k.GetHardLiquidityProviderClaim(ctx, borrow.Borrower)
	if !found {
		claim = types.NewHardLiquidityProviderClaim(borrow.Borrower, sdk.Coins{}, nil, nil)
	}

	borrowDenoms := getDenoms(borrow.Amount)
	borrowRewardIndexDenoms := claim.BorrowRewardIndexes.GetCollateralTypes()

	borrowRewardIndexes := claim.BorrowRewardIndexes

	// Create a new multi-reward index in the claim for every new borrow denom
	uniqueBorrowDenoms := setDifference(borrowDenoms, borrowRewardIndexDenoms)

	for _, denom := range uniqueBorrowDenoms {
		globalBorrowRewardIndexes, found := k.GetHardBorrowRewardIndexes(ctx, denom)
		if !found {
			globalBorrowRewardIndexes = types.RewardIndexes{}
		}
		borrowRewardIndexes = borrowRewardIndexes.With(denom, globalBorrowRewardIndexes)
	}

	// Delete multi-reward index from claim if the collateral type is no longer borrowed
	uniqueBorrowRewardDenoms := setDifference(borrowRewardIndexDenoms, borrowDenoms)

	for _, denom := range uniqueBorrowRewardDenoms {
		borrowRewardIndexes = borrowRewardIndexes.RemoveRewardIndex(denom)
	}

	claim.BorrowRewardIndexes = borrowRewardIndexes
	k.SetHardLiquidityProviderClaim(ctx, claim)
}

// CalculateRewards computes how much rewards should have accrued to a source (eg a user's hard borrowed btc amount)
// between two index values.
//
// oldIndex is normally the index stored on a claim, newIndex the current global value, and rewardSource a hard borrowed/supplied amount.
//
// Returns an error if newIndexes does not contain all CollateralTypes from oldIndexes, or if any value of oldIndex.RewardFactor > newIndex.RewardFactor.
// This should never happen, as it would mean that a global reward index has decreased in value, or that a global reward index has been deleted from state.
func (k Keeper) CalculateRewards(oldIndexes, newIndexes types.RewardIndexes, rewardSource sdk.Dec) (sdk.Coins, error) {
	// check for missing CollateralType's
	for _, oldIndex := range oldIndexes {
		if newIndex, found := newIndexes.Get(oldIndex.CollateralType); !found {
			return nil, sdkerrors.Wrapf(types.ErrDecreasingRewardFactor, "old: %v, new: %v", oldIndex, newIndex)
		}
	}
	var reward sdk.Coins
	for _, newIndex := range newIndexes {
		oldFactor, found := oldIndexes.Get(newIndex.CollateralType)
		if !found {
			oldFactor = sdk.ZeroDec()
		}

		rewardAmount, err := k.CalculateSingleReward(oldFactor, newIndex.RewardFactor, rewardSource)
		if err != nil {
			return nil, err
		}

		reward = reward.Add(
			sdk.NewCoin(newIndex.CollateralType, rewardAmount),
		)
	}
	return reward, nil
}

// CalculateSingleReward computes how much rewards should have accrued to a source (eg a user's btcb-a cdp principal)
// between two index values.
//
// oldIndex is normally the index stored on a claim, newIndex the current global value, and rewardSource a cdp principal amount.
//
// Returns an error if oldIndex > newIndex. This should never happen, as it would mean that a global reward index has decreased in value,
// or that a global reward index has been deleted from state.
func (k Keeper) CalculateSingleReward(oldIndex, newIndex, rewardSource sdk.Dec) (sdk.Int, error) {
	increase := newIndex.Sub(oldIndex)
	if increase.IsNegative() {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrDecreasingRewardFactor, "old: %v, new: %v", oldIndex, newIndex)
	}
	reward := increase.Mul(rewardSource).RoundInt()
	return reward, nil
}
