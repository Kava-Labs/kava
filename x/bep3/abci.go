package bep3

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker on every block expires outdated atomic swaps and removes closed
// swap from long term storage (default storage time of 1 week)
func BeginBlocker(ctx sdk.Context, k Keeper) {
	k.UpdateExpiredAtomicSwaps(ctx)
	k.DeleteClosedAtomicSwapsFromLongtermStorage(ctx)
	k.UpdateAssetSupplies(ctx)
}
