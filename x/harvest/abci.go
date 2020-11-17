package harvest

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

	coins, found := k.GetBorrowedCoins(ctx)
	if found {
		for _, coin := range coins {
			// TODO: consider implications of panic in begin blocker
			_ = k.AccrueInterest(ctx, coin.Denom)
		}
	}

	k.ApplyInterestRateUpdates(ctx)
	k.SetPreviousBlockTime(ctx, ctx.BlockTime())
}
