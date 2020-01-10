package auction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlocker runs at the end of every block.
func EndBlocker(ctx sdk.Context, k Keeper) {
	err := k.CloseExpiredAuctions(ctx)
	if err != nil {
		panic(err)
	}
}
