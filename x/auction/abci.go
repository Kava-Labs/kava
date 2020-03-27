package auction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker runs at the start of every block.
func BeginBlocker(ctx sdk.Context, k Keeper) {
	err := k.CloseExpiredAuctions(ctx)
	if err != nil {
		panic(err)
	}
}
