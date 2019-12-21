package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/auction/keeper"
	"github.com/kava-labs/kava/x/auction/types"
)

// func TestKeeper_ForwardAuction(t *testing.T) {
// 	// Setup
// 	_, addrs := app.GeneratePrivKeyAddressPairs(2)
// 	seller := addrs[0]
// 	buyer := addrs[1]

// 	tApp := app.NewTestApp()
// 	tApp.InitializeFromGenesisStates(
// 		app.NewAuthGenState(addrs, []sdk.Coins{cs(c("token1", 100), c("token2", 100)), cs(c("token1", 100), c("token2", 100))}),
// 	)

// 	ctx := tApp.NewContext(false, abci.Header{})
// 	keeper := tApp.GetAuctionKeeper()

// 	// Create an auction (lot: 20 t1, initialBid: 0 t2)
// 	auctionID, err := keeper.StartForwardAuction(ctx, seller, c("token1", 20), c("token2", 0)) // lot, initialBid
// 	require.NoError(t, err)
// 	// Check seller's coins have decreased
// 	tApp.CheckBalance(t, ctx, seller, cs(c("token1", 80), c("token2", 100)))

// 	// PlaceBid (bid: 10 t2, lot: same as starting)
// 	require.NoError(t, keeper.PlaceBid(ctx, 0, buyer, c("token2", 10), c("token1", 20))) // bid, lot
// 	// Check buyer's coins have decreased
// 	tApp.CheckBalance(t, ctx, buyer, cs(c("token1", 100), c("token2", 90)))
// 	// Check seller's coins have increased
// 	tApp.CheckBalance(t, ctx, seller, cs(c("token1", 80), c("token2", 110)))

// 	// Close auction at just after auction expiry
// 	ctx = ctx.WithBlockHeight(int64(types.DefaultMaxBidDuration))
// 	require.NoError(t, keeper.CloseAuction(ctx, auctionID))
// 	// Check buyer's coins increased
// 	tApp.CheckBalance(t, ctx, buyer, cs(c("token1", 120), c("token2", 90)))
// }

// func TestKeeper_ReverseAuction(t *testing.T) {
// 	// Setup
// 	_, addrs := app.GeneratePrivKeyAddressPairs(2)
// 	seller := addrs[0]
// 	buyer := addrs[1]

// 	tApp := app.NewTestApp()
// 	tApp.InitializeFromGenesisStates(
// 		app.NewAuthGenState(addrs, []sdk.Coins{cs(c("token1", 100), c("token2", 100)), cs(c("token1", 100), c("token2", 100))}),
// 	)

// 	ctx := tApp.NewContext(false, abci.Header{})
// 	keeper := tApp.GetAuctionKeeper()

// 	// Start auction
// 	auctionID, err := keeper.StartReverseAuction(ctx, buyer, c("token1", 20), c("token2", 99)) // buyer, bid, initialLot
// 	require.NoError(t, err)
// 	// Check buyer's coins have decreased
// 	tApp.CheckBalance(t, ctx, buyer, cs(c("token1", 100), c("token2", 1)))

// 	// Place a bid
// 	require.NoError(t, keeper.PlaceBid(ctx, 0, seller, c("token1", 20), c("token2", 10))) // bid, lot
// 	// Check seller's coins have decreased
// 	tApp.CheckBalance(t, ctx, seller, cs(c("token1", 80), c("token2", 100)))
// 	// Check buyer's coins have increased
// 	tApp.CheckBalance(t, ctx, buyer, cs(c("token1", 120), c("token2", 90)))

// 	// Close auction at just after auction expiry
// 	ctx = ctx.WithBlockHeight(int64(types.DefaultMaxBidDuration))
// 	require.NoError(t, keeper.CloseAuction(ctx, auctionID))
// 	// Check seller's coins increased
// 	tApp.CheckBalance(t, ctx, seller, cs(c("token1", 80), c("token2", 110)))
// }

// func TestKeeper_ForwardReverseAuction(t *testing.T) {
// 	// Setup
// 	_, addrs := app.GeneratePrivKeyAddressPairs(3)
// 	seller := addrs[0]
// 	buyer := addrs[1]
// 	recipient := addrs[2]

// 	tApp := app.NewTestApp()
// 	tApp.InitializeFromGenesisStates(
// 		app.NewAuthGenState(addrs, []sdk.Coins{cs(c("token1", 100), c("token2", 100)), cs(c("token1", 100), c("token2", 100)), cs(c("token1", 100), c("token2", 100))}),
// 	)

// 	ctx := tApp.NewContext(false, abci.Header{})
// 	keeper := tApp.GetAuctionKeeper()

// 	// Start auction
// 	auctionID, err := keeper.StartForwardReverseAuction(ctx, seller, c("token1", 20), c("token2", 50), recipient) // seller, lot, maxBid, otherPerson
// 	require.NoError(t, err)
// 	// Check seller's coins have decreased
// 	tApp.CheckBalance(t, ctx, seller, cs(c("token1", 80), c("token2", 100)))

// 	// Place a bid
// 	require.NoError(t, keeper.PlaceBid(ctx, 0, buyer, c("token2", 50), c("token1", 15))) // bid, lot
// 	// Check bidder's coins have decreased
// 	tApp.CheckBalance(t, ctx, buyer, cs(c("token1", 100), c("token2", 50)))
// 	// Check seller's coins have increased
// 	tApp.CheckBalance(t, ctx, seller, cs(c("token1", 80), c("token2", 150)))
// 	// Check "recipient" has received coins
// 	tApp.CheckBalance(t, ctx, recipient, cs(c("token1", 105), c("token2", 100)))

// 	// Close auction at just after auction expiry
// 	ctx = ctx.WithBlockHeight(int64(types.DefaultMaxBidDuration))
// 	require.NoError(t, keeper.CloseAuction(ctx, auctionID))
// 	// Check buyer's coins increased
// 	tApp.CheckBalance(t, ctx, buyer, cs(c("token1", 115), c("token2", 50)))
// }

func SetGetDeleteAuction(t *testing.T) {
	// setup keeper, create auction
	tApp := app.NewTestApp()
	keeper := tApp.GetAuctionKeeper()
	ctx := tApp.NewContext(true, abci.Header{})
	someTime := time.Date(43, time.January, 1, 0, 0, 0, 0, time.UTC) // need to specify UTC as tz info is lost on unmarshal
	auction := types.NewForwardAuction("some_module", c("usdx", 100), "kava", someTime)
	id := types.ID(5)
	auction.SetID(id)

	// write and read from store
	keeper.SetAuction(ctx, &auction)
	readAuction, found := keeper.GetAuction(ctx, id)

	// check before and after match
	require.True(t, found)
	require.Equal(t, &auction, readAuction)
	// check auction is in queue
	// iter := keeper.GetQueueIterator(ctx, 100000)
	// require.Equal(t, 1, len(convertIteratorToSlice(keeper, iter)))
	// iter.Close()

	// delete auction
	keeper.DeleteAuction(ctx, id)

	// check auction does not exist
	_, found = keeper.GetAuction(ctx, id)
	require.False(t, found)
	// check auction not in queue
	// iter = keeper.GetQueueIterator(ctx, 100000)
	// require.Equal(t, 0, len(convertIteratorToSlice(keeper, iter)))
	// iter.Close()

}

func TestIncrementNextAuctionID(t *testing.T) {
	// setup keeper
	tApp := app.NewTestApp()
	keeper := tApp.GetAuctionKeeper()
	ctx := tApp.NewContext(true, abci.Header{})

	// store id
	id := types.ID(123456)
	keeper.SetNextAuctionID(ctx, id)

	require.NoError(t, keeper.IncrementNextAuctionID(ctx))

	// check id was incremented
	readID, err := keeper.GetNextAuctionID(ctx)
	require.NoError(t, err)
	require.Equal(t, id+1, readID)

}

// func TestIterateAuctions(t *testing.T) {
// 	// setup keeper
// 	tApp := app.NewTestApp()
// 	keeper := tApp.GetAuctionKeeper()
// 	ctx := tApp.NewContext(true, abci.Header{})

// 	auctions := []types.Auction{
// 		&types.ForwardAuction{},
// 	}
// 	for _, a := range auctions {
// 		keeper.SetAuction(ctx, a)
// 	}

// 	var readAuctions []types.Auction
// 	keeper.IterateAuctions(ctx, func(a types.Auction) bool {
// 		readAuctions = append(readAuctions, a)
// 		return false
// 	})

// 	require.Equal(t, auctions, readAuctions)
// }

func TestIterateAuctionsByTime(t *testing.T) {
	// setup keeper
	tApp := app.NewTestApp()
	keeper := tApp.GetAuctionKeeper()
	ctx := tApp.NewContext(true, abci.Header{})

	// create a list of times
	queue := []struct {
		endTime   time.Time
		auctionID types.ID
	}{
		{time.Date(84, time.January, 1, 0, 0, 0, 0, time.UTC), 34345345},
		{time.Date(98, time.January, 2, 0, 0, 0, 0, time.UTC), 5},
		{time.Date(98, time.January, 2, 13, 5, 0, 0, time.UTC), 6},
		{time.Date(98, time.January, 2, 16, 0, 0, 0, time.UTC), 1},
		{time.Date(98, time.January, 2, 16, 0, 0, 0, time.UTC), 3},
		{time.Date(98, time.January, 2, 16, 0, 0, 0, time.UTC), 4},
		{time.Date(98, time.January, 2, 16, 0, 0, 1, time.UTC), 0}, // TODO tidy up redundant entries
	}
	cutoffTime := time.Date(98, time.January, 2, 16, 0, 0, 0, time.UTC)

	var expectedQueue []types.ID
	for _, i := range queue {
		if i.endTime.After(cutoffTime) { // only append items where endTime ≤ cutoffTime
			break
		}
		expectedQueue = append(expectedQueue, i.auctionID)
	}

	// write and read queue
	for _, v := range queue {
		keeper.InsertIntoQueue(ctx, v.endTime, v.auctionID)
	}
	var readQueue []types.ID
	keeper.IterateAuctionsByTime(ctx, cutoffTime, func(id types.ID) bool {
		readQueue = append(readQueue, id)
		return false
	})

	require.Equal(t, expectedQueue, readQueue)
}

func convertIteratorToSlice(keeper keeper.Keeper, iterator sdk.Iterator) []types.ID {
	var queue []types.ID
	for ; iterator.Valid(); iterator.Next() {
		var auctionID types.ID
		types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &auctionID)
		queue = append(queue, auctionID)
	}
	return queue
}

func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }
