package bep3

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker on every block expires outdated atomic swaps and removes closed
// swap from long term storage (default storage time of 1 week)
func BeginBlocker(ctx sdk.Context, k Keeper) {
	if ctx.BlockHeight() == 999999999 { // TODO
		err := k.EnsureModuleAccountPermissions(ctx)
		if err != nil {
			k.Logger(ctx).Error("%v", err)
		}
	}
	k.UpdateTimeBasedSupplyLimits(ctx)
	k.UpdateExpiredAtomicSwaps(ctx)
	k.DeleteClosedAtomicSwapsFromLongtermStorage(ctx)
}
