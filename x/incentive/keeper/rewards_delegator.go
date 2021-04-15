package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/types"
)

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
		k.SynchronizeHardDelegatorRewards(ctx, delegator, nil, false)
		claim, _ = k.GetHardLiquidityProviderClaim(ctx, delegator)
	}

	claim.DelegatorRewardIndexes = types.RewardIndexes{delegatorRewardIndexes}
	k.SetHardLiquidityProviderClaim(ctx, claim)
}

// SynchronizeHardDelegatorRewards updates the claim object by adding any accumulated rewards, and setting the reward indexes to the global values.
// valAddr and shouldIncludeValidator are used to ignore or include delegations to a particular validator when summing up the total delegation.
// Normally only delegations to Bonded validators are included in the total. This is needed as staking hooks are sometimes called on the wrong side of a validator's state update (from this module's perspective).
func (k Keeper) SynchronizeHardDelegatorRewards(ctx sdk.Context, delegator sdk.AccAddress, valAddr sdk.ValAddress, shouldIncludeValidator bool) {
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
	if rewardsAccumulatedFactor.IsNegative() {
		panic(fmt.Sprintf("reward accumulation factor cannot be negative: %s", rewardsAccumulatedFactor))
	}
	claim.DelegatorRewardIndexes[delegatorIndex].RewardFactor = delagatorFactor

	totalDelegated := sdk.ZeroDec()

	delegations := k.stakingKeeper.GetDelegatorDelegations(ctx, delegator, 200)
	for _, delegation := range delegations {
		validator, found := k.stakingKeeper.GetValidator(ctx, delegation.GetValidatorAddr())
		if !found {
			continue
		}

		if valAddr == nil {
			// Delegators don't accumulate rewards if their validator is unbonded
			if validator.GetStatus() != sdk.Bonded {
				continue
			}
		} else {
			if !shouldIncludeValidator && validator.OperatorAddress.Equals(valAddr) {
				// ignore tokens delegated to the validator
				continue
			}
		}

		if validator.GetTokens().IsZero() {
			continue
		}

		delegatedTokens := validator.TokensFromShares(delegation.GetShares())
		if delegatedTokens.IsNegative() {
			continue
		}
		totalDelegated = totalDelegated.Add(delegatedTokens)
	}
	rewardsEarned := rewardsAccumulatedFactor.Mul(totalDelegated).RoundInt()

	// Add rewards to delegator's hard claim
	newRewardsCoin := sdk.NewCoin(types.HardLiquidityRewardDenom, rewardsEarned)
	claim.Reward = claim.Reward.Add(newRewardsCoin)
	k.SetHardLiquidityProviderClaim(ctx, claim)
}
