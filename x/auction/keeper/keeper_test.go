package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/auction/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestKeeper_SetGetDeleteAuction(t *testing.T) {
	// setup keeper, create auction
	mapp, keeper, addresses, _ := setUpMockApp()
	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header}) // Without this it panics about "invalid memory address or nil pointer dereference"
	ctx := mapp.BaseApp.NewContext(false, header)
	auction, _ := types.NewForwardAuction(addresses[0], sdk.NewInt64Coin("usdx", 100), sdk.NewInt64Coin("kava", 0), types.EndTime(1000))
	id := types.ID(5)
	auction.SetID(id)

	// write and read from store
	keeper.SetAuction(ctx, &auction)
	readAuction, found := keeper.GetAuction(ctx, id)

	// check before and after match
	require.True(t, found)
	require.Equal(t, &auction, readAuction)
	t.Log(auction)
	t.Log(readAuction.GetID())
	// check auction is in queue
	iter := keeper.GetQueueIterator(ctx, 100000)
	require.Equal(t, 1, len(convertIteratorToSlice(keeper, iter)))
	iter.Close()

	// delete auction
	keeper.deleteAuction(ctx, id)

	// check auction does not exist
	_, found = keeper.GetAuction(ctx, id)
	require.False(t, found)
	// check auction not in queue
	iter = keeper.GetQueueIterator(ctx, 100000)
	require.Equal(t, 0, len(convertIteratorToSlice(keeper, iter)))
	iter.Close()

}

// TODO convert to table driven test with more test cases
func TestKeeper_ExpiredAuctionQueue(t *testing.T) {
	// setup keeper
	mapp, keeper, _, _ := setUpMockApp()
	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := mapp.BaseApp.NewContext(false, header)
	// create an example queue
	type queue []struct {
		endTime   types.EndTime
		auctionID types.ID
	}
	q := queue{{1000, 0}, {1300, 2}, {5200, 1}}

	// write and read queue
	for _, v := range q {
		keeper.insertIntoQueue(ctx, v.endTime, v.auctionID)
	}
	iter := keeper.GetQueueIterator(ctx, 1000)

	// check before and after match
	i := 0
	for ; iter.Valid(); iter.Next() {
		var auctionID types.ID
		keeper.cdc.MustUnmarshalBinaryLengthPrefixed(iter.Value(), &auctionID)
		require.Equal(t, q[i].auctionID, auctionID)
		i++
	}

}

func convertIteratorToSlice(keeper Keeper, iterator sdk.Iterator) []types.ID {
	var queue []types.ID
	for ; iterator.Valid(); iterator.Next() {
		var auctionID types.ID
		keeper.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &auctionID)
		queue = append(queue, auctionID)
	}
	return queue
}
