package bep3

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker runs at the start of every block
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k Keeper) {
	//
}

// EndBlocker runs at the end of every block.
func EndBlocker(ctx sdk.Context, k Keeper) {
	// err := k.CloseExpiredAuctions(ctx)
	// if err != nil {
	// 	panic(err)
	// }
}
