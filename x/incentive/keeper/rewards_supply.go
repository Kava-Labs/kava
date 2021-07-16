package keeper

import (
	"fmt"
	"math"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	hardtypes "github.com/kava-labs/kava/x/hard/types"
	"github.com/kava-labs/kava/x/incentive/types"
)

// AccumulateHardSupplyRewards calculates new rewards to distribute this block and updates the global indexes to reflect this.
// The provided rewardPeriod must be valid to avoid panics in calculating time durations.
func (k Keeper) AccumulateHardSupplyRewards(ctx sdk.Context, rewardPeriod types.MultiRewardPeriod) {

	previousAccrualTime, found := k.GetPreviousHardSupplyRewardAccrualTime(ctx, rewardPeriod.CollateralType)
	if !found {
		previousAccrualTime = ctx.BlockTime()
	}

	indexes, found := k.GetHardSupplyRewardIndexes(ctx, rewardPeriod.CollateralType)
	if !found {
		indexes = types.RewardIndexes{}
	}

	acc := types.NewAccumulator(previousAccrualTime, indexes)

	totalSource := k.getHardSupplyTotalSourceShares(ctx, rewardPeriod.CollateralType)

	acc.Accumulate(rewardPeriod, totalSource, ctx.BlockTime())

	k.SetPreviousHardSupplyRewardAccrualTime(ctx, rewardPeriod.CollateralType, acc.PreviousAccumulationTime)
	if len(acc.Indexes) > 0 {
		// the store panics when setting empty or nil indexes
		k.SetHardSupplyRewardIndexes(ctx, rewardPeriod.CollateralType, acc.Indexes)
	}
}

// getHardSupplyTotalSourceShares fetches the sum of all source shares for a supply reward.
// In the case of hard supply, this is the total supplied divided by the supply interest factor.
// This give the "pre interest" value of the total supplied.
func (k Keeper) getHardSupplyTotalSourceShares(ctx sdk.Context, denom string) sdk.Dec {
	totalSuppliedCoins, found := k.hardKeeper.GetSuppliedCoins(ctx)
	if !found {
		// assume no coins have been supplied
		totalSuppliedCoins = sdk.NewCoins()
	}
	totalSupplied := totalSuppliedCoins.AmountOf(denom)

	interestFactor, found := k.hardKeeper.GetSupplyInterestFactor(ctx, denom)
	if !found {
		// assume nothing has been borrowed so the factor starts at it's default value
		interestFactor = sdk.OneDec()
	}

	// return supplied/factor to get the "pre interest" value of the current total supplied
	return totalSupplied.ToDec().Quo(interestFactor)
}

// InitializeHardSupplyReward initializes the supply-side of a hard liquidity provider claim
// by creating the claim and setting the supply reward factor index
func (k Keeper) InitializeHardSupplyReward(ctx sdk.Context, deposit hardtypes.Deposit) {
	claim, found := k.GetHardLiquidityProviderClaim(ctx, deposit.Depositor)
	if !found {
		claim = types.NewHardLiquidityProviderClaim(deposit.Depositor, sdk.Coins{}, nil, nil)
	}

	var supplyRewardIndexes types.MultiRewardIndexes
	for _, coin := range deposit.Amount {
		globalRewardIndexes, found := k.GetHardSupplyRewardIndexes(ctx, coin.Denom)
		if !found {
			globalRewardIndexes = types.RewardIndexes{}
		}
		supplyRewardIndexes = supplyRewardIndexes.With(coin.Denom, globalRewardIndexes)
	}

	claim.SupplyRewardIndexes = supplyRewardIndexes
	k.SetHardLiquidityProviderClaim(ctx, claim)
}

// SynchronizeHardSupplyReward updates the claim object by adding any accumulated rewards
// and updating the reward index value
func (k Keeper) SynchronizeHardSupplyReward(ctx sdk.Context, deposit hardtypes.Deposit) {
	claim, found := k.GetHardLiquidityProviderClaim(ctx, deposit.Depositor)
	if !found {
		return
	}

	for _, coin := range deposit.Amount {
		globalRewardIndexes, found := k.GetHardSupplyRewardIndexes(ctx, coin.Denom)
		if !found {
			// The global factor is only not found if
			// - the supply denom has not started accumulating rewards yet (either there is no reward specified in params, or the reward start time hasn't been hit)
			// - OR it was wrongly deleted from state (factors should never be removed while unsynced claims exist)
			// If not found we could either skip this sync, or assume the global factor is zero.
			// Skipping will avoid storing unnecessary factors in the claim for non rewarded denoms.
			// And in the event a global factor is wrongly deleted, it will avoid this function panicking when calculating rewards.
			continue
		}

		userRewardIndexes, found := claim.SupplyRewardIndexes.Get(coin.Denom)
		if !found {
			// Normally the reward indexes should always be found.
			// But if a denom was not rewarded then becomes rewarded (ie a reward period is added to params), then the indexes will be missing from claims for that supplied denom.
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
		claim.SupplyRewardIndexes = claim.SupplyRewardIndexes.With(coin.Denom, globalRewardIndexes)
	}
	k.SetHardLiquidityProviderClaim(ctx, claim)
}

// UpdateHardSupplyIndexDenoms adds any new deposit denoms to the claim's supply reward index
func (k Keeper) UpdateHardSupplyIndexDenoms(ctx sdk.Context, deposit hardtypes.Deposit) {
	claim, found := k.GetHardLiquidityProviderClaim(ctx, deposit.Depositor)
	if !found {
		claim = types.NewHardLiquidityProviderClaim(deposit.Depositor, sdk.Coins{}, nil, nil)
	}

	depositDenoms := getDenoms(deposit.Amount)
	supplyRewardIndexDenoms := claim.SupplyRewardIndexes.GetCollateralTypes()

	supplyRewardIndexes := claim.SupplyRewardIndexes

	// Create a new multi-reward index in the claim for every new deposit denom
	uniqueDepositDenoms := setDifference(depositDenoms, supplyRewardIndexDenoms)

	for _, denom := range uniqueDepositDenoms {
		globalSupplyRewardIndexes, found := k.GetHardSupplyRewardIndexes(ctx, denom)
		if !found {
			globalSupplyRewardIndexes = types.RewardIndexes{}
		}
		supplyRewardIndexes = supplyRewardIndexes.With(denom, globalSupplyRewardIndexes)
	}

	// Delete multi-reward index from claim if the collateral type is no longer deposited
	uniqueSupplyRewardDenoms := setDifference(supplyRewardIndexDenoms, depositDenoms)

	for _, denom := range uniqueSupplyRewardDenoms {
		supplyRewardIndexes = supplyRewardIndexes.RemoveRewardIndex(denom)
	}

	claim.SupplyRewardIndexes = supplyRewardIndexes
	k.SetHardLiquidityProviderClaim(ctx, claim)
}

// SynchronizeHardLiquidityProviderClaim adds any accumulated rewards
func (k Keeper) SynchronizeHardLiquidityProviderClaim(ctx sdk.Context, owner sdk.AccAddress) {
	// Synchronize any hard liquidity supply-side rewards
	deposit, foundDeposit := k.hardKeeper.GetDeposit(ctx, owner)
	if foundDeposit {
		k.SynchronizeHardSupplyReward(ctx, deposit)
	}

	// Synchronize any hard liquidity borrow-side rewards
	borrow, foundBorrow := k.hardKeeper.GetBorrow(ctx, owner)
	if foundBorrow {
		k.SynchronizeHardBorrowReward(ctx, borrow)
	}
}

// ZeroHardLiquidityProviderClaim zeroes out the claim object's rewards and returns the updated claim object
func (k Keeper) ZeroHardLiquidityProviderClaim(ctx sdk.Context, claim types.HardLiquidityProviderClaim) types.HardLiquidityProviderClaim {
	claim.Reward = sdk.NewCoins()
	k.SetHardLiquidityProviderClaim(ctx, claim)
	return claim
}

// SimulateHardSynchronization calculates a user's outstanding hard rewards by simulating reward synchronization
func (k Keeper) SimulateHardSynchronization(ctx sdk.Context, claim types.HardLiquidityProviderClaim) types.HardLiquidityProviderClaim {
	// 1. Simulate Hard supply-side rewards
	for _, ri := range claim.SupplyRewardIndexes {
		globalRewardIndexes, foundGlobalRewardIndexes := k.GetHardSupplyRewardIndexes(ctx, ri.CollateralType)
		if !foundGlobalRewardIndexes {
			continue
		}

		userRewardIndexes, foundUserRewardIndexes := claim.SupplyRewardIndexes.GetRewardIndex(ri.CollateralType)
		if !foundUserRewardIndexes {
			continue
		}

		userRewardIndexIndex, foundUserRewardIndexIndex := claim.SupplyRewardIndexes.GetRewardIndexIndex(ri.CollateralType)
		if !foundUserRewardIndexIndex {
			continue
		}

		for _, globalRewardIndex := range globalRewardIndexes {
			userRewardIndex, foundUserRewardIndex := userRewardIndexes.RewardIndexes.GetRewardIndex(globalRewardIndex.CollateralType)
			if !foundUserRewardIndex {
				userRewardIndex = types.NewRewardIndex(globalRewardIndex.CollateralType, sdk.ZeroDec())
				userRewardIndexes.RewardIndexes = append(userRewardIndexes.RewardIndexes, userRewardIndex)
				claim.SupplyRewardIndexes[userRewardIndexIndex].RewardIndexes = append(claim.SupplyRewardIndexes[userRewardIndexIndex].RewardIndexes, userRewardIndex)
			}

			globalRewardFactor := globalRewardIndex.RewardFactor
			userRewardFactor := userRewardIndex.RewardFactor
			rewardsAccumulatedFactor := globalRewardFactor.Sub(userRewardFactor)
			if rewardsAccumulatedFactor.IsZero() {
				continue
			}
			deposit, found := k.hardKeeper.GetDeposit(ctx, claim.GetOwner())
			if !found {
				continue
			}
			newRewardsAmount := rewardsAccumulatedFactor.Mul(deposit.Amount.AmountOf(ri.CollateralType).ToDec()).RoundInt()
			if newRewardsAmount.IsZero() || newRewardsAmount.IsNegative() {
				continue
			}

			factorIndex, foundFactorIndex := userRewardIndexes.RewardIndexes.GetFactorIndex(globalRewardIndex.CollateralType)
			if !foundFactorIndex {
				continue
			}
			claim.SupplyRewardIndexes[userRewardIndexIndex].RewardIndexes[factorIndex].RewardFactor = globalRewardIndex.RewardFactor
			newRewardsCoin := sdk.NewCoin(userRewardIndex.CollateralType, newRewardsAmount)
			claim.Reward = claim.Reward.Add(newRewardsCoin)
		}
	}

	// 2. Simulate Hard borrow-side rewards
	for _, ri := range claim.BorrowRewardIndexes {
		globalRewardIndexes, foundGlobalRewardIndexes := k.GetHardBorrowRewardIndexes(ctx, ri.CollateralType)
		if !foundGlobalRewardIndexes {
			continue
		}

		userRewardIndexes, foundUserRewardIndexes := claim.BorrowRewardIndexes.GetRewardIndex(ri.CollateralType)
		if !foundUserRewardIndexes {
			continue
		}

		userRewardIndexIndex, foundUserRewardIndexIndex := claim.BorrowRewardIndexes.GetRewardIndexIndex(ri.CollateralType)
		if !foundUserRewardIndexIndex {
			continue
		}

		for _, globalRewardIndex := range globalRewardIndexes {
			userRewardIndex, foundUserRewardIndex := userRewardIndexes.RewardIndexes.GetRewardIndex(globalRewardIndex.CollateralType)
			if !foundUserRewardIndex {
				userRewardIndex = types.NewRewardIndex(globalRewardIndex.CollateralType, sdk.ZeroDec())
				userRewardIndexes.RewardIndexes = append(userRewardIndexes.RewardIndexes, userRewardIndex)
				claim.BorrowRewardIndexes[userRewardIndexIndex].RewardIndexes = append(claim.BorrowRewardIndexes[userRewardIndexIndex].RewardIndexes, userRewardIndex)
			}

			globalRewardFactor := globalRewardIndex.RewardFactor
			userRewardFactor := userRewardIndex.RewardFactor
			rewardsAccumulatedFactor := globalRewardFactor.Sub(userRewardFactor)
			if rewardsAccumulatedFactor.IsZero() {
				continue
			}
			borrow, found := k.hardKeeper.GetBorrow(ctx, claim.GetOwner())
			if !found {
				continue
			}
			newRewardsAmount := rewardsAccumulatedFactor.Mul(borrow.Amount.AmountOf(ri.CollateralType).ToDec()).RoundInt()
			if newRewardsAmount.IsZero() || newRewardsAmount.IsNegative() {
				continue
			}

			factorIndex, foundFactorIndex := userRewardIndexes.RewardIndexes.GetFactorIndex(globalRewardIndex.CollateralType)
			if !foundFactorIndex {
				continue
			}
			claim.BorrowRewardIndexes[userRewardIndexIndex].RewardIndexes[factorIndex].RewardFactor = globalRewardIndex.RewardFactor
			newRewardsCoin := sdk.NewCoin(userRewardIndex.CollateralType, newRewardsAmount)
			claim.Reward = claim.Reward.Add(newRewardsCoin)
		}
	}

	return claim
}

// CalculateTimeElapsed calculates the number of reward-eligible seconds that have passed since the previous
// time rewards were accrued, taking into account the end time of the reward period
func CalculateTimeElapsed(start, end, blockTime time.Time, previousAccrualTime time.Time) sdk.Int {
	if (end.Before(blockTime) &&
		(end.Before(previousAccrualTime) || end.Equal(previousAccrualTime))) ||
		(start.After(blockTime)) ||
		(start.Equal(blockTime)) {
		return sdk.ZeroInt()
	}
	if start.After(previousAccrualTime) && start.Before(blockTime) {
		previousAccrualTime = start
	}

	if end.Before(blockTime) {
		return sdk.MaxInt(sdk.ZeroInt(), sdk.NewInt(int64(math.RoundToEven(
			end.Sub(previousAccrualTime).Seconds(),
		))))
	}
	return sdk.MaxInt(sdk.ZeroInt(), sdk.NewInt(int64(math.RoundToEven(
		blockTime.Sub(previousAccrualTime).Seconds(),
	))))
}

// Set setDifference: A - B
func setDifference(a, b []string) (diff []string) {
	m := make(map[string]bool)

	for _, item := range b {
		m[item] = true
	}

	for _, item := range a {
		if _, ok := m[item]; !ok {
			diff = append(diff, item)
		}
	}
	return
}

func getDenoms(coins sdk.Coins) []string {
	denoms := []string{}
	for _, coin := range coins {
		denoms = append(denoms, coin.Denom)
	}
	return denoms
}
