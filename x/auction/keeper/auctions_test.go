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
	"github.com/kava-labs/kava/x/cdp"
)

func TestSurplusAuctionBasic(t *testing.T) {
	// Setup
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	buyer := addrs[0]
	sellerModName := cdp.LiquidatorMacc
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
	auctionID, err := keeper.StartSurplusAuction(ctx, sellerModName, c("token1", 20), "token2") // lot, bid denom
	require.NoError(t, err)
	// Check seller's coins have decreased
	tApp.CheckBalance(t, ctx, sellerAddr, cs(c("token1", 80), c("token2", 100)))

	// PlaceBid (bid: 10 token, lot: same as starting)
	require.NoError(t, keeper.PlaceBid(ctx, auctionID, buyer, c("token2", 10)))
	// Check buyer's coins have decreased
	tApp.CheckBalance(t, ctx, buyer, cs(c("token1", 100), c("token2", 90)))
	// Check seller's coins have not increased (because proceeds are burned)
	tApp.CheckBalance(t, ctx, sellerAddr, cs(c("token1", 80), c("token2", 100)))

	// increment bid same bidder
	err = keeper.PlaceBid(ctx, auctionID, buyer, c("token2", 20))
	require.NoError(t, err)

	// Close auction at just at auction expiry time
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(types.DefaultBidDuration))
	require.NoError(t, keeper.CloseAuction(ctx, auctionID))
	// Check buyer's coins increased
	tApp.CheckBalance(t, ctx, buyer, cs(c("token1", 120), c("token2", 80)))
}

func TestDebtAuctionBasic(t *testing.T) {
	// Setup
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	seller := addrs[0]
	buyerModName := cdp.LiquidatorMacc
	buyerAddr := supply.NewModuleAddress(buyerModName)

	tApp := app.NewTestApp()

	buyerAcc := supply.NewEmptyModuleAccount(buyerModName, supply.Minter) // reverse auctions mint payout
	require.NoError(t, buyerAcc.SetCoins(cs(c("debt", 100))))
	tApp.InitializeFromGenesisStates(
		NewAuthGenStateFromAccs(authexported.GenesisAccounts{
			auth.NewBaseAccount(seller, cs(c("token1", 100), c("token2", 100)), nil, 0, 0),
			buyerAcc,
		}),
	)
	ctx := tApp.NewContext(false, abci.Header{})
	keeper := tApp.GetAuctionKeeper()

	// Start auction
	auctionID, err := keeper.StartDebtAuction(ctx, buyerModName, c("token1", 20), c("token2", 99999), c("debt", 20))
	require.NoError(t, err)
	// Check buyer's coins have not decreased (except for debt), as lot is minted at the end
	tApp.CheckBalance(t, ctx, buyerAddr, cs(c("debt", 80)))

	// Place a bid
	require.NoError(t, keeper.PlaceBid(ctx, 0, seller, c("token2", 10)))
	// Check seller's coins have decreased
	tApp.CheckBalance(t, ctx, seller, cs(c("token1", 80), c("token2", 100)))
	// Check buyer's coins have increased
	tApp.CheckBalance(t, ctx, buyerAddr, cs(c("token1", 20), c("debt", 100)))

	// Close auction at just after auction expiry
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(types.DefaultBidDuration))
	require.NoError(t, keeper.CloseAuction(ctx, auctionID))
	// Check seller's coins increased
	tApp.CheckBalance(t, ctx, seller, cs(c("token1", 80), c("token2", 110)))
}

func TestDebtAuctionDebtRemaining(t *testing.T) {
	// Setup
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	seller := addrs[0]
	buyerModName := cdp.LiquidatorMacc
	buyerAddr := supply.NewModuleAddress(buyerModName)

	tApp := app.NewTestApp()

	buyerAcc := supply.NewEmptyModuleAccount(buyerModName, supply.Minter) // reverse auctions mint payout
	require.NoError(t, buyerAcc.SetCoins(cs(c("debt", 100))))
	tApp.InitializeFromGenesisStates(
		NewAuthGenStateFromAccs(authexported.GenesisAccounts{
			auth.NewBaseAccount(seller, cs(c("token1", 100), c("token2", 100)), nil, 0, 0),
			buyerAcc,
		}),
	)
	ctx := tApp.NewContext(false, abci.Header{})
	keeper := tApp.GetAuctionKeeper()

	// Start auction
	auctionID, err := keeper.StartDebtAuction(ctx, buyerModName, c("token1", 10), c("token2", 99999), c("debt", 20))
	require.NoError(t, err)
	// Check buyer's coins have not decreased (except for debt), as lot is minted at the end
	tApp.CheckBalance(t, ctx, buyerAddr, cs(c("debt", 80)))

	// Place a bid
	require.NoError(t, keeper.PlaceBid(ctx, 0, seller, c("token2", 10)))
	// Check seller's coins have decreased
	tApp.CheckBalance(t, ctx, seller, cs(c("token1", 90), c("token2", 100)))
	// Check buyer's coins have increased
	tApp.CheckBalance(t, ctx, buyerAddr, cs(c("token1", 10), c("debt", 90)))

	// Close auction at just after auction expiry
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(types.DefaultBidDuration))
	require.NoError(t, keeper.CloseAuction(ctx, auctionID))
	// Check seller's coins increased
	tApp.CheckBalance(t, ctx, seller, cs(c("token1", 90), c("token2", 110)))
	// check that debt has increased due to corresponding debt being greater than bid
	tApp.CheckBalance(t, ctx, buyerAddr, cs(c("token1", 10), c("debt", 100)))
}

func TestCollateralAuctionBasic(t *testing.T) {
	// Setup
	_, addrs := app.GeneratePrivKeyAddressPairs(4)
	buyer := addrs[0]
	returnAddrs := addrs[1:]
	returnWeights := is(30, 20, 10)
	sellerModName := cdp.LiquidatorMacc
	sellerAddr := supply.NewModuleAddress(sellerModName)

	tApp := app.NewTestApp()
	sellerAcc := supply.NewEmptyModuleAccount(sellerModName)
	require.NoError(t, sellerAcc.SetCoins(cs(c("token1", 100), c("token2", 100), c("debt", 100))))
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
	auctionID, err := keeper.StartCollateralAuction(ctx, sellerModName, c("token1", 20), c("token2", 50), returnAddrs, returnWeights, c("debt", 40))
	require.NoError(t, err)
	// Check seller's coins have decreased
	tApp.CheckBalance(t, ctx, sellerAddr, cs(c("token1", 80), c("token2", 100), c("debt", 60)))

	// Place a forward bid
	require.NoError(t, keeper.PlaceBid(ctx, 0, buyer, c("token2", 10)))
	// Check bidder's coins have decreased
	tApp.CheckBalance(t, ctx, buyer, cs(c("token1", 100), c("token2", 90)))
	// Check seller's coins have increased
	tApp.CheckBalance(t, ctx, sellerAddr, cs(c("token1", 80), c("token2", 110), c("debt", 70)))
	// Check return addresses have not received coins
	for _, ra := range returnAddrs {
		tApp.CheckBalance(t, ctx, ra, cs(c("token1", 100), c("token2", 100)))
	}

	// Place a reverse bid
	require.NoError(t, keeper.PlaceBid(ctx, 0, buyer, c("token2", 50))) // first bid up to max bid to switch phases
	require.NoError(t, keeper.PlaceBid(ctx, 0, buyer, c("token1", 15)))
	// Check bidder's coins have decreased
	tApp.CheckBalance(t, ctx, buyer, cs(c("token1", 100), c("token2", 50)))
	// Check seller's coins have increased
	tApp.CheckBalance(t, ctx, sellerAddr, cs(c("token1", 80), c("token2", 150), c("debt", 100)))
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

func TestCollateralAuctionDebtRemaining(t *testing.T) {
	// Setup
	_, addrs := app.GeneratePrivKeyAddressPairs(4)
	buyer := addrs[0]
	returnAddrs := addrs[1:]
	returnWeights := is(30, 20, 10)
	sellerModName := cdp.LiquidatorMacc
	sellerAddr := supply.NewModuleAddress(sellerModName)

	tApp := app.NewTestApp()
	sellerAcc := supply.NewEmptyModuleAccount(sellerModName)
	require.NoError(t, sellerAcc.SetCoins(cs(c("token1", 100), c("token2", 100), c("debt", 100))))
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
	auctionID, err := keeper.StartCollateralAuction(ctx, sellerModName, c("token1", 20), c("token2", 50), returnAddrs, returnWeights, c("debt", 40))
	require.NoError(t, err)
	// Check seller's coins have decreased
	tApp.CheckBalance(t, ctx, sellerAddr, cs(c("token1", 80), c("token2", 100), c("debt", 60)))

	// Place a forward bid
	require.NoError(t, keeper.PlaceBid(ctx, 0, buyer, c("token2", 10)))
	// Check bidder's coins have decreased
	tApp.CheckBalance(t, ctx, buyer, cs(c("token1", 100), c("token2", 90)))
	// Check seller's coins have increased
	tApp.CheckBalance(t, ctx, sellerAddr, cs(c("token1", 80), c("token2", 110), c("debt", 70)))
	// Check return addresses have not received coins
	for _, ra := range returnAddrs {
		tApp.CheckBalance(t, ctx, ra, cs(c("token1", 100), c("token2", 100)))
	}
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(types.DefaultBidDuration))
	require.NoError(t, keeper.CloseAuction(ctx, auctionID))

	// check that buyers coins have increased
	tApp.CheckBalance(t, ctx, buyer, cs(c("token1", 120), c("token2", 90)))
	// Check return addresses have not received coins
	for _, ra := range returnAddrs {
		tApp.CheckBalance(t, ctx, ra, cs(c("token1", 100), c("token2", 100)))
	}
	// check that token2 has increased by 10, debt by 40, for a net debt increase of 30 debt
	tApp.CheckBalance(t, ctx, sellerAddr, cs(c("token1", 80), c("token2", 110), c("debt", 100)))
}

func TestStartSurplusAuction(t *testing.T) {
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
			args{cdp.LiquidatorMacc, c("stable", 10), "gov"},
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
			args{cdp.LiquidatorMacc, c("stable", 101), "gov"},
			false,
		},
		{
			"incorrect denom",
			someTime,
			args{cdp.LiquidatorMacc, c("notacoin", 10), "gov"},
			false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// setup
			initialLiquidatorCoins := cs(c("stable", 100))
			tApp := app.NewTestApp()

			liqAcc := supply.NewEmptyModuleAccount(cdp.LiquidatorMacc, supply.Burner)
			require.NoError(t, liqAcc.SetCoins(initialLiquidatorCoins))
			tApp.InitializeFromGenesisStates(
				NewAuthGenStateFromAccs(authexported.GenesisAccounts{liqAcc}),
			)
			ctx := tApp.NewContext(false, abci.Header{}).WithBlockTime(tc.blockTime)
			keeper := tApp.GetAuctionKeeper()

			// run function under test
			id, err := keeper.StartSurplusAuction(ctx, tc.args.seller, tc.args.lot, tc.args.bidDenom)

			// check
			sk := tApp.GetSupplyKeeper()
			liquidatorCoins := sk.GetModuleAccount(ctx, cdp.LiquidatorMacc).GetCoins()
			actualAuc, found := keeper.GetAuction(ctx, id)

			if tc.expectPass {
				require.NoError(t, err)
				// check coins moved
				require.Equal(t, initialLiquidatorCoins.Sub(cs(tc.args.lot)), liquidatorCoins)
				// check auction in store and is correct
				require.True(t, found)
				expectedAuction := types.Auction(types.SurplusAuction{BaseAuction: types.BaseAuction{
					ID:              0,
					Initiator:       tc.args.seller,
					Lot:             tc.args.lot,
					Bidder:          nil,
					Bid:             c(tc.args.bidDenom, 0),
					HasReceivedBids: false,
					EndTime:         types.DistantFuture,
					MaxEndTime:      types.DistantFuture,
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

func TestCloseAuction(t *testing.T) {
	// Set up
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	buyer := addrs[0]
	sellerModName := cdp.LiquidatorMacc

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
	id, err := keeper.StartSurplusAuction(ctx, sellerModName, c("token1", 20), "token2") // lot, bid denom
	require.NoError(t, err)

	// Attempt to close the auction before EndTime
	require.Error(t, keeper.CloseAuction(ctx, id))

	// Attempt to close auction that does not exist
	require.Error(t, keeper.CloseAuction(ctx, 999))
}

func TestCloseExpiredAuctions(t *testing.T) {
	// Set up
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	buyer := addrs[0]
	sellerModName := "liquidator"

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

	// Start auction 1
	_, err := keeper.StartSurplusAuction(ctx, sellerModName, c("token1", 20), "token2") // lot, bid denom
	require.NoError(t, err)

	// Start auction 2
	_, err = keeper.StartSurplusAuction(ctx, sellerModName, c("token1", 20), "token2") // lot, bid denom
	require.NoError(t, err)

	// Fast forward the block time
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(types.DefaultMaxAuctionDuration).Add(1))

	// Close expired auctions
	err = keeper.CloseExpiredAuctions(ctx)
	require.NoError(t, err)
}
