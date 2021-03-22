package keeper

import (
	"fmt"
	"math"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	hardtypes "github.com/kava-labs/kava/x/hard/types"
	"github.com/kava-labs/kava/x/incentive/types"
)

// AccumulateUSDXMintingRewards updates the rewards accumulated for the input reward period
func (k Keeper) AccumulateUSDXMintingRewards(ctx sdk.Context, rewardPeriod types.RewardPeriod) error {
	previousAccrualTime, found := k.GetPreviousUSDXMintingAccrualTime(ctx, rewardPeriod.CollateralType)
	if !found {
		k.SetPreviousUSDXMintingAccrualTime(ctx, rewardPeriod.CollateralType, ctx.BlockTime())
		return nil
	}
	timeElapsed := CalculateTimeElapsed(rewardPeriod.Start, rewardPeriod.End, ctx.BlockTime(), previousAccrualTime)
	if timeElapsed.IsZero() {
		return nil
	}
	if rewardPeriod.RewardsPerSecond.Amount.IsZero() {
		k.SetPreviousUSDXMintingAccrualTime(ctx, rewardPeriod.CollateralType, ctx.BlockTime())
		return nil
	}
	totalPrincipal := k.cdpKeeper.GetTotalPrincipal(ctx, rewardPeriod.CollateralType, types.PrincipalDenom).ToDec()
	if totalPrincipal.IsZero() {
		k.SetPreviousUSDXMintingAccrualTime(ctx, rewardPeriod.CollateralType, ctx.BlockTime())
		return nil
	}
	newRewards := timeElapsed.Mul(rewardPeriod.RewardsPerSecond.Amount)
	cdpFactor, found := k.cdpKeeper.GetInterestFactor(ctx, rewardPeriod.CollateralType)
	if !found {
		k.SetPreviousUSDXMintingAccrualTime(ctx, rewardPeriod.CollateralType, ctx.BlockTime())
		return nil
	}
	rewardFactor := newRewards.ToDec().Mul(cdpFactor).Quo(totalPrincipal)

	previousRewardFactor, found := k.GetUSDXMintingRewardFactor(ctx, rewardPeriod.CollateralType)
	if !found {
		previousRewardFactor = sdk.ZeroDec()
	}
	newRewardFactor := previousRewardFactor.Add(rewardFactor)
	k.SetUSDXMintingRewardFactor(ctx, rewardPeriod.CollateralType, newRewardFactor)
	k.SetPreviousUSDXMintingAccrualTime(ctx, rewardPeriod.CollateralType, ctx.BlockTime())
	return nil
}

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

// InitializeUSDXMintingClaim creates or updates a claim such that no new rewards are accrued, but any existing rewards are not lost.
// this function should be called after a cdp is created. If a user previously had a cdp, then closed it, they shouldn't
// accrue rewards during the period the cdp was closed. By setting the reward factor to the current global reward factor,
// any unclaimed rewards are preserved, but no new rewards are added.
func (k Keeper) InitializeUSDXMintingClaim(ctx sdk.Context, cdp cdptypes.CDP) {
	_, found := k.GetUSDXMintingRewardPeriod(ctx, cdp.Type)
	if !found {
		// this collateral type is not incentivized, do nothing
		return
	}
	rewardFactor, found := k.GetUSDXMintingRewardFactor(ctx, cdp.Type)
	if !found {
		rewardFactor = sdk.ZeroDec()
	}
	claim, found := k.GetUSDXMintingClaim(ctx, cdp.Owner)
	if !found { // this is the owner's first usdx minting reward claim
		claim = types.NewUSDXMintingClaim(cdp.Owner, sdk.NewCoin(types.USDXMintingRewardDenom, sdk.ZeroInt()), types.RewardIndexes{types.NewRewardIndex(cdp.Type, rewardFactor)})
		k.SetUSDXMintingClaim(ctx, claim)
		return
	}
	// the owner has an existing usdx minting reward claim
	index, hasRewardIndex := claim.HasRewardIndex(cdp.Type)
	if !hasRewardIndex { // this is the owner's first usdx minting reward for this collateral type
		claim.RewardIndexes = append(claim.RewardIndexes, types.NewRewardIndex(cdp.Type, rewardFactor))
	} else { // the owner has a previous usdx minting reward for this collateral type
		claim.RewardIndexes[index] = types.NewRewardIndex(cdp.Type, rewardFactor)
	}
	k.SetUSDXMintingClaim(ctx, claim)
}

// SynchronizeUSDXMintingReward updates the claim object by adding any accumulated rewards and updating the reward index value.
// this should be called before a cdp is modified, immediately after the 'SynchronizeInterest' method is called in the cdp module
func (k Keeper) SynchronizeUSDXMintingReward(ctx sdk.Context, cdp cdptypes.CDP) {
	_, found := k.GetUSDXMintingRewardPeriod(ctx, cdp.Type)
	if !found {
		// this collateral type is not incentivized, do nothing
		return
	}

	globalRewardFactor, found := k.GetUSDXMintingRewardFactor(ctx, cdp.Type)
	if !found {
		globalRewardFactor = sdk.ZeroDec()
	}
	claim, found := k.GetUSDXMintingClaim(ctx, cdp.Owner)
	if !found {
		claim = types.NewUSDXMintingClaim(cdp.Owner, sdk.NewCoin(types.USDXMintingRewardDenom, sdk.ZeroInt()), types.RewardIndexes{types.NewRewardIndex(cdp.Type, globalRewardFactor)})
		k.SetUSDXMintingClaim(ctx, claim)
		return
	}

	// the owner has an existing usdx minting reward claim
	index, hasRewardIndex := claim.HasRewardIndex(cdp.Type)
	if !hasRewardIndex { // this is the owner's first usdx minting reward for this collateral type
		claim.RewardIndexes = append(claim.RewardIndexes, types.NewRewardIndex(cdp.Type, globalRewardFactor))
		k.SetUSDXMintingClaim(ctx, claim)
		return
	}
	userRewardFactor := claim.RewardIndexes[index].RewardFactor
	rewardsAccumulatedFactor := globalRewardFactor.Sub(userRewardFactor)
	if rewardsAccumulatedFactor.IsZero() {
		return
	}
	claim.RewardIndexes[index].RewardFactor = globalRewardFactor
	newRewardsAmount := rewardsAccumulatedFactor.Mul(cdp.GetTotalPrincipal().Amount.ToDec()).RoundInt()
	if newRewardsAmount.IsZero() {
		k.SetUSDXMintingClaim(ctx, claim)
		return
	}
	newRewardsCoin := sdk.NewCoin(types.USDXMintingRewardDenom, newRewardsAmount)
	claim.Reward = claim.Reward.Add(newRewardsCoin)
	k.SetUSDXMintingClaim(ctx, claim)
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
	if found {
		// Reset borrow reward indexes
		claim.BorrowRewardIndexes = types.MultiRewardIndexes{}
	} else {
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

			globalRewardFactor := globalRewardIndex.RewardFactor
			userRewardFactor := userRewardIndex.RewardFactor
			rewardsAccumulatedFactor := globalRewardFactor.Sub(userRewardFactor)
			if rewardsAccumulatedFactor.IsZero() {
				continue
			}
			newRewardsAmount := rewardsAccumulatedFactor.Mul(borrow.Amount.AmountOf(coin.Denom).ToDec()).RoundInt()
			if newRewardsAmount.IsZero() || newRewardsAmount.IsNegative() {
				continue
			}

			factorIndex, foundFactorIndex := userMultiRewardIndex.RewardIndexes.GetFactorIndex(globalRewardIndex.CollateralType)
			if !foundFactorIndex {
				continue
			}
			claim.BorrowRewardIndexes[userRewardIndexIndex].RewardIndexes[factorIndex].RewardFactor = globalRewardIndex.RewardFactor
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

	supplyRewardIndexes := claim.SupplyRewardIndexes
	for _, coin := range deposit.Amount {
		_, foundUserRewardIndexes := claim.SupplyRewardIndexes.GetRewardIndex(coin.Denom)
		if !foundUserRewardIndexes {
			globalSupplyRewardIndexes, foundGlobalSupplyRewardIndexes := k.GetHardSupplyRewardIndexes(ctx, coin.Denom)
			var multiRewardIndex types.MultiRewardIndex
			if foundGlobalSupplyRewardIndexes {
				multiRewardIndex = types.NewMultiRewardIndex(coin.Denom, globalSupplyRewardIndexes)
			} else {
				multiRewardIndex = types.NewMultiRewardIndex(coin.Denom, types.RewardIndexes{})
			}
			supplyRewardIndexes = append(supplyRewardIndexes, multiRewardIndex)
		}
	}
	if len(supplyRewardIndexes) == 0 {
		return
	}
	claim.SupplyRewardIndexes = supplyRewardIndexes
	k.SetHardLiquidityProviderClaim(ctx, claim)
}

// UpdateHardBorrowIndexDenoms adds any new borrow denoms to the claim's supply reward index
func (k Keeper) UpdateHardBorrowIndexDenoms(ctx sdk.Context, borrow hardtypes.Borrow) {
	claim, found := k.GetHardLiquidityProviderClaim(ctx, borrow.Borrower)
	if !found {
		claim = types.NewHardLiquidityProviderClaim(borrow.Borrower, sdk.Coins{}, nil, nil, nil)
	}

	borrowRewardIndexes := claim.BorrowRewardIndexes
	for _, coin := range borrow.Amount {
		_, foundUserRewardIndexes := claim.BorrowRewardIndexes.GetRewardIndex(coin.Denom)
		if !foundUserRewardIndexes {
			globalBorrowRewardIndexes, foundGlobalBorrowRewardIndexes := k.GetHardBorrowRewardIndexes(ctx, coin.Denom)
			var multiRewardIndex types.MultiRewardIndex
			if foundGlobalBorrowRewardIndexes {
				multiRewardIndex = types.NewMultiRewardIndex(coin.Denom, globalBorrowRewardIndexes)
			} else {
				multiRewardIndex = types.NewMultiRewardIndex(coin.Denom, types.RewardIndexes{})
			}
			borrowRewardIndexes = append(borrowRewardIndexes, multiRewardIndex)
		}
	}
	if len(borrowRewardIndexes) == 0 {
		return
	}
	claim.BorrowRewardIndexes = borrowRewardIndexes
	k.SetHardLiquidityProviderClaim(ctx, claim)
}

// SynchronizeHardDelegatorRewards updates the claim object by adding any accumulated rewards
func (k Keeper) SynchronizeHardDelegatorRewards(ctx sdk.Context, delegator sdk.AccAddress) {
	claim, found := k.GetHardLiquidityProviderClaim(ctx, delegator)
	if !found {
		return
	}

	delagatorFactor, found := k.GetHardDelegatorRewardFactor(ctx, types.BondDenom)
	if !found {
		return
	}

	delegatorIndex, hasDelegatorRewardIndex := claim.HasDelegatorRewardIndex(types.BondDenom)
	if !hasDelegatorRewardIndex {
		return
	}

	userRewardFactor := claim.DelegatorRewardIndexes[delegatorIndex].RewardFactor
	rewardsAccumulatedFactor := delagatorFactor.Sub(userRewardFactor)
	if rewardsAccumulatedFactor.IsZero() {
		return
	}
	claim.DelegatorRewardIndexes[delegatorIndex].RewardFactor = delagatorFactor

	totalDelegated := sdk.ZeroDec()

	delegations := k.stakingKeeper.GetDelegatorDelegations(ctx, delegator, 200)
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
		return
	}

	// Add rewards to delegator's hard claim
	newRewardsCoin := sdk.NewCoin(types.HardLiquidityRewardDenom, rewardsEarned)
	claim.Reward = claim.Reward.Add(newRewardsCoin)
	k.SetHardLiquidityProviderClaim(ctx, claim)
}

// AccumulateHardDelegatorRewards updates the rewards accumulated for the input reward period
func (k Keeper) AccumulateHardDelegatorRewards(ctx sdk.Context, rewardPeriod types.RewardPeriod) error {
	previousAccrualTime, found := k.GetPreviousHardDelegatorRewardAccrualTime(ctx, rewardPeriod.CollateralType)
	if !found {
		k.SetPreviousHardDelegatorRewardAccrualTime(ctx, rewardPeriod.CollateralType, ctx.BlockTime())
		return nil
	}
	timeElapsed := CalculateTimeElapsed(rewardPeriod.Start, rewardPeriod.End, ctx.BlockTime(), previousAccrualTime)
	if timeElapsed.IsZero() {
		return nil
	}
	if rewardPeriod.RewardsPerSecond.Amount.IsZero() {
		k.SetPreviousHardDelegatorRewardAccrualTime(ctx, rewardPeriod.CollateralType, ctx.BlockTime())
		return nil
	}

	totalBonded := k.stakingKeeper.TotalBondedTokens(ctx).ToDec()
	if totalBonded.IsZero() {
		k.SetPreviousHardDelegatorRewardAccrualTime(ctx, rewardPeriod.CollateralType, ctx.BlockTime())
		return nil
	}

	newRewards := timeElapsed.Mul(rewardPeriod.RewardsPerSecond.Amount)
	rewardFactor := newRewards.ToDec().Quo(totalBonded)

	previousRewardFactor, found := k.GetHardDelegatorRewardFactor(ctx, rewardPeriod.CollateralType)
	if !found {
		previousRewardFactor = sdk.ZeroDec()
	}
	newRewardFactor := previousRewardFactor.Add(rewardFactor)
	k.SetHardDelegatorRewardFactor(ctx, rewardPeriod.CollateralType, newRewardFactor)
	k.SetPreviousHardDelegatorRewardAccrualTime(ctx, rewardPeriod.CollateralType, ctx.BlockTime())
	return nil
}

// InitializeHardDelegatorReward initializes the delegator reward index of a hard claim
func (k Keeper) InitializeHardDelegatorReward(ctx sdk.Context, delegator sdk.AccAddress) {
	delegatorFactor, foundDelegatorFactor := k.GetHardDelegatorRewardFactor(ctx, types.BondDenom)
	if !foundDelegatorFactor { // Should always be found...
		delegatorFactor = sdk.ZeroDec()
	}

	delegatorRewardIndexes := types.NewRewardIndex(types.BondDenom, delegatorFactor)

	claim, found := k.GetHardLiquidityProviderClaim(ctx, delegator)
	if !found {
		// Instantiate claim object
		claim = types.NewHardLiquidityProviderClaim(delegator, sdk.Coins{}, nil, nil, nil)
	} else {
		k.SynchronizeHardDelegatorRewards(ctx, delegator)
		claim, _ = k.GetHardLiquidityProviderClaim(ctx, delegator)
	}

	claim.DelegatorRewardIndexes = types.RewardIndexes{delegatorRewardIndexes}
	k.SetHardLiquidityProviderClaim(ctx, claim)
}

// ZeroUSDXMintingClaim zeroes out the claim object's rewards and returns the updated claim object
func (k Keeper) ZeroUSDXMintingClaim(ctx sdk.Context, claim types.USDXMintingClaim) types.USDXMintingClaim {
	claim.Reward = sdk.NewCoin(claim.Reward.Denom, sdk.ZeroInt())
	k.SetUSDXMintingClaim(ctx, claim)
	return claim
}

// SynchronizeUSDXMintingClaim updates the claim object by adding any rewards that have accumulated.
// Returns the updated claim object
func (k Keeper) SynchronizeUSDXMintingClaim(ctx sdk.Context, claim types.USDXMintingClaim) (types.USDXMintingClaim, error) {
	for _, ri := range claim.RewardIndexes {
		cdp, found := k.cdpKeeper.GetCdpByOwnerAndCollateralType(ctx, claim.Owner, ri.CollateralType)
		if !found {
			// if the cdp for this collateral type has been closed, no updates are needed
			continue
		}
		claim = k.synchronizeRewardAndReturnClaim(ctx, cdp)
	}
	return claim, nil
}

// this function assumes a claim already exists, so don't call it if that's not the case
func (k Keeper) synchronizeRewardAndReturnClaim(ctx sdk.Context, cdp cdptypes.CDP) types.USDXMintingClaim {
	k.SynchronizeUSDXMintingReward(ctx, cdp)
	claim, _ := k.GetUSDXMintingClaim(ctx, cdp.Owner)
	return claim
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
	k.SynchronizeHardDelegatorRewards(ctx, owner)
}

// ZeroHardLiquidityProviderClaim zeroes out the claim object's rewards and returns the updated claim object
func (k Keeper) ZeroHardLiquidityProviderClaim(ctx sdk.Context, claim types.HardLiquidityProviderClaim) types.HardLiquidityProviderClaim {
	var zeroRewards sdk.Coins
	for _, coin := range claim.Reward {
		zeroRewards = append(zeroRewards, sdk.NewCoin(coin.Denom, sdk.ZeroInt()))
	}
	claim.Reward = zeroRewards
	k.SetHardLiquidityProviderClaim(ctx, claim)
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

// SimulateUSDXMintingSynchronization calculates a user's outstanding USDX minting rewards by simulating reward synchronization
func (k Keeper) SimulateUSDXMintingSynchronization(ctx sdk.Context, claim types.USDXMintingClaim) types.USDXMintingClaim {
	for _, ri := range claim.RewardIndexes {
		_, found := k.GetUSDXMintingRewardPeriod(ctx, ri.CollateralType)
		if !found {
			continue
		}

		globalRewardFactor, found := k.GetUSDXMintingRewardFactor(ctx, ri.CollateralType)
		if !found {
			globalRewardFactor = sdk.ZeroDec()
		}

		// the owner has an existing usdx minting reward claim
		index, hasRewardIndex := claim.HasRewardIndex(ri.CollateralType)
		if !hasRewardIndex { // this is the owner's first usdx minting reward for this collateral type
			claim.RewardIndexes = append(claim.RewardIndexes, types.NewRewardIndex(ri.CollateralType, globalRewardFactor))
		}
		userRewardFactor := claim.RewardIndexes[index].RewardFactor
		rewardsAccumulatedFactor := globalRewardFactor.Sub(userRewardFactor)
		if rewardsAccumulatedFactor.IsZero() {
			continue
		}

		claim.RewardIndexes[index].RewardFactor = globalRewardFactor

		cdp, found := k.cdpKeeper.GetCdpByOwnerAndCollateralType(ctx, claim.GetOwner(), ri.CollateralType)
		if !found {
			continue
		}
		newRewardsAmount := rewardsAccumulatedFactor.Mul(cdp.GetTotalPrincipal().Amount.ToDec()).RoundInt()
		if newRewardsAmount.IsZero() {
			continue
		}
		newRewardsCoin := sdk.NewCoin(types.USDXMintingRewardDenom, newRewardsAmount)
		claim.Reward = claim.Reward.Add(newRewardsCoin)
	}

	return claim
}
