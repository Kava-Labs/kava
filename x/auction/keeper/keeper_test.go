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

func SetGetDeleteAuction(t *testing.T) {
	// setup keeper, create auction
	tApp := app.NewTestApp()
	keeper := tApp.GetAuctionKeeper()
	ctx := tApp.NewContext(true, abci.Header{})
	someTime := time.Date(43, time.January, 1, 0, 0, 0, 0, time.UTC) // need to specify UTC as tz info is lost on unmarshal
	id := types.ID(5)
	auction := types.NewForwardAuction("some_module", c("usdx", 100), "kava", someTime).WithID(id)

	// write and read from store
	keeper.SetAuction(ctx, auction)
	readAuction, found := keeper.GetAuction(ctx, id)

	// check before and after match
	require.True(t, found)
	require.Equal(t, auction, readAuction)
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

	// setup byTime index
	byTimeIndex := []struct {
		endTime   time.Time
		auctionID types.ID
	}{
		{time.Date(0, time.January, 1, 0, 0, 0, 0, time.UTC), 9999},            // distant past
		{time.Date(1998, time.January, 1, 11, 59, 59, 999999999, time.UTC), 1}, // just before cutoff
		{time.Date(1998, time.January, 1, 11, 59, 59, 999999999, time.UTC), 2}, //
		{time.Date(1998, time.January, 1, 12, 0, 0, 0, time.UTC), 3},           // equal to cutoff
		{time.Date(1998, time.January, 1, 12, 0, 0, 0, time.UTC), 4},           //
		{time.Date(1998, time.January, 1, 12, 0, 0, 1, time.UTC), 5},           // just after cutoff
		{time.Date(1998, time.January, 1, 12, 0, 0, 1, time.UTC), 6},           //
		{time.Date(9999, time.January, 1, 0, 0, 0, 0, time.UTC), 0},            // distant future
	}
	for _, v := range byTimeIndex {
		keeper.InsertIntoQueue(ctx, v.endTime, v.auctionID)
	}

	// read out values from index up to a cutoff time and check they are as expected
	cutoffTime := time.Date(1998, time.January, 1, 12, 0, 0, 0, time.UTC)
	var expectedIndex []types.ID
	for _, v := range byTimeIndex {
		if v.endTime.Before(cutoffTime) || v.endTime.Equal(cutoffTime) { // endTime ≤ cutoffTime
			expectedIndex = append(expectedIndex, v.auctionID)
		}

	}
	var readIndex []types.ID
	keeper.IterateAuctionsByTime(ctx, cutoffTime, func(id types.ID) bool {
		readIndex = append(readIndex, id)
		return false
	})

	require.Equal(t, expectedIndex, readIndex)
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
