package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/auction/types"
	"github.com/kava-labs/kava/x/liquidator"
)

func TestForwardAuctionBasic(t *testing.T) {
	// Setup
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	buyer := addrs[0]
	sellerModName := liquidator.ModuleName
	sellerAddr := supply.NewModuleAddress(sellerModName)

	tApp := app.NewTestApp()

	sellerAcc := supply.NewEmptyModuleAccount(sellerModName, supply.Burner) // forward auctions burn proceeds
	require.NoError(t, sellerAcc.SetCoins(cs(c("token1", 100), c("token2", 100))))
	tApp.InitializeFromGenesisStates(
		NewAuthGenStateFromAccs(authexported.GenesisAccounts{
			auth.NewBaseAccount(buyer, cs(c("token1", 100), c("token2", 100)), nil, 0, 0),
			sellerAcc,
		}),
	)
	ctx := tApp.NewContext(false, abci.Header{})
	keeper := tApp.GetAuctionKeeper()

	// Create an auction (lot: 20 token1, initialBid: 0 token2)
	auctionID, err := keeper.StartForwardAuction(ctx, sellerModName, c("token1", 20), "token2") // lot, bid denom
	require.NoError(t, err)
	// Check seller's coins have decreased
	tApp.CheckBalance(t, ctx, sellerAddr, cs(c("token1", 80), c("token2", 100)))

	// PlaceBid (bid: 10 token, lot: same as starting)
	require.NoError(t, keeper.PlaceBid(ctx, auctionID, buyer, c("token2", 10), c("token1", 20))) // bid, lot
	// Check buyer's coins have decreased
	tApp.CheckBalance(t, ctx, buyer, cs(c("token1", 100), c("token2", 90)))
	// Check seller's coins have not increased (because proceeds are burned)
	tApp.CheckBalance(t, ctx, sellerAddr, cs(c("token1", 80), c("token2", 100)))

	// Close auction at just at auction expiry time
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(types.DefaultMaxBidDuration))
	require.NoError(t, keeper.CloseAuction(ctx, auctionID))
	// Check buyer's coins increased
	tApp.CheckBalance(t, ctx, buyer, cs(c("token1", 120), c("token2", 90)))
}

func TestReverseAuctionBasic(t *testing.T) {
	// Setup
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	seller := addrs[0]
	buyerModName := liquidator.ModuleName
	buyerAddr := supply.NewModuleAddress(buyerModName)

	tApp := app.NewTestApp()

	tApp.InitializeFromGenesisStates(
		NewAuthGenStateFromAccs(authexported.GenesisAccounts{
			auth.NewBaseAccount(seller, cs(c("token1", 100), c("token2", 100)), nil, 0, 0),
			supply.NewEmptyModuleAccount(buyerModName, supply.Minter), // reverse auctions mint payout
		}),
	)
	ctx := tApp.NewContext(false, abci.Header{})
	keeper := tApp.GetAuctionKeeper()

	// Start auction
	auctionID, err := keeper.StartReverseAuction(ctx, buyerModName, c("token1", 20), c("token2", 99999)) // buyer, bid, initialLot
	require.NoError(t, err)
	// Check buyer's coins have not decreased, as lot is minted at the end
	tApp.CheckBalance(t, ctx, buyerAddr, nil) // zero coins

	// Place a bid
	require.NoError(t, keeper.PlaceBid(ctx, 0, seller, c("token1", 20), c("token2", 10))) // bid, lot
	// Check seller's coins have decreased
	tApp.CheckBalance(t, ctx, seller, cs(c("token1", 80), c("token2", 100)))
	// Check buyer's coins have increased
	tApp.CheckBalance(t, ctx, buyerAddr, cs(c("token1", 20)))

	// Close auction at just after auction expiry
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(types.DefaultMaxBidDuration))
	require.NoError(t, keeper.CloseAuction(ctx, auctionID))
	// Check seller's coins increased
	tApp.CheckBalance(t, ctx, seller, cs(c("token1", 80), c("token2", 110)))
}

func TestForwardReverseAuctionBasic(t *testing.T) {
	// Setup
	_, addrs := app.GeneratePrivKeyAddressPairs(2)
	buyer := addrs[0]
	recipient := addrs[1]
	sellerModName := liquidator.ModuleName
	sellerAddr := supply.NewModuleAddress(sellerModName)

	tApp := app.NewTestApp()
	sellerAcc := supply.NewEmptyModuleAccount(sellerModName)
	require.NoError(t, sellerAcc.SetCoins(cs(c("token1", 100), c("token2", 100))))
	tApp.InitializeFromGenesisStates(
		NewAuthGenStateFromAccs(authexported.GenesisAccounts{
			auth.NewBaseAccount(buyer, cs(c("token1", 100), c("token2", 100)), nil, 0, 0),
			auth.NewBaseAccount(recipient, cs(c("token1", 100), c("token2", 100)), nil, 0, 0),
			sellerAcc,
		}),
	)
	ctx := tApp.NewContext(false, abci.Header{})
	keeper := tApp.GetAuctionKeeper()

	// Start auction
	auctionID, err := keeper.StartForwardReverseAuction(ctx, sellerModName, c("token1", 20), c("token2", 50), recipient) // seller, lot, maxBid, otherPerson
	require.NoError(t, err)
	// Check seller's coins have decreased
	tApp.CheckBalance(t, ctx, sellerAddr, cs(c("token1", 80), c("token2", 100)))

	// Place a forward bid
	require.NoError(t, keeper.PlaceBid(ctx, 0, buyer, c("token2", 10), c("token1", 20))) // bid, lot
	// Check bidder's coins have decreased
	tApp.CheckBalance(t, ctx, buyer, cs(c("token1", 100), c("token2", 90)))
	// Check seller's coins have increased
	tApp.CheckBalance(t, ctx, sellerAddr, cs(c("token1", 80), c("token2", 110)))
	// Check recipient has not received coins
	tApp.CheckBalance(t, ctx, recipient, cs(c("token1", 100), c("token2", 100)))

	// Place a reverse bid
	require.NoError(t, keeper.PlaceBid(ctx, 0, buyer, c("token2", 50), c("token1", 15))) // bid, lot
	// Check bidder's coins have decreased
	tApp.CheckBalance(t, ctx, buyer, cs(c("token1", 100), c("token2", 50)))
	// Check seller's coins have increased
	tApp.CheckBalance(t, ctx, sellerAddr, cs(c("token1", 80), c("token2", 150)))
	// Check "recipient" has received coins
	tApp.CheckBalance(t, ctx, recipient, cs(c("token1", 105), c("token2", 100)))

	// Close auction at just after auction expiry
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(types.DefaultMaxBidDuration))
	require.NoError(t, keeper.CloseAuction(ctx, auctionID))
	// Check buyer's coins increased
	tApp.CheckBalance(t, ctx, buyer, cs(c("token1", 115), c("token2", 50)))
}
