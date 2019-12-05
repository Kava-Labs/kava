package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/auction/keeper"
	"github.com/kava-labs/kava/x/auction/types"
)

func TestKeeper_ForwardAuction(t *testing.T) {
	// Setup
	_, addrs := app.GeneratePrivKeyAddressPairs(2)
	seller := addrs[0]
	buyer := addrs[1]

	tApp := app.NewTestApp()
	authGenState := tApp.NewAuthGenStateFromAccounts(addrs, []sdk.Coins{cs(c("token1", 100), c("token2", 100)), cs(c("token1", 100), c("token2", 100))})
	tApp.InitializeFromGenesisStates(authGenState)

	ctx := tApp.NewContext(false, abci.Header{})
	keeper := tApp.GetAuctionKeeper()

	// Create an auction (lot: 20 t1, initialBid: 0 t2)
	auctionID, err := keeper.StartForwardAuction(ctx, seller, c("token1", 20), c("token2", 0)) // lot, initialBid
	require.NoError(t, err)
	// Check seller's coins have decreased
	tApp.CheckBalance(t, ctx, seller, cs(c("token1", 80), c("token2", 100)))

	// PlaceBid (bid: 10 t2, lot: same as starting)
	require.NoError(t, keeper.PlaceBid(ctx, 0, buyer, c("token2", 10), c("token1", 20))) // bid, lot
	// Check buyer's coins have decreased
	tApp.CheckBalance(t, ctx, buyer, cs(c("token1", 100), c("token2", 90)))
	// Check seller's coins have increased
	tApp.CheckBalance(t, ctx, seller, cs(c("token1", 80), c("token2", 110)))

	// Close auction at just after auction expiry
	ctx = ctx.WithBlockHeight(int64(types.DefaultMaxBidDuration))
	require.NoError(t, keeper.CloseAuction(ctx, auctionID))
	// Check buyer's coins increased
	tApp.CheckBalance(t, ctx, buyer, cs(c("token1", 120), c("token2", 90)))
}

func TestKeeper_ReverseAuction(t *testing.T) {
	// Setup
	_, addrs := app.GeneratePrivKeyAddressPairs(2)
	seller := addrs[0]
	buyer := addrs[1]

	tApp := app.NewTestApp()
	authGenState := tApp.NewAuthGenStateFromAccounts(addrs, []sdk.Coins{cs(c("token1", 100), c("token2", 100)), cs(c("token1", 100), c("token2", 100))})
	tApp.InitializeFromGenesisStates(authGenState)

	ctx := tApp.NewContext(false, abci.Header{})
	keeper := tApp.GetAuctionKeeper()

	// Start auction
	auctionID, err := keeper.StartReverseAuction(ctx, buyer, c("token1", 20), c("token2", 99)) // buyer, bid, initialLot
	require.NoError(t, err)
	// Check buyer's coins have decreased
	tApp.CheckBalance(t, ctx, buyer, cs(c("token1", 100), c("token2", 1)))

	// Place a bid
	require.NoError(t, keeper.PlaceBid(ctx, 0, seller, c("token1", 20), c("token2", 10))) // bid, lot
	// Check seller's coins have decreased
	tApp.CheckBalance(t, ctx, seller, cs(c("token1", 80), c("token2", 100)))
	// Check buyer's coins have increased
	tApp.CheckBalance(t, ctx, buyer, cs(c("token1", 120), c("token2", 90)))

	// Close auction at just after auction expiry
	ctx = ctx.WithBlockHeight(int64(types.DefaultMaxBidDuration))
	require.NoError(t, keeper.CloseAuction(ctx, auctionID))
	// Check seller's coins increased
	tApp.CheckBalance(t, ctx, seller, cs(c("token1", 80), c("token2", 110)))
}

func TestKeeper_ForwardReverseAuction(t *testing.T) {
	// Setup
	_, addrs := app.GeneratePrivKeyAddressPairs(3)
	seller := addrs[0]
	buyer := addrs[1]
	recipient := addrs[2]

	tApp := app.NewTestApp()
	authGenState := tApp.NewAuthGenStateFromAccounts(addrs, []sdk.Coins{cs(c("token1", 100), c("token2", 100)), cs(c("token1", 100), c("token2", 100)), cs(c("token1", 100), c("token2", 100))})
	tApp.InitializeFromGenesisStates(authGenState)

	ctx := tApp.NewContext(false, abci.Header{})
	keeper := tApp.GetAuctionKeeper()

	// Start auction
	auctionID, err := keeper.StartForwardReverseAuction(ctx, seller, c("token1", 20), c("token2", 50), recipient) // seller, lot, maxBid, otherPerson
	require.NoError(t, err)
	// Check seller's coins have decreased
	tApp.CheckBalance(t, ctx, seller, cs(c("token1", 80), c("token2", 100)))

	// Place a bid
	require.NoError(t, keeper.PlaceBid(ctx, 0, buyer, c("token2", 50), c("token1", 15))) // bid, lot
	// Check bidder's coins have decreased
	tApp.CheckBalance(t, ctx, buyer, cs(c("token1", 100), c("token2", 50)))
	// Check seller's coins have increased
	tApp.CheckBalance(t, ctx, seller, cs(c("token1", 80), c("token2", 150)))
	// Check "recipient" has received coins
	tApp.CheckBalance(t, ctx, recipient, cs(c("token1", 105), c("token2", 100)))

	// Close auction at just after auction expiry
	ctx = ctx.WithBlockHeight(int64(types.DefaultMaxBidDuration))
	require.NoError(t, keeper.CloseAuction(ctx, auctionID))
	// Check buyer's coins increased
	tApp.CheckBalance(t, ctx, buyer, cs(c("token1", 115), c("token2", 50)))
}

func TestKeeper_SetGetDeleteAuction(t *testing.T) {
	// setup keeper, create auction
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	tApp := app.NewTestApp()
	keeper := tApp.GetAuctionKeeper()
	ctx := tApp.NewContext(true, abci.Header{})
	auction, _ := types.NewForwardAuction(addrs[0], c("usdx", 100), c("kava", 0), types.EndTime(1000))
	id := types.ID(5)
	auction.SetID(id)

	// write and read from store
	keeper.SetAuction(ctx, &auction)
	readAuction, found := keeper.GetAuction(ctx, id)

	// check before and after match
	require.True(t, found)
	require.Equal(t, &auction, readAuction)
	// check auction is in queue
	iter := keeper.GetQueueIterator(ctx, 100000)
	require.Equal(t, 1, len(convertIteratorToSlice(keeper, iter)))
	iter.Close()

	// delete auction
	keeper.DeleteAuction(ctx, id)

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
	tApp := app.NewTestApp()
	keeper := tApp.GetAuctionKeeper()
	ctx := tApp.NewContext(true, abci.Header{})

	// create an example queue
	type queue []struct {
		endTime   types.EndTime
		auctionID types.ID
	}
	q := queue{{1000, 0}, {1300, 2}, {5200, 1}}

	// write and read queue
	for _, v := range q {
		keeper.InsertIntoQueue(ctx, v.endTime, v.auctionID)
	}
	iter := keeper.GetQueueIterator(ctx, 1000)

	// check before and after match
	i := 0
	for ; iter.Valid(); iter.Next() {
		var auctionID types.ID
		tApp.Codec().MustUnmarshalBinaryLengthPrefixed(iter.Value(), &auctionID)
		require.Equal(t, q[i].auctionID, auctionID)
		i++
	}

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
