package bep3

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/bep3/keeper"
)

// BeginBlocker on every block expires outdated atomic swaps and removes closed
// swap from long term storage (default storage time of 1 week)
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	k.UpdateTimeBasedSupplyLimits(ctx)
	k.UpdateExpiredAtomicSwaps(ctx)
	k.DeleteClosedAtomicSwapsFromLongtermStorage(ctx)
}
