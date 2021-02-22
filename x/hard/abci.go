package hard

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker updates interest rates
func BeginBlocker(ctx sdk.Context, k Keeper) {
	k.ApplyInterestRateUpdates(ctx)
}
