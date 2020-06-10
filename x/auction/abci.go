package auction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker closes all expired auctions at the end of each block. It panics if
// there's an error other than ErrAuctionNotFound.
func BeginBlocker(ctx sdk.Context, k Keeper) {
	err := k.CloseExpiredAuctions(ctx)
	if err != nil {
		panic(err)
	}
}
