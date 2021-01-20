package hard

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker updates interest rates and attempts liquidations
func BeginBlocker(ctx sdk.Context, k Keeper) {
	k.ApplyInterestRateUpdates(ctx)
	k.AttemptIndexLiquidations(ctx)
	k.SetPreviousBlockTime(ctx, ctx.BlockTime())
}
