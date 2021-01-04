package hard

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker applies rewards to liquidity providers and delegators according to params
func BeginBlocker(ctx sdk.Context, k Keeper) {
	k.ApplyDepositRewards(ctx)
	if k.ShouldDistributeValidatorRewards(ctx, k.BondDenom(ctx)) {
		k.ApplyDelegationRewards(ctx, k.BondDenom(ctx))
		k.SetPreviousDelegationDistribution(ctx, ctx.BlockTime(), k.BondDenom(ctx))
	}
	k.ApplyInterestRateUpdates(ctx)
	k.AttemptIndexLiquidations(ctx)
	k.SetPreviousBlockTime(ctx, ctx.BlockTime())
}
