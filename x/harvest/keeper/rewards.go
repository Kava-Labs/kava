package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingexported "github.com/cosmos/cosmos-sdk/x/staking/exported"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/kava-labs/kava/x/harvest/types"
)

// ApplyDepositRewards iterates over lp and gov deposits and updates the amount of rewards for each depositor
func (k Keeper) ApplyDepositRewards(ctx sdk.Context) {
	previousBlockTime, found := k.GetPreviousBlockTime(ctx)
	if !found {
		previousBlockTime = ctx.BlockTime()
		k.SetPreviousBlockTime(ctx, previousBlockTime)
		return
	}
	params := k.GetParams(ctx)
	if !params.Active {
		return
	}
	timeElapsed := sdk.NewInt(ctx.BlockTime().Unix() - previousBlockTime.Unix())

	for _, lps := range params.LiquidityProviderSchedules {
		if !lps.Active {
			continue
		}
		if lps.End.Before(ctx.BlockTime()) {
			continue
		}
		totalDeposited := k.GetTotalDeposited(ctx, types.LP, lps.DepositDenom)
		if totalDeposited.IsZero() {
			continue
		}
		rewardsToDistribute := lps.RewardsPerSecond.Amount.Mul(timeElapsed)
		if rewardsToDistribute.IsZero() {
			continue
		}
		rewardsDistributed := sdk.ZeroInt()
		k.IterateDepositsByTypeAndDenom(ctx, types.LP, lps.DepositDenom, func(dep types.Deposit) (stop bool) {
			rewardsShare := sdk.NewDecFromInt(dep.Amount.Amount).Quo(sdk.NewDecFromInt(totalDeposited))
			if rewardsShare.IsZero() {
				return false
			}
			rewardsEarned := rewardsShare.Mul(sdk.NewDecFromInt(rewardsToDistribute)).RoundInt()
			if rewardsEarned.IsZero() {
				return false
			}
			k.AddToClaim(ctx, dep.Depositor, dep.Amount.Denom, dep.Type, sdk.NewCoin(lps.RewardsPerSecond.Denom, rewardsEarned))
			rewardsDistributed = rewardsDistributed.Add(rewardsEarned)
			return false
		})
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeHarvestLPDistribution,
				sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
				sdk.NewAttribute(types.AttributeKeyRewardsDistribution, rewardsDistributed.String()),
				sdk.NewAttribute(types.AttributeKeyDepositDenom, lps.DepositDenom),
			),
		)
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
	if !params.Active {
		return false
	}
	for _, dds := range params.DelegatorDistributionSchedules {
		if denom != dds.DistributionSchedule.DepositDenom {
			continue
		}
		timeElapsed := sdk.NewInt(ctx.BlockTime().Unix() - previousDistributionTime.Unix())
		if timeElapsed.GTE(sdk.NewInt(int64(dds.DistributionFrequency.Seconds()))) {
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
	if !dds.DistributionSchedule.Active {
		return
	}
	bondMacc := k.stakingKeeper.GetBondedPool(ctx)
	bondedCoinAmount := bondMacc.GetCoins().AmountOf(dds.DistributionSchedule.DepositDenom)
	if bondedCoinAmount.IsZero() {
		return
	}
	previousDistributionTime, found := k.GetPreviousDelegatorDistribution(ctx, dds.DistributionSchedule.DepositDenom)
	if !found {
		return
	}
	timeElapsed := sdk.NewInt(ctx.BlockTime().Unix() - previousDistributionTime.Unix())
	rewardsToDistribute := dds.DistributionSchedule.RewardsPerSecond.Amount.Mul(timeElapsed)

	// create a map that has each validator address (sdk.ValAddress) as a key and the coversion factor for going from delegator shares to tokens for delegations to that validator.
	// If a validator has never been slashed, the conversion factor will be 1.0, if they have been, it will be < 1.0
	sharesToTokens := make(map[string]sdk.Dec)
	k.stakingKeeper.IterateValidators(ctx, func(index int64, validator stakingexported.ValidatorI) (stop bool) {
		if validator.GetTokens().IsZero() {
			return false
		}
		// don't include a validator if it's unbonded - ie delegators don't accumulate rewards when delegated to an unbonded validator
		if validator.GetStatus() == sdk.Unbonded {
			return false
		}
		sharesToTokens[validator.GetOperator().String()] = (validator.GetDelegatorShares()).Quo(sdk.NewDecFromInt(validator.GetTokens()))
		return false
	})

	rewardsDistributed := sdk.ZeroInt()

	k.stakingKeeper.IterateAllDelegations(ctx, func(delegation stakingtypes.Delegation) (stop bool) {
		conversionFactor, ok := sharesToTokens[delegation.ValidatorAddress.String()]
		if ok {
			delegationTokens := conversionFactor.Mul(delegation.Shares)
			delegationShare := delegationTokens.Quo(sdk.NewDecFromInt(bondedCoinAmount))
			rewardsEarned := delegationShare.Mul(sdk.NewDecFromInt(rewardsToDistribute)).RoundInt()
			if rewardsEarned.IsZero() {
				return false
			}
			k.AddToClaim(
				ctx, delegation.DelegatorAddress, dds.DistributionSchedule.DepositDenom,
				types.Stake, sdk.NewCoin(dds.DistributionSchedule.RewardsPerSecond.Denom, rewardsEarned))
			rewardsDistributed = rewardsDistributed.Add(rewardsEarned)
		}
		return false
	})

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeHarvestDelegatorDistribution,
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
			sdk.NewAttribute(types.AttributeKeyRewardsDistribution, rewardsDistributed.String()),
			sdk.NewAttribute(types.AttributeKeyDepositDenom, denom),
		),
	)

}

// AddToClaim adds the input amount to an existing claim or creates a new one
func (k Keeper) AddToClaim(ctx sdk.Context, owner sdk.AccAddress, depositDenom string, depositType types.DepositType, amountToAdd sdk.Coin) {
	claim, found := k.GetClaim(ctx, owner, depositDenom, depositType)
	if !found {
		claim = types.NewClaim(owner, depositDenom, amountToAdd, depositType)
	} else {
		claim.Amount = claim.Amount.Add(amountToAdd)
	}
	k.SetClaim(ctx, claim)
}
