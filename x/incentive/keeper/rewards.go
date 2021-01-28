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
	timeElapsed := CalculateTimeElapsed(rewardPeriod, ctx.BlockTime(), previousAccrualTime)
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
func (k Keeper) AccumulateHardBorrowRewards(ctx sdk.Context, rewardPeriod types.RewardPeriod) error {
	previousAccrualTime, found := k.GetPreviousHardBorrowRewardAccrualTime(ctx, rewardPeriod.CollateralType)
	if !found {
		k.SetPreviousHardBorrowRewardAccrualTime(ctx, rewardPeriod.CollateralType, ctx.BlockTime())
		return nil
	}
	timeElapsed := CalculateTimeElapsed(rewardPeriod, ctx.BlockTime(), previousAccrualTime)
	if timeElapsed.IsZero() {
		return nil
	}
	if rewardPeriod.RewardsPerSecond.Amount.IsZero() {
		k.SetPreviousHardBorrowRewardAccrualTime(ctx, rewardPeriod.CollateralType, ctx.BlockTime())
		return nil
	}
	totalBorrowedCoins, foundTotalBorrowedCoins := k.hardKeeper.GetBorrowedCoins(ctx)
	if foundTotalBorrowedCoins {
		totalBorrowed := totalBorrowedCoins.AmountOf(rewardPeriod.CollateralType).ToDec()
		if totalBorrowed.IsZero() {
			k.SetPreviousHardBorrowRewardAccrualTime(ctx, rewardPeriod.CollateralType, ctx.BlockTime())
			return nil
		}
		newRewards := timeElapsed.Mul(rewardPeriod.RewardsPerSecond.Amount)
		hardFactor, found := k.hardKeeper.GetBorrowInterestFactor(ctx, rewardPeriod.CollateralType)
		if !found {
			k.SetPreviousHardBorrowRewardAccrualTime(ctx, rewardPeriod.CollateralType, ctx.BlockTime())
			return nil
		}
		rewardFactor := newRewards.ToDec().Mul(hardFactor).Quo(totalBorrowed)

		previousRewardFactor, found := k.GetHardBorrowRewardFactor(ctx, rewardPeriod.CollateralType)
		if !found {
			previousRewardFactor = sdk.ZeroDec()
		}
		newRewardFactor := previousRewardFactor.Add(rewardFactor)
		k.SetHardBorrowRewardFactor(ctx, rewardPeriod.CollateralType, newRewardFactor)
	}
	k.SetPreviousHardBorrowRewardAccrualTime(ctx, rewardPeriod.CollateralType, ctx.BlockTime())

	return nil
}

// AccumulateHardSupplyRewards updates the rewards accumulated for the input reward period
func (k Keeper) AccumulateHardSupplyRewards(ctx sdk.Context, rewardPeriod types.RewardPeriod) error {
	previousAccrualTime, found := k.GetPreviousHardSupplyRewardAccrualTime(ctx, rewardPeriod.CollateralType)
	if !found {
		k.SetPreviousHardSupplyRewardAccrualTime(ctx, rewardPeriod.CollateralType, ctx.BlockTime())
		return nil
	}
	timeElapsed := CalculateTimeElapsed(rewardPeriod, ctx.BlockTime(), previousAccrualTime)
	if timeElapsed.IsZero() {
		return nil
	}
	if rewardPeriod.RewardsPerSecond.Amount.IsZero() {
		k.SetPreviousHardSupplyRewardAccrualTime(ctx, rewardPeriod.CollateralType, ctx.BlockTime())
		return nil
	}

	totalSuppliedCoins, foundTotalSuppliedCoins := k.hardKeeper.GetSuppliedCoins(ctx)
	if foundTotalSuppliedCoins {
		totalSupplied := totalSuppliedCoins.AmountOf(rewardPeriod.CollateralType).ToDec()
		if totalSupplied.IsZero() {
			k.SetPreviousHardSupplyRewardAccrualTime(ctx, rewardPeriod.CollateralType, ctx.BlockTime())
			return nil
		}
		newRewards := timeElapsed.Mul(rewardPeriod.RewardsPerSecond.Amount)
		hardFactor, found := k.hardKeeper.GetSupplyInterestFactor(ctx, rewardPeriod.CollateralType)
		if !found {
			k.SetPreviousHardSupplyRewardAccrualTime(ctx, rewardPeriod.CollateralType, ctx.BlockTime())
			return nil
		}
		rewardFactor := newRewards.ToDec().Mul(hardFactor).Quo(totalSupplied)

		previousRewardFactor, found := k.GetHardSupplyRewardFactor(ctx, rewardPeriod.CollateralType)
		if !found {
			previousRewardFactor = sdk.ZeroDec()
		}
		newRewardFactor := previousRewardFactor.Add(rewardFactor)
		k.SetHardSupplyRewardFactor(ctx, rewardPeriod.CollateralType, newRewardFactor)
	}
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
	return
}

// InitializeHardSupplyReward initializes the supply-side of a hard liquidity provider claim
// by creating the claim and setting the supply reward factor index
func (k Keeper) InitializeHardSupplyReward(ctx sdk.Context, deposit hardtypes.Deposit) {
	var supplyRewardIndexes types.RewardIndexes
	for _, coin := range deposit.Amount {
		_, rpFound := k.GetHardSupplyRewardPeriod(ctx, coin.Denom)
		if !rpFound {
			continue
		}

		supplyFactor, foundSupplyFactor := k.GetHardSupplyRewardFactor(ctx, coin.Denom)
		if !foundSupplyFactor {
			supplyFactor = sdk.ZeroDec()
		}

		supplyRewardIndexes = append(supplyRewardIndexes, types.NewRewardIndex(coin.Denom, supplyFactor))
	}

	claim, found := k.GetHardLiquidityProviderClaim(ctx, deposit.Depositor)
	if found {
		// Reset borrow reward indexes
		claim.BorrowRewardIndexes = types.RewardIndexes{}
	} else {
		// Instantiate claim object
		claim = types.NewHardLiquidityProviderClaim(deposit.Depositor,
			sdk.NewCoin(types.HardLiquidityRewardDenom, sdk.ZeroInt()),
			nil, nil, nil)
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
		supplyFactor, found := k.GetHardSupplyRewardFactor(ctx, coin.Denom)
		if !found {
			fmt.Printf("\n[LOG]: %s does not have a supply factor", coin.Denom) // TODO: remove before production
			continue
		}

		supplyIndex, hasSupplyRewardIndex := claim.HasSupplyRewardIndex(coin.Denom)
		if !hasSupplyRewardIndex {
			continue
		}

		userRewardFactor := claim.SupplyRewardIndexes[supplyIndex].RewardFactor
		rewardsAccumulatedFactor := supplyFactor.Sub(userRewardFactor)
		if rewardsAccumulatedFactor.IsZero() {
			continue
		}
		claim.SupplyRewardIndexes[supplyIndex].RewardFactor = supplyFactor

		newRewardsAmount := rewardsAccumulatedFactor.Mul(deposit.Amount.AmountOf(coin.Denom).ToDec()).RoundInt()
		if newRewardsAmount.IsZero() || newRewardsAmount.IsNegative() {
			continue
		}

		newRewardsCoin := sdk.NewCoin(types.HardLiquidityRewardDenom, newRewardsAmount)
		claim.Reward = claim.Reward.Add(newRewardsCoin)
	}

	k.SetHardLiquidityProviderClaim(ctx, claim)
}

// InitializeHardBorrowReward initializes the borrow-side of a hard liquidity provider claim
// by creating the claim and setting the borrow reward factor index
func (k Keeper) InitializeHardBorrowReward(ctx sdk.Context, borrow hardtypes.Borrow) {
	claim, found := k.GetHardLiquidityProviderClaim(ctx, borrow.Borrower)
	if !found {
		claim = types.NewHardLiquidityProviderClaim(borrow.Borrower,
			sdk.NewCoin(types.HardLiquidityRewardDenom, sdk.ZeroInt()),
			nil, nil, nil)
	}

	var borrowRewardIndexes types.RewardIndexes
	for _, coin := range borrow.Amount {
		_, rpFound := k.GetHardBorrowRewardPeriod(ctx, coin.Denom)
		if !rpFound {
			continue
		}

		borrowFactor, foundBorrowFactor := k.GetHardBorrowRewardFactor(ctx, coin.Denom)
		if !foundBorrowFactor {
			borrowFactor = sdk.ZeroDec()
		}

		borrowRewardIndexes = append(borrowRewardIndexes, types.NewRewardIndex(coin.Denom, borrowFactor))
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
		borrowFactor, found := k.GetHardBorrowRewardFactor(ctx, coin.Denom)
		if !found {
			continue
		}

		borrowIndex, BorrowRewardIndex := claim.HasBorrowRewardIndex(coin.Denom)
		if !BorrowRewardIndex {
			continue
		}

		userRewardFactor := claim.BorrowRewardIndexes[borrowIndex].RewardFactor
		rewardsAccumulatedFactor := borrowFactor.Sub(userRewardFactor)
		if rewardsAccumulatedFactor.IsZero() {
			continue
		}
		claim.BorrowRewardIndexes[borrowIndex].RewardFactor = borrowFactor

		newRewardsAmount := rewardsAccumulatedFactor.Mul(borrow.Amount.AmountOf(coin.Denom).ToDec()).RoundInt()
		if newRewardsAmount.IsZero() || newRewardsAmount.IsNegative() {
			continue
		}

		newRewardsCoin := sdk.NewCoin(types.HardLiquidityRewardDenom, newRewardsAmount)
		claim.Reward = claim.Reward.Add(newRewardsCoin)
	}

	k.SetHardLiquidityProviderClaim(ctx, claim)
}

// UpdateHardSupplyIndexDenoms adds any new deposit denoms to the claim's supply reward index
func (k Keeper) UpdateHardSupplyIndexDenoms(ctx sdk.Context, deposit hardtypes.Deposit) {
	claim, found := k.GetHardLiquidityProviderClaim(ctx, deposit.Depositor)
	if !found {
		claim = types.NewHardLiquidityProviderClaim(deposit.Depositor,
			sdk.NewCoin(types.HardLiquidityRewardDenom, sdk.ZeroInt()),
			nil, nil, nil)
	}

	supplyRewardIndexes := claim.SupplyRewardIndexes
	for _, coin := range deposit.Amount {
		_, hasIndex := claim.HasSupplyRewardIndex(coin.Denom)
		if !hasIndex {
			supplyFactor, foundSupplyFactor := k.GetHardSupplyRewardFactor(ctx, coin.Denom)
			if foundSupplyFactor {
				supplyRewardIndexes = append(supplyRewardIndexes, types.NewRewardIndex(coin.Denom, supplyFactor))
			}
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
		claim = types.NewHardLiquidityProviderClaim(borrow.Borrower,
			sdk.NewCoin(types.HardLiquidityRewardDenom, sdk.ZeroInt()),
			nil, nil, nil)
	}

	borrowRewardIndexes := claim.BorrowRewardIndexes
	for _, coin := range borrow.Amount {
		_, hasIndex := claim.HasBorrowRewardIndex(coin.Denom)
		if !hasIndex {
			borrowFactor, foundBorrowFactor := k.GetHardBorrowRewardFactor(ctx, coin.Denom)
			if foundBorrowFactor {
				borrowRewardIndexes = append(borrowRewardIndexes, types.NewRewardIndex(coin.Denom, borrowFactor))
			}
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

	// TODO: set reasonable max limit on delegation iteration
	maxUInt := ^uint16(0)
	delegations := k.stakingKeeper.GetDelegatorDelegations(ctx, delegator, maxUInt)
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
	timeElapsed := CalculateTimeElapsed(rewardPeriod, ctx.BlockTime(), previousAccrualTime)
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
		claim = types.NewHardLiquidityProviderClaim(delegator,
			sdk.NewCoin(types.HardLiquidityRewardDenom, sdk.ZeroInt()),
			nil, nil, nil)
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
	claim.Reward = sdk.NewCoin(claim.Reward.Denom, sdk.ZeroInt())
	k.SetHardLiquidityProviderClaim(ctx, claim)
	return claim
}

// CalculateTimeElapsed calculates the number of reward-eligible seconds that have passed since the previous
// time rewards were accrued, taking into account the end time of the reward period
func CalculateTimeElapsed(rewardPeriod types.RewardPeriod, blockTime time.Time, previousAccrualTime time.Time) sdk.Int {
	if rewardPeriod.End.Before(blockTime) &&
		(rewardPeriod.End.Before(previousAccrualTime) || rewardPeriod.End.Equal(previousAccrualTime)) {
		return sdk.ZeroInt()
	}
	if rewardPeriod.End.Before(blockTime) {
		return sdk.NewInt(int64(math.RoundToEven(
			rewardPeriod.End.Sub(previousAccrualTime).Seconds(),
		)))
	}
	return sdk.NewInt(int64(math.RoundToEven(
		blockTime.Sub(previousAccrualTime).Seconds(),
	)))
}

// SimulateHardSynchronization calculates a user's outstanding hard rewards by simulating reward synchronization
func (k Keeper) SimulateHardSynchronization(ctx sdk.Context, claim types.HardLiquidityProviderClaim) types.HardLiquidityProviderClaim {
	// 1. Simulate Hard supply-side rewards
	for _, ri := range claim.SupplyRewardIndexes {
		supplyFactor, found := k.GetHardSupplyRewardFactor(ctx, ri.CollateralType)
		if !found {
			continue
		}

		supplyIndex, hasSupplyRewardIndex := claim.HasSupplyRewardIndex(ri.CollateralType)
		if !hasSupplyRewardIndex {
			continue
		}
		claim.SupplyRewardIndexes[supplyIndex].RewardFactor = supplyFactor

		rewardsAccumulatedFactor := supplyFactor.Sub(ri.RewardFactor)
		if rewardsAccumulatedFactor.IsZero() {
			continue
		}

		deposit, found := k.hardKeeper.GetDeposit(ctx, claim.GetOwner())
		if !found {
			continue
		}

		var newRewardsAmount sdk.Int
		if deposit.Amount.AmountOf(ri.CollateralType).GT(sdk.ZeroInt()) {
			newRewardsAmount = rewardsAccumulatedFactor.Mul(deposit.Amount.AmountOf(ri.CollateralType).ToDec()).RoundInt()
			if newRewardsAmount.IsZero() || newRewardsAmount.IsNegative() {
				continue
			}
		}
		newRewardsCoin := sdk.NewCoin(types.HardLiquidityRewardDenom, newRewardsAmount)
		claim.Reward = claim.Reward.Add(newRewardsCoin)
	}

	// 2. Simulate Hard borrow-side rewards
	for _, ri := range claim.BorrowRewardIndexes {
		borrowFactor, found := k.GetHardBorrowRewardFactor(ctx, ri.CollateralType)
		if !found {
			continue
		}

		borrowIndex, hasBorrowRewardIndex := claim.HasBorrowRewardIndex(ri.CollateralType)
		if !hasBorrowRewardIndex {
			continue
		}
		claim.BorrowRewardIndexes[borrowIndex].RewardFactor = borrowFactor

		rewardsAccumulatedFactor := borrowFactor.Sub(ri.RewardFactor)
		if rewardsAccumulatedFactor.IsZero() {
			continue
		}

		borrow, found := k.hardKeeper.GetBorrow(ctx, claim.GetOwner())
		if !found {
			continue
		}

		var newRewardsAmount sdk.Int
		if borrow.Amount.AmountOf(ri.CollateralType).GT(sdk.ZeroInt()) {
			newRewardsAmount = rewardsAccumulatedFactor.Mul(borrow.Amount.AmountOf(ri.CollateralType).ToDec()).RoundInt()
			if newRewardsAmount.IsZero() || newRewardsAmount.IsNegative() {
				continue
			}
		}
		newRewardsCoin := sdk.NewCoin(types.HardLiquidityRewardDenom, newRewardsAmount)
		claim.Reward = claim.Reward.Add(newRewardsCoin)
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

	// TODO: set reasonable max limit on delegation iteration
	maxUInt := ^uint16(0)
	delegations := k.stakingKeeper.GetDelegatorDelegations(ctx, claim.GetOwner(), maxUInt)
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
