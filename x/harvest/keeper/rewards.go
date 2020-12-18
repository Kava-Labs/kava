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
		if lps.Start.After(ctx.BlockTime()) {
			continue
		}
		totalDeposited := k.GetTotalDeposited(ctx, lps.DepositDenom)
		if totalDeposited.IsZero() {
			continue
		}
		rewardsToDistribute := lps.RewardsPerSecond.Amount.Mul(timeElapsed)
		if rewardsToDistribute.IsZero() {
			continue
		}
		rewardsDistributed := sdk.ZeroInt()
		k.IterateDeposits(ctx, func(dep types.Deposit) (stop bool) {
			rewardsShare := sdk.NewDecFromInt(dep.Amount.AmountOf(lps.DepositDenom)).Quo(sdk.NewDecFromInt(totalDeposited))
			if rewardsShare.IsZero() {
				return false
			}
			rewardsEarned := rewardsShare.Mul(sdk.NewDecFromInt(rewardsToDistribute)).RoundInt()
			if rewardsEarned.IsZero() {
				return false
			}
			k.AddToClaim(ctx, dep.Depositor, lps.DepositDenom, types.LP, sdk.NewCoin(lps.RewardsPerSecond.Denom, rewardsEarned))
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
		if dds.DistributionSchedule.End.Before(ctx.BlockTime()) {
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
	if dds.DistributionSchedule.Start.After(ctx.BlockTime()) {
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
		// don't include a validator if it's unbonded or unbonding- ie delegators don't accumulate rewards when delegated to an unbonded/slashed validator
		if validator.GetStatus() != sdk.Bonded {
			return false
		}
		sharesToTokens[validator.GetOperator().String()] = sdk.NewDecFromInt(validator.GetTokens()).Quo(validator.GetDelegatorShares())
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
func (k Keeper) AddToClaim(ctx sdk.Context, owner sdk.AccAddress, depositDenom string, claimType types.ClaimType, amountToAdd sdk.Coin) {
	claim, found := k.GetClaim(ctx, owner, depositDenom, claimType)
	if !found {
		claim = types.NewClaim(owner, depositDenom, amountToAdd, claimType)
	} else {
		claim.Amount = claim.Amount.Add(amountToAdd)
	}
	k.SetClaim(ctx, claim)
}
