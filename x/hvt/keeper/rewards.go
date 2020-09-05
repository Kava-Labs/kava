package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingexported "github.com/cosmos/cosmos-sdk/x/staking/exported"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/kava-labs/kava/x/hvt/types"
)

// ApplyDepositRewards iterates over lp and gov deposits and updates the amount of rewards for each depositor
func (k Keeper) ApplyDepositRewards(ctx sdk.Context) {
	previousBlockTime, found := k.GetPreviousBlockTime(ctx)
	if !found {
		previousBlockTime = ctx.BlockTime()
		k.SetPreviousBlockTime(ctx, previousBlockTime)
		return
	}
	timeElapsed := sdk.NewInt(ctx.BlockTime().Unix() - previousBlockTime.Unix())

	params := k.GetParams(ctx)

	for _, lps := range params.LiquidityProviderSchedules {
		if lps.End.Before(ctx.BlockTime()) {
			continue
		}
		totalDeposited := k.GetTotalDeposited(ctx, types.LP, lps.DepositDenom)
		rewardsToDistribute := lps.Reward.Amount.Mul(timeElapsed)

		k.IterateDepositsByTypeAndDenom(ctx, types.LP, lps.DepositDenom, func(dep types.Deposit) (stop bool) {
			rewardsShare := sdk.NewDecFromInt(dep.Amount.Amount).Quo(sdk.NewDecFromInt(totalDeposited))
			if rewardsShare.IsZero() {
				return false
			}
			rewardsEarned := rewardsShare.Mul(sdk.NewDecFromInt(rewardsToDistribute)).RoundInt()
			if rewardsEarned.GT(rewardsToDistribute) {
				rewardsEarned = rewardsToDistribute
			}
			if rewardsEarned.IsZero() {
				return false
			}
			k.AddToClaim(ctx, dep.Depositor, dep.Amount.Denom, dep.Type, sdk.NewCoin(lps.Reward.Denom, rewardsEarned))
			rewardsToDistribute = rewardsToDistribute.Sub(rewardsEarned)
			return false
		})
	}

	for _, gds := range params.GovernanceDistributionSchedules {
		if gds.End.Before(ctx.BlockTime()) {
			continue
		}
		totalDeposited := k.GetTotalDeposited(ctx, types.Gov, gds.DepositDenom)

		k.IterateDepositsByTypeAndDenom(ctx, types.Gov, gds.DepositDenom, func(dep types.Deposit) (stop bool) {
			rewardsToDistribute := gds.Reward.Amount.Mul(timeElapsed)
			rewardsShare := sdk.NewDecFromInt(dep.Amount.Amount).Quo(sdk.NewDecFromInt(totalDeposited))
			if rewardsShare.IsZero() {
				return false
			}
			rewardsEarned := rewardsShare.Mul(sdk.NewDecFromInt(rewardsToDistribute)).RoundInt()
			if rewardsEarned.IsZero() {
				return false
			}
			k.AddToClaim(ctx, dep.Depositor, dep.Amount.Denom, dep.Type, sdk.NewCoin(gds.Reward.Denom, rewardsEarned))
			return false
		})
	}
	k.SetPreviousBlockTime(ctx, ctx.BlockTime())
}

// ShouldDistributeValidatorRewards returns true if enough time has elapsed such that rewards should be distributed to delegators
func (k Keeper) ShouldDistributeValidatorRewards(ctx sdk.Context, denom string) bool {
	previousDistributionTime, found := k.GetPreviousDelegatorDistribution(ctx, denom)
	if !found {
		k.SetPreviousDelegationDistribution(ctx, ctx.BlockTime(), denom)
		return false
	}
	params := k.GetParams(ctx)
	for _, dds := range params.DelegatorDistributionSchedules {
		if denom != dds.DistributionSchedule.DepositDenom {
			continue
		}
		timeElapsed := sdk.NewInt(ctx.BlockTime().Unix() - previousDistributionTime.Unix())
		if timeElapsed.GTE(sdk.NewInt(int64(dds.DistributionFrequency))) {
			return true
		}
	}
	return false
}

// ApplyDelegationRewards iterates over each delegation object in the staking store and applies rewards according to the input delegation distribution schedule
func (k Keeper) ApplyDelegationRewards(ctx sdk.Context, denom string) {
	dds, found := k.GetDelegatorSchedule(ctx, denom)
	if !found {
		return
	}
	bondMacc := k.stakingKeeper.GetBondedPool(ctx)
	bondedCoinAmount := bondMacc.GetCoins().AmountOf(dds.DistributionSchedule.DepositDenom)
	previousDistributionTime, found := k.GetPreviousDelegatorDistribution(ctx, dds.DistributionSchedule.DepositDenom)
	if !found {
		return
	}
	timeElapsed := sdk.NewInt(ctx.BlockTime().Unix() - previousDistributionTime.Unix())
	rewardsToDistribute := dds.DistributionSchedule.Reward.Amount.Mul(timeElapsed)

	// map that has each validator address (sdk.ValAddress) as a key and the coversion factor for going from delegator shares to tokens for delegations to that validator. If a validator has never been slashed, the conversion factor will be 1.0
	sharesToTokens := make(map[string]sdk.Dec)
	k.stakingKeeper.IterateLastValidators(ctx, func(index int64, validator stakingexported.ValidatorI) (stop bool) {
		sharesToTokens[validator.GetOperator().String()] = (validator.GetDelegatorShares()).Quo(sdk.NewDecFromInt(validator.GetTokens()))
		return false
	})

	k.stakingKeeper.IterateAllDelegations(ctx, func(delegation stakingtypes.Delegation) (stop bool) {
		conversionFactor, ok := sharesToTokens[delegation.ValidatorAddress.String()]
		if ok {
			delegationTokens := conversionFactor.Mul(delegation.Shares)
			delegationShare := delegationTokens.Quo(sdk.NewDecFromInt(bondedCoinAmount))
			rewardsEarned := delegationShare.Mul(sdk.NewDecFromInt(rewardsToDistribute)).RoundInt()
			if rewardsEarned.GT(rewardsToDistribute) {
				rewardsEarned = rewardsToDistribute
			}
			if rewardsEarned.IsZero() {
				return false
			}
			k.AddToClaim(ctx, delegation.DelegatorAddress, dds.DistributionSchedule.DepositDenom, types.Stake, sdk.NewCoin(dds.DistributionSchedule.Reward.Denom, rewardsEarned))
			rewardsToDistribute = rewardsToDistribute.Sub(rewardsEarned)
		}
		return false
	})

}
