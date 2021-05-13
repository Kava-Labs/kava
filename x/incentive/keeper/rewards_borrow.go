package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

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
		claim = types.NewHardLiquidityProviderClaim(borrow.Borrower, sdk.Coins{}, nil, nil, nil)
	}

	var borrowRewardIndexes types.MultiRewardIndexes
	for _, coin := range borrow.Amount {
		globalRewardIndexes, foundGlobalRewardIndexes := k.GetHardBorrowRewardIndexes(ctx, coin.Denom)
		if !foundGlobalRewardIndexes {
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
		globalRewardIndexes, foundGlobalRewardIndexes := k.GetHardBorrowRewardIndexes(ctx, coin.Denom)
		if !foundGlobalRewardIndexes {
			continue
		}

		userRewardIndexes, found := claim.BorrowRewardIndexes.Get(coin.Denom)
		if !found {
			continue
		}

		for _, globalRewardIndex := range globalRewardIndexes {
			factor, found := userRewardIndexes.Get(globalRewardIndex.CollateralType)
			if !found {
				factor = sdk.ZeroDec()
			}

			claim.Reward = claim.Reward.Add(
				sdk.NewCoin(
					globalRewardIndex.CollateralType,
					k.calculateRewardAmount(
						factor,
						globalRewardIndex.RewardFactor,
						coin.Amount,
					),
				),
			)
			userRewardIndexes = userRewardIndexes.With(globalRewardIndex.CollateralType, globalRewardIndex.RewardFactor)
		}
		claim.BorrowRewardIndexes = claim.BorrowRewardIndexes.With(coin.Denom, userRewardIndexes)
	}
	k.SetHardLiquidityProviderClaim(ctx, claim)
}

// UpdateHardBorrowIndexDenoms adds any new borrow denoms to the claim's borrow reward index
func (k Keeper) UpdateHardBorrowIndexDenoms(ctx sdk.Context, borrow hardtypes.Borrow) {
	claim, found := k.GetHardLiquidityProviderClaim(ctx, borrow.Borrower)
	if !found {
		claim = types.NewHardLiquidityProviderClaim(borrow.Borrower, sdk.Coins{}, nil, nil, nil)
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

// calculateRewardAmount computes the rewards that should accumulate between two index values.

// oldIndex is normally the index stored on a claim, newIndex is the current global value, and rewardSource is hard borrow/supply amount.
// newIndex MUST be greater than oldIndex otherwise it will panic
func (k Keeper) calculateRewardAmount(oldIndex, newIndex sdk.Dec, rewardSource sdk.Int) sdk.Int {
	increase := newIndex.Sub(oldIndex)
	if increase.IsNegative() {
		panic(fmt.Sprintf("new reward index cannot be less than previous: new %s, old %s", newIndex, oldIndex))
	}

	return increase.Mul(rewardSource.ToDec()).RoundInt()
}

// calculateReward computes the reward that should accumulate between two index values.
//
// oldIndex is normally the index stored on a claim, newIndex is the current global value, and rewardSource is hard borrow/supply amount.
// newIndex RewardFactor MUST be greater than oldIndex RewardFactor otherwise it will panic.
// Index CollateralTypes MUST match or it will panic.
func (k Keeper) calculateReward(oldIndex, newIndex types.RewardIndex, rewardSource sdk.Int) sdk.Coin {
	if oldIndex.CollateralType != newIndex.CollateralType {
		panic(fmt.Sprintf(
			"cannot calculate reward for reward indexes with different denoms: old %s new %s",
			oldIndex.CollateralType,
			newIndex.CollateralType,
		)) // TODO should this error instead?
	}
	return sdk.NewCoin(
		oldIndex.CollateralType,
		k.calculateRewardAmount(oldIndex.RewardFactor, newIndex.RewardFactor, rewardSource),
	)
}
