package bep3

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker on every block expires outdated atomic swaps and removes closed
// swap from long term storage (default storage time of 1 week)
func BeginBlocker(ctx sdk.Context, k Keeper) {
	if ctx.BlockTime().After(ModulePermissionsUpgradeTime) {
		err := k.EnsureModuleAccountPermissions(ctx)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("couldn't update module account permissions: %v", err))
		}
	}
	k.UpdateTimeBasedSupplyLimits(ctx)
	k.UpdateExpiredAtomicSwaps(ctx)
	k.DeleteClosedAtomicSwapsFromLongtermStorage(ctx)
}
