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
		var multiRewardIndex types.MultiRewardIndex
		if foundGlobalRewardIndexes {
			multiRewardIndex = types.NewMultiRewardIndex(coin.Denom, globalRewardIndexes)
		} else {
			multiRewardIndex = types.NewMultiRewardIndex(coin.Denom, types.RewardIndexes{})
		}
		borrowRewardIndexes = append(borrowRewardIndexes, multiRewardIndex)
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

		userMultiRewardIndex, foundUserMultiRewardIndex := claim.BorrowRewardIndexes.GetRewardIndex(coin.Denom)
		if !foundUserMultiRewardIndex {
			continue
		}

		userRewardIndexIndex, foundUserRewardIndexIndex := claim.BorrowRewardIndexes.GetRewardIndexIndex(coin.Denom)
		if !foundUserRewardIndexIndex {
			continue
		}

		for _, globalRewardIndex := range globalRewardIndexes {
			userRewardIndex, foundUserRewardIndex := userMultiRewardIndex.RewardIndexes.GetRewardIndex(globalRewardIndex.CollateralType)
			if !foundUserRewardIndex {
				// User borrowed this coin type before it had rewards. When new rewards are added, legacy borrowers
				// should immediately begin earning rewards. Enable users to do so by updating their claim with the global
				// reward index denom and start their reward factor at 0.0
				userRewardIndex = types.NewRewardIndex(globalRewardIndex.CollateralType, sdk.ZeroDec())
				userMultiRewardIndex.RewardIndexes = append(userMultiRewardIndex.RewardIndexes, userRewardIndex)
				claim.BorrowRewardIndexes[userRewardIndexIndex] = userMultiRewardIndex
			}

			newRewardsAmount := k.calculateReward(
				userRewardIndex.RewardFactor,
				globalRewardIndex.RewardFactor,
				borrow.Amount.AmountOf(coin.Denom),
			)

			factorIndex, foundFactorIndex := userMultiRewardIndex.RewardIndexes.GetFactorIndex(globalRewardIndex.CollateralType)
			if !foundFactorIndex { // should never trigger
				continue
			}
			claim.BorrowRewardIndexes[userRewardIndexIndex].RewardIndexes[factorIndex].RewardFactor = globalRewardIndex.RewardFactor
			newRewardsCoin := sdk.NewCoin(userRewardIndex.CollateralType, newRewardsAmount)
			claim.Reward = claim.Reward.Add(newRewardsCoin)
		}
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

	uniqueBorrowDenoms := setDifference(borrowDenoms, borrowRewardIndexDenoms)
	uniqueBorrowRewardDenoms := setDifference(borrowRewardIndexDenoms, borrowDenoms)

	borrowRewardIndexes := claim.BorrowRewardIndexes
	// Create a new multi-reward index in the claim for every new borrow denom
	for _, denom := range uniqueBorrowDenoms {
		_, foundUserRewardIndexes := claim.BorrowRewardIndexes.GetRewardIndex(denom)
		if !foundUserRewardIndexes {
			globalBorrowRewardIndexes, foundGlobalBorrowRewardIndexes := k.GetHardBorrowRewardIndexes(ctx, denom)
			var multiRewardIndex types.MultiRewardIndex
			if foundGlobalBorrowRewardIndexes {
				multiRewardIndex = types.NewMultiRewardIndex(denom, globalBorrowRewardIndexes)
			} else {
				multiRewardIndex = types.NewMultiRewardIndex(denom, types.RewardIndexes{})
			}
			borrowRewardIndexes = append(borrowRewardIndexes, multiRewardIndex)
		}
	}

	// Delete multi-reward index from claim if the collateral type is no longer borrowed
	for _, denom := range uniqueBorrowRewardDenoms {
		borrowRewardIndexes = borrowRewardIndexes.RemoveRewardIndex(denom)
	}

	claim.BorrowRewardIndexes = borrowRewardIndexes
	k.SetHardLiquidityProviderClaim(ctx, claim)
}

func (k Keeper) calculateReward(oldIndex, newIndex sdk.Dec, rewardSource sdk.Int) sdk.Int {
	increase := newIndex.Sub(oldIndex)
	if increase.IsNegative() {
		panic(fmt.Sprintf("new reward index cannot be less than previous: new %s, old %s", newIndex, oldIndex))
	}

	return increase.Mul(rewardSource.ToDec()).RoundInt()
}
