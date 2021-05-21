package keeper

import (
	"fmt"
	"math"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	hardtypes "github.com/kava-labs/kava/x/hard/types"
	"github.com/kava-labs/kava/x/incentive/types"
)

// AccumulateHardSupplyRewards updates the rewards accumulated for the input reward period
func (k Keeper) AccumulateHardSupplyRewards(ctx sdk.Context, rewardPeriod types.MultiRewardPeriod) error {
	previousAccrualTime, found := k.GetPreviousHardSupplyRewardAccrualTime(ctx, rewardPeriod.CollateralType)
	if !found {
		k.SetPreviousHardSupplyRewardAccrualTime(ctx, rewardPeriod.CollateralType, ctx.BlockTime())
		return nil
	}
	timeElapsed := CalculateTimeElapsed(rewardPeriod.Start, rewardPeriod.End, ctx.BlockTime(), previousAccrualTime)
	if timeElapsed.IsZero() {
		return nil
	}
	if rewardPeriod.RewardsPerSecond.IsZero() {
		k.SetPreviousHardSupplyRewardAccrualTime(ctx, rewardPeriod.CollateralType, ctx.BlockTime())
		return nil
	}

	totalSuppliedCoins, foundTotalSuppliedCoins := k.hardKeeper.GetSuppliedCoins(ctx)
	if !foundTotalSuppliedCoins {
		k.SetPreviousHardSupplyRewardAccrualTime(ctx, rewardPeriod.CollateralType, ctx.BlockTime())
		return nil
	}

	totalSupplied := totalSuppliedCoins.AmountOf(rewardPeriod.CollateralType).ToDec()
	if totalSupplied.IsZero() {
		k.SetPreviousHardSupplyRewardAccrualTime(ctx, rewardPeriod.CollateralType, ctx.BlockTime())
		return nil
	}

	previousRewardIndexes, found := k.GetHardSupplyRewardIndexes(ctx, rewardPeriod.CollateralType)
	if !found {
		for _, rewardCoin := range rewardPeriod.RewardsPerSecond {
			rewardIndex := types.NewRewardIndex(rewardCoin.Denom, sdk.ZeroDec())
			previousRewardIndexes = append(previousRewardIndexes, rewardIndex)
		}
		k.SetHardSupplyRewardIndexes(ctx, rewardPeriod.CollateralType, previousRewardIndexes)
	}
	hardFactor, found := k.hardKeeper.GetSupplyInterestFactor(ctx, rewardPeriod.CollateralType)
	if !found {
		k.SetPreviousHardSupplyRewardAccrualTime(ctx, rewardPeriod.CollateralType, ctx.BlockTime())
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
		rewardFactor := newRewards.Mul(hardFactor).Quo(totalSupplied)
		newRewardFactorValue := previousRewardIndex.RewardFactor.Add(rewardFactor)
		newRewardIndex := types.NewRewardIndex(rewardCoin.Denom, newRewardFactorValue)
		i, found := newRewardIndexes.GetFactorIndex(rewardCoin.Denom)
		if found {
			newRewardIndexes[i] = newRewardIndex
		} else {
			newRewardIndexes = append(newRewardIndexes, newRewardIndex)
		}
	}
	k.SetHardSupplyRewardIndexes(ctx, rewardPeriod.CollateralType, newRewardIndexes)
	k.SetPreviousHardSupplyRewardAccrualTime(ctx, rewardPeriod.CollateralType, ctx.BlockTime())
	return nil
}

// InitializeHardSupplyReward initializes the supply-side of a hard liquidity provider claim
// by creating the claim and setting the supply reward factor index
func (k Keeper) InitializeHardSupplyReward(ctx sdk.Context, deposit hardtypes.Deposit) {
	var supplyRewardIndexes types.MultiRewardIndexes
	for _, coin := range deposit.Amount {
		globalRewardIndexes, foundGlobalRewardIndexes := k.GetHardSupplyRewardIndexes(ctx, coin.Denom)
		var multiRewardIndex types.MultiRewardIndex
		if foundGlobalRewardIndexes {
			multiRewardIndex = types.NewMultiRewardIndex(coin.Denom, globalRewardIndexes)
		} else {
			multiRewardIndex = types.NewMultiRewardIndex(coin.Denom, types.RewardIndexes{})
		}
		supplyRewardIndexes = append(supplyRewardIndexes, multiRewardIndex)
	}

	claim, found := k.GetHardLiquidityProviderClaim(ctx, deposit.Depositor)
	if !found {
		// Instantiate claim object
		claim = types.NewHardLiquidityProviderClaim(deposit.Depositor, sdk.Coins{}, nil, nil, nil)
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
		globalRewardIndexes, foundGlobalRewardIndexes := k.GetHardSupplyRewardIndexes(ctx, coin.Denom)
		if !foundGlobalRewardIndexes {
			continue
		}

		userMultiRewardIndex, foundUserMultiRewardIndex := claim.SupplyRewardIndexes.GetRewardIndex(coin.Denom)
		if !foundUserMultiRewardIndex {
			continue
		}

		userRewardIndexIndex, foundUserRewardIndexIndex := claim.SupplyRewardIndexes.GetRewardIndexIndex(coin.Denom)
		if !foundUserRewardIndexIndex {
			continue
		}

		for _, globalRewardIndex := range globalRewardIndexes {
			userRewardIndex, foundUserRewardIndex := userMultiRewardIndex.RewardIndexes.GetRewardIndex(globalRewardIndex.CollateralType)
			if !foundUserRewardIndex {
				// User deposited this coin type before it had rewards. When new rewards are added, legacy depositors
				// should immediately begin earning rewards. Enable users to do so by updating their claim with the global
				// reward index denom and start their reward factor at 0.0
				userRewardIndex = types.NewRewardIndex(globalRewardIndex.CollateralType, sdk.ZeroDec())
				userMultiRewardIndex.RewardIndexes = append(userMultiRewardIndex.RewardIndexes, userRewardIndex)
				claim.SupplyRewardIndexes[userRewardIndexIndex] = userMultiRewardIndex
			}

			globalRewardFactor := globalRewardIndex.RewardFactor
			userRewardFactor := userRewardIndex.RewardFactor
			rewardsAccumulatedFactor := globalRewardFactor.Sub(userRewardFactor)
			if rewardsAccumulatedFactor.IsNegative() {
				panic(fmt.Sprintf("reward accumulation factor cannot be negative: %s", rewardsAccumulatedFactor))
			}

			newRewardsAmount := rewardsAccumulatedFactor.Mul(deposit.Amount.AmountOf(coin.Denom).ToDec()).RoundInt()

			factorIndex, foundFactorIndex := userMultiRewardIndex.RewardIndexes.GetFactorIndex(globalRewardIndex.CollateralType)
			if !foundFactorIndex { // should never trigger, as we basically do this check at the start of this loop
				continue
			}
			claim.SupplyRewardIndexes[userRewardIndexIndex].RewardIndexes[factorIndex].RewardFactor = globalRewardIndex.RewardFactor

			newRewardsCoin := sdk.NewCoin(userRewardIndex.CollateralType, newRewardsAmount)
			claim.Reward = claim.Reward.Add(newRewardsCoin)
		}
	}
	k.SetHardLiquidityProviderClaim(ctx, claim)
}

// UpdateHardSupplyIndexDenoms adds any new deposit denoms to the claim's supply reward index
func (k Keeper) UpdateHardSupplyIndexDenoms(ctx sdk.Context, deposit hardtypes.Deposit) {
	claim, found := k.GetHardLiquidityProviderClaim(ctx, deposit.Depositor)
	if !found {
		claim = types.NewHardLiquidityProviderClaim(deposit.Depositor, sdk.Coins{}, nil, nil, nil)
	}

	depositDenoms := getDenoms(deposit.Amount)
	supplyRewardIndexDenoms := claim.SupplyRewardIndexes.GetCollateralTypes()

	uniqueDepositDenoms := setDifference(depositDenoms, supplyRewardIndexDenoms)
	uniqueSupplyRewardDenoms := setDifference(supplyRewardIndexDenoms, depositDenoms)

	supplyRewardIndexes := claim.SupplyRewardIndexes
	// Create a new multi-reward index in the claim for every new deposit denom
	for _, denom := range uniqueDepositDenoms {
		_, foundUserRewardIndexes := claim.SupplyRewardIndexes.GetRewardIndex(denom)
		if !foundUserRewardIndexes {
			globalSupplyRewardIndexes, foundGlobalSupplyRewardIndexes := k.GetHardSupplyRewardIndexes(ctx, denom)
			var multiRewardIndex types.MultiRewardIndex
			if foundGlobalSupplyRewardIndexes {
				multiRewardIndex = types.NewMultiRewardIndex(denom, globalSupplyRewardIndexes)
			} else {
				multiRewardIndex = types.NewMultiRewardIndex(denom, types.RewardIndexes{})
			}
			supplyRewardIndexes = append(supplyRewardIndexes, multiRewardIndex)
		}
	}

	// Delete multi-reward index from claim if the collateral type is no longer deposited
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

	// Synchronize any hard delegator rewards
	k.SynchronizeHardDelegatorRewards(ctx, owner, nil, false)
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

	// 3. Simulate Hard delegator rewards
	delagatorFactor, found := k.GetHardDelegatorRewardFactor(ctx, types.BondDenom)
	if !found {
		return claim
	}

	delegatorIndex, hasDelegatorRewardIndex := claim.HasDelegatorRewardIndex(types.BondDenom)
	if !hasDelegatorRewardIndex {
		return claim
	}

	userRewardFactor := claim.DelegatorRewardIndexes[delegatorIndex].RewardFactor
	rewardsAccumulatedFactor := delagatorFactor.Sub(userRewardFactor)
	if rewardsAccumulatedFactor.IsZero() {
		return claim
	}
	claim.DelegatorRewardIndexes[delegatorIndex].RewardFactor = delagatorFactor

	totalDelegated := sdk.ZeroDec()

	delegations := k.stakingKeeper.GetDelegatorDelegations(ctx, claim.GetOwner(), 200)
	for _, delegation := range delegations {
		validator, found := k.stakingKeeper.GetValidator(ctx, delegation.GetValidatorAddr())
		if !found {
			continue
		}

		// Delegators don't accumulate rewards if their validator is unbonded/slashed
		if validator.GetStatus() != sdk.Bonded {
			continue
		}

		if validator.GetTokens().IsZero() {
			continue
		}

		delegatedTokens := validator.TokensFromShares(delegation.GetShares())
		if delegatedTokens.IsZero() || delegatedTokens.IsNegative() {
			continue
		}
		totalDelegated = totalDelegated.Add(delegatedTokens)
	}

	rewardsEarned := rewardsAccumulatedFactor.Mul(totalDelegated).RoundInt()
	if rewardsEarned.IsZero() || rewardsEarned.IsNegative() {
		return claim
	}

	// Add rewards to delegator's hard claim
	newRewardsCoin := sdk.NewCoin(types.HardLiquidityRewardDenom, rewardsEarned)
	claim.Reward = claim.Reward.Add(newRewardsCoin)

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
