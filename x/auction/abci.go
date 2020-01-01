package auction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlocker runs at the end of every block.
func EndBlocker(ctx sdk.Context, k Keeper) {

	var expiredAuctions []uint64
	k.IterateAuctionsByTime(ctx, ctx.BlockTime(), func(id uint64) bool {
		expiredAuctions = append(expiredAuctions, id)
		return false
	})
	// Note: iteration and auction closing are in separate loops as db should not be modified during iteration // TODO is this correct? gov modifies during iteration
	for _, id := range expiredAuctions {
		err := k.CloseAuction(ctx, id)
		if err != nil {
			panic(err)
		}
	}
}
