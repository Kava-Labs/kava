package auction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlocker runs at the end of every block.
func EndBlocker(ctx sdk.Context, k Keeper) {

	// get an iterator of expired auctions
	expiredAuctions := k.GetQueueIterator(ctx, EndTime(ctx.BlockHeight()))
	defer expiredAuctions.Close()

	// loop through and close them - distribute funds, delete from store (and queue)
	for ; expiredAuctions.Valid(); expiredAuctions.Next() {

		auctionID := k.DecodeAuctionID(ctx, expiredAuctions.Value())
		err := k.CloseAuction(ctx, auctionID)
		if err != nil {
			panic(err) // TODO how should errors be handled here?
		}
	}
}
