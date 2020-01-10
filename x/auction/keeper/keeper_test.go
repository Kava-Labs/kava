package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/auction/types"
)

func SetGetDeleteAuction(t *testing.T) {
	// setup keeper, create auction
	tApp := app.NewTestApp()
	keeper := tApp.GetAuctionKeeper()
	ctx := tApp.NewContext(true, abci.Header{})
	someTime := time.Date(43, time.January, 1, 0, 0, 0, 0, time.UTC) // need to specify UTC as tz info is lost on unmarshal
	var id uint64 = 5
	auction := types.NewSurplusAuction("some_module", c("usdx", 100), "kava", someTime).WithID(id)

	// write and read from store
	keeper.SetAuction(ctx, auction)
	readAuction, found := keeper.GetAuction(ctx, id)

	// check before and after match
	require.True(t, found)
	require.Equal(t, auction, readAuction)
	// check auction is in the index
	keeper.IterateAuctionsByTime(ctx, auction.GetEndTime(), func(readID uint64) bool {
		require.Equal(t, auction.GetID(), readID)
		return false
	})

	// delete auction
	keeper.DeleteAuction(ctx, id)

	// check auction does not exist
	_, found = keeper.GetAuction(ctx, id)
	require.False(t, found)
	// check auction not in index
	keeper.IterateAuctionsByTime(ctx, time.Unix(999999999, 0), func(readID uint64) bool {
		require.Fail(t, "index should be empty", " found auction ID '%s", readID)
		return false
	})
}

func TestIncrementNextAuctionID(t *testing.T) {
	// setup keeper
	tApp := app.NewTestApp()
	keeper := tApp.GetAuctionKeeper()
	ctx := tApp.NewContext(true, abci.Header{})

	// store id
	var id uint64 = 123456
	keeper.SetNextAuctionID(ctx, id)

	require.NoError(t, keeper.IncrementNextAuctionID(ctx))

	// check id was incremented
	readID, err := keeper.GetNextAuctionID(ctx)
	require.NoError(t, err)
	require.Equal(t, id+1, readID)

}

func TestIterateAuctions(t *testing.T) {
	// setup
	tApp := app.NewTestApp()
	tApp.InitializeFromGenesisStates()
	keeper := tApp.GetAuctionKeeper()
	ctx := tApp.NewContext(true, abci.Header{})

	auctions := []types.Auction{
		types.NewSurplusAuction("sellerMod", c("denom", 12345678), "anotherdenom", time.Date(1998, time.January, 1, 0, 0, 0, 0, time.UTC)).WithID(0),
		types.NewDebtAuction("buyerMod", c("denom", 12345678), c("anotherdenom", 12345678), time.Date(1998, time.January, 1, 0, 0, 0, 0, time.UTC)).WithID(1),
		types.NewCollateralAuction("sellerMod", c("denom", 12345678), time.Date(1998, time.January, 1, 0, 0, 0, 0, time.UTC), c("anotherdenom", 12345678), types.WeightedAddresses{}).WithID(2),
	}
	for _, a := range auctions {
		keeper.SetAuction(ctx, a)
	}

	// run
	var readAuctions []types.Auction
	keeper.IterateAuctions(ctx, func(a types.Auction) bool {
		readAuctions = append(readAuctions, a)
		return false
	})

	// check
	require.Equal(t, auctions, readAuctions)
}

func TestIterateAuctionsByTime(t *testing.T) {
	// setup keeper
	tApp := app.NewTestApp()
	keeper := tApp.GetAuctionKeeper()
	ctx := tApp.NewContext(true, abci.Header{})

	// setup byTime index
	byTimeIndex := []struct {
		endTime   time.Time
		auctionID uint64
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
		keeper.InsertIntoByTimeIndex(ctx, v.endTime, v.auctionID)
	}

	// read out values from index up to a cutoff time and check they are as expected
	cutoffTime := time.Date(1998, time.January, 1, 12, 0, 0, 0, time.UTC)
	var expectedIndex []uint64
	for _, v := range byTimeIndex {
		if v.endTime.Before(cutoffTime) || v.endTime.Equal(cutoffTime) { // endTime ≤ cutoffTime
			expectedIndex = append(expectedIndex, v.auctionID)
		}

	}
	var readIndex []uint64
	keeper.IterateAuctionsByTime(ctx, cutoffTime, func(id uint64) bool {
		readIndex = append(readIndex, id)
		return false
	})

	require.Equal(t, expectedIndex, readIndex)
}
