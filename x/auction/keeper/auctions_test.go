package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(types.DefaultBidDuration))
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
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(types.DefaultBidDuration))
	require.NoError(t, keeper.CloseAuction(ctx, auctionID))
	// Check seller's coins increased
	tApp.CheckBalance(t, ctx, seller, cs(c("token1", 80), c("token2", 110)))
}

func TestForwardReverseAuctionBasic(t *testing.T) {
	// Setup
	_, addrs := app.GeneratePrivKeyAddressPairs(4)
	buyer := addrs[0]
	returnAddrs := addrs[1:]
	returnWeights := is(30, 20, 10)
	sellerModName := liquidator.ModuleName
	sellerAddr := supply.NewModuleAddress(sellerModName)

	tApp := app.NewTestApp()
	sellerAcc := supply.NewEmptyModuleAccount(sellerModName)
	require.NoError(t, sellerAcc.SetCoins(cs(c("token1", 100), c("token2", 100))))
	tApp.InitializeFromGenesisStates(
		NewAuthGenStateFromAccs(authexported.GenesisAccounts{
			auth.NewBaseAccount(buyer, cs(c("token1", 100), c("token2", 100)), nil, 0, 0),
			auth.NewBaseAccount(returnAddrs[0], cs(c("token1", 100), c("token2", 100)), nil, 0, 0),
			auth.NewBaseAccount(returnAddrs[1], cs(c("token1", 100), c("token2", 100)), nil, 0, 0),
			auth.NewBaseAccount(returnAddrs[2], cs(c("token1", 100), c("token2", 100)), nil, 0, 0),
			sellerAcc,
		}),
	)
	ctx := tApp.NewContext(false, abci.Header{})
	keeper := tApp.GetAuctionKeeper()

	// Start auction
	auctionID, err := keeper.StartForwardReverseAuction(ctx, sellerModName, c("token1", 20), c("token2", 50), returnAddrs, returnWeights) // seller, lot, maxBid, otherPerson
	require.NoError(t, err)
	// Check seller's coins have decreased
	tApp.CheckBalance(t, ctx, sellerAddr, cs(c("token1", 80), c("token2", 100)))

	// Place a forward bid
	require.NoError(t, keeper.PlaceBid(ctx, 0, buyer, c("token2", 10), c("token1", 20))) // bid, lot
	// Check bidder's coins have decreased
	tApp.CheckBalance(t, ctx, buyer, cs(c("token1", 100), c("token2", 90)))
	// Check seller's coins have increased
	tApp.CheckBalance(t, ctx, sellerAddr, cs(c("token1", 80), c("token2", 110)))
	// Check return addresses have not received coins
	for _, ra := range returnAddrs {
		tApp.CheckBalance(t, ctx, ra, cs(c("token1", 100), c("token2", 100)))
	}

	// Place a reverse bid
	require.NoError(t, keeper.PlaceBid(ctx, 0, buyer, c("token2", 50), c("token1", 15))) // bid, lot
	// Check bidder's coins have decreased
	tApp.CheckBalance(t, ctx, buyer, cs(c("token1", 100), c("token2", 50)))
	// Check seller's coins have increased
	tApp.CheckBalance(t, ctx, sellerAddr, cs(c("token1", 80), c("token2", 150)))
	// Check return addresses have received coins
	tApp.CheckBalance(t, ctx, returnAddrs[0], cs(c("token1", 102), c("token2", 100)))
	tApp.CheckBalance(t, ctx, returnAddrs[1], cs(c("token1", 102), c("token2", 100)))
	tApp.CheckBalance(t, ctx, returnAddrs[2], cs(c("token1", 101), c("token2", 100)))

	// Close auction at just after auction expiry
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(types.DefaultBidDuration))
	require.NoError(t, keeper.CloseAuction(ctx, auctionID))
	// Check buyer's coins increased
	tApp.CheckBalance(t, ctx, buyer, cs(c("token1", 115), c("token2", 50)))
}

func TestStartForwardAuction(t *testing.T) {
	someTime := time.Date(1998, time.January, 1, 0, 0, 0, 0, time.UTC)
	type args struct {
		seller   string
		lot      sdk.Coin
		bidDenom string
	}
	testCases := []struct {
		name       string
		blockTime  time.Time
		args       args
		expectPass bool
	}{
		{
			"normal",
			someTime,
			args{liquidator.ModuleName, c("stable", 10), "gov"},
			true,
		},
		{
			"no module account",
			someTime,
			args{"nonExistentModule", c("stable", 10), "gov"},
			false,
		},
		{
			"not enough coins",
			someTime,
			args{liquidator.ModuleName, c("stable", 101), "gov"},
			false,
		},
		{
			"incorrect denom",
			someTime,
			args{liquidator.ModuleName, c("notacoin", 10), "gov"},
			false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// setup
			initialLiquidatorCoins := cs(c("stable", 100))
			tApp := app.NewTestApp()

			liqAcc := supply.NewEmptyModuleAccount(liquidator.ModuleName, supply.Burner) // TODO could add test to check for burner permissions
			require.NoError(t, liqAcc.SetCoins(initialLiquidatorCoins))
			tApp.InitializeFromGenesisStates(
				NewAuthGenStateFromAccs(authexported.GenesisAccounts{liqAcc}),
			)
			ctx := tApp.NewContext(false, abci.Header{}).WithBlockTime(tc.blockTime)
			keeper := tApp.GetAuctionKeeper()

			// run function under test
			id, err := keeper.StartForwardAuction(ctx, tc.args.seller, tc.args.lot, tc.args.bidDenom)

			// check
			sk := tApp.GetSupplyKeeper()
			liquidatorCoins := sk.GetModuleAccount(ctx, liquidator.ModuleName).GetCoins()
			actualAuc, found := keeper.GetAuction(ctx, id)

			if tc.expectPass {
				require.NoError(t, err)
				// check coins moved
				require.Equal(t, initialLiquidatorCoins.Sub(cs(tc.args.lot)), liquidatorCoins)
				// check auction in store and is correct
				require.True(t, found)
				expectedAuction := types.Auction(types.ForwardAuction{BaseAuction: types.BaseAuction{
					ID:         types.ID(0),
					Initiator:  tc.args.seller,
					Lot:        tc.args.lot,
					Bidder:     nil,
					Bid:        c(tc.args.bidDenom, 0),
					EndTime:    tc.blockTime.Add(types.DefaultMaxAuctionDuration),
					MaxEndTime: tc.blockTime.Add(types.DefaultMaxAuctionDuration),
				}})
				require.Equal(t, expectedAuction, actualAuc)
			} else {
				require.Error(t, err)
				// check coins not moved
				require.Equal(t, initialLiquidatorCoins, liquidatorCoins)
				// check auction not in store
				require.False(t, found)
			}
		})
	}
}
