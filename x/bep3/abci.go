package bep3

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// BeginBlocker runs at the start of every block
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k Keeper) {
	err := k.UpdateExpiredAtomicSwaps(ctx)
	if err != nil {
		panic(err)
	}
}
