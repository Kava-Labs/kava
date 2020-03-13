package bep3

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker runs at the start of every block
func BeginBlocker(ctx sdk.Context, k Keeper) {
	err := k.UpdateExpiredAtomicSwaps(ctx)
	if err != nil {
		panic(err)
	}

	err := k.DeleteClosedAtomicSwapsFromLongtermStorage(ctx)
	if err != nil {
		panic(err)
	}
}
