package keeper_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/auction/types"
)

type AuctionType int

const (
	Invalid    AuctionType = 0
	Surplus    AuctionType = 1
	Debt       AuctionType = 2
	Collateral AuctionType = 3
)

func TestAuctionBidding(t *testing.T) {
	someTime := time.Date(0001, time.January, 1, 0, 0, 0, 0, time.UTC)

	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	buyer := addrs[0]
	secondBuyer := addrs[1]
	modName := "liquidator"
	collateralAddrs := addrs[2:]
	collateralWeights := is(30, 20, 10)

	type auctionArgs struct {
		auctionType AuctionType
		seller      string
		lot         sdk.Coin
		bid         sdk.Coin
		debt        sdk.Coin
		addresses   []sdk.AccAddress
		weights     []sdk.Int
	}

	type bidArgs struct {
		bidder sdk.AccAddress
		amount sdk.Coin
	}

	tests := []struct {
		name            string
		auctionArgs     auctionArgs
		setupBids       []bidArgs
		bidArgs         bidArgs
		expectedError   error
		expectedEndTime time.Time
		expectedBidder  sdk.AccAddress
		expectedBid     sdk.Coin
		expectPass      bool
		expectPanic     bool
	}{
		{
			"basic: auction doesn't exist",
			auctionArgs{Surplus, "", c("token1", 1), c("token2", 1), sdk.Coin{}, []sdk.AccAddress{}, []sdk.Int{}},
			nil,
			bidArgs{buyer, c("token2", 10)},
			types.ErrAuctionNotFound,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 10),
			false,
			true,
		},
		{
			"basic: closed auction",
			auctionArgs{Surplus, modName, c("token1", 100), c("token2", 10), sdk.Coin{}, []sdk.AccAddress{}, []sdk.Int{}},
			nil,
			bidArgs{buyer, c("token2", 10)},
			types.ErrAuctionHasExpired,
			types.DistantFuture,
			nil,
			c("token2", 0),
			false,
			false,
		},
		{
			"surplus: normal",
			auctionArgs{Surplus, modName, c("token1", 100), c("token2", 10), sdk.Coin{}, []sdk.AccAddress{}, []sdk.Int{}},
			nil,
			bidArgs{buyer, c("token2", 10)},
			nil,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 10),
			true,
			false,
		},
		{
			"surplus: second bidder",
			auctionArgs{Surplus, modName, c("token1", 100), c("token2", 10), sdk.Coin{}, []sdk.AccAddress{}, []sdk.Int{}},
			[]bidArgs{{buyer, c("token2", 10)}},
			bidArgs{secondBuyer, c("token2", 11)},
			nil,
			someTime.Add(types.DefaultBidDuration),
			secondBuyer,
			c("token2", 11),
			true,
			false,
		},
		{
			"surplus: invalid bid denom",
			auctionArgs{Surplus, modName, c("token1", 100), c("token2", 10), sdk.Coin{}, []sdk.AccAddress{}, []sdk.Int{}},
			nil,
			bidArgs{buyer, c("badtoken", 10)},
			types.ErrInvalidBidDenom,
			types.DistantFuture,
			nil, // surplus auctions are created with initial bidder as a nil address
			c("token2", 0),
			false,
			false,
		},
		{
			"surplus: invalid bid (less than)",
			auctionArgs{Surplus, modName, c("token1", 100), c("token2", 0), sdk.Coin{}, []sdk.AccAddress{}, []sdk.Int{}},
			[]bidArgs{{buyer, c("token2", 100)}},
			bidArgs{buyer, c("token2", 99)},
			types.ErrBidTooSmall,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 100),
			false,
			false,
		},
		{
			"surplus: invalid bid (equal)",
			auctionArgs{Surplus, modName, c("token1", 100), c("token2", 0), sdk.Coin{}, []sdk.AccAddress{}, []sdk.Int{}},
			nil,
			bidArgs{buyer, c("token2", 0)}, // min bid is technically 0 at default 5%, but it's capped at 1
			types.ErrBidTooSmall,
			types.DistantFuture,
			nil, // surplus auctions are created with initial bidder as a nil address
			c("token2", 0),
			false,
			false,
		},
		{
			"surplus: invalid bid (less than min increment)",
			auctionArgs{Surplus, modName, c("token1", 100), c("token2", 0), sdk.Coin{}, []sdk.AccAddress{}, []sdk.Int{}},
			[]bidArgs{{buyer, c("token2", 100)}},
			bidArgs{buyer, c("token2", 104)}, // min bid is 105 at default 5%
			types.ErrBidTooSmall,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 100),
			false,
			false,
		},
		{
			"debt: normal",
			auctionArgs{Debt, modName, c("token1", 20), c("token2", 100), c("debt", 100), []sdk.AccAddress{}, []sdk.Int{}}, // initial bid, lot
			nil,
			bidArgs{buyer, c("token1", 10)},
			nil,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 100),
			true,
			false,
		},
		{
			"debt: second bidder",
			auctionArgs{Debt, modName, c("token1", 20), c("token2", 100), c("debt", 100), []sdk.AccAddress{}, []sdk.Int{}}, // initial bid, lot
			[]bidArgs{{buyer, c("token1", 10)}},
			bidArgs{secondBuyer, c("token1", 9)},
			nil,
			someTime.Add(types.DefaultBidDuration),
			secondBuyer,
			c("token2", 100),
			true,
			false,
		},
		{
			"debt: invalid lot denom",
			auctionArgs{Debt, modName, c("token1", 20), c("token2", 100), c("debt", 100), []sdk.AccAddress{}, []sdk.Int{}}, // initial bid, lot
			nil,
			bidArgs{buyer, c("badtoken", 10)},
			types.ErrInvalidLotDenom,
			types.DistantFuture,
			supply.NewModuleAddress(modName),
			c("token2", 100),
			false,
			false,
		},
		{
			"debt: invalid lot size (larger)",
			auctionArgs{Debt, modName, c("token1", 20), c("token2", 100), c("debt", 100), []sdk.AccAddress{}, []sdk.Int{}},
			nil,
			bidArgs{buyer, c("token1", 21)},
			types.ErrLotTooLarge,
			types.DistantFuture,
			supply.NewModuleAddress(modName),
			c("token2", 100),
			false,
			false,
		},
		{
			"debt: invalid lot size (equal)",
			auctionArgs{Debt, modName, c("token1", 20), c("token2", 100), c("debt", 100), []sdk.AccAddress{}, []sdk.Int{}},
			nil,
			bidArgs{buyer, c("token1", 20)},
			types.ErrLotTooLarge,
			types.DistantFuture,
			supply.NewModuleAddress(modName),
			c("token2", 100),
			false,
			false,
		},
		{
			"debt: invalid lot size (larger than min increment)",
			auctionArgs{Debt, modName, c("token1", 60), c("token2", 100), c("debt", 100), []sdk.AccAddress{}, []sdk.Int{}},
			nil,
			bidArgs{buyer, c("token1", 58)}, // max lot at default 5% is 57
			types.ErrLotTooLarge,
			types.DistantFuture,
			supply.NewModuleAddress(modName),
			c("token2", 100),
			false, false,
		},
		{
			"collateral [forward]: normal",
			auctionArgs{Collateral, modName, c("token1", 20), c("token2", 100), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			nil,
			bidArgs{buyer, c("token2", 10)},
			nil,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 10),
			true,
			false,
		},
		{
			"collateral [forward]: second bidder",
			auctionArgs{Collateral, modName, c("token1", 20), c("token2", 100), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			[]bidArgs{{buyer, c("token2", 10)}},
			bidArgs{secondBuyer, c("token2", 11)},
			nil,
			someTime.Add(types.DefaultBidDuration),
			secondBuyer,
			c("token2", 11),
			true,
			false,
		},
		{
			"collateral [forward]: invalid bid denom",
			auctionArgs{Collateral, modName, c("token1", 20), c("token2", 100), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			nil,
			bidArgs{buyer, c("badtoken", 10)},
			types.ErrInvalidBidDenom,
			types.DistantFuture,
			nil,
			c("token2", 0),
			false,
			false,
		},
		{
			"collateral [forward]: invalid bid size (smaller)",
			auctionArgs{Collateral, modName, c("token1", 20), c("token2", 100), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			[]bidArgs{{buyer, c("token2", 10)}},
			bidArgs{buyer, c("token2", 9)},
			types.ErrBidTooSmall,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 10),
			false,
			false,
		},
		{
			"collateral [forward]: invalid bid size (equal)",
			auctionArgs{Collateral, modName, c("token1", 20), c("token2", 100), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			nil,
			bidArgs{buyer, c("token2", 0)},
			types.ErrBidTooSmall,
			types.DistantFuture,
			nil,
			c("token2", 0),
			false,
			false,
		},
		{
			"collateral [forward]: invalid bid size (less than min increment)",
			auctionArgs{Collateral, modName, c("token1", 20), c("token2", 100), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			[]bidArgs{{buyer, c("token2", 50)}},
			bidArgs{buyer, c("token2", 51)},
			types.ErrBidTooSmall,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 50),
			false,
			false,
		},
		{
			"collateral [forward]: less than min increment but equal to maxBid",
			auctionArgs{Collateral, modName, c("token1", 20), c("token2", 100), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			[]bidArgs{{buyer, c("token2", 99)}},
			bidArgs{buyer, c("token2", 100)}, // min bid at default 5% is 104
			nil,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 100),
			true,
			false,
		},
		{
			"collateral [forward]: invalid bid size (greater than max)",
			auctionArgs{Collateral, modName, c("token1", 20), c("token2", 100), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			nil,
			bidArgs{buyer, c("token2", 101)},
			types.ErrBidTooLarge,
			types.DistantFuture,
			nil,
			c("token2", 0),
			false,
			false,
		},
		{
			"collateral [reverse]: normal",
			auctionArgs{Collateral, modName, c("token1", 20), c("token2", 50), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			[]bidArgs{{buyer, c("token2", 50)}}, // put auction into reverse phase
			bidArgs{buyer, c("token1", 15)},
			nil,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 50),
			true,
			false,
		},
		{
			"collateral [reverse]: second bidder",
			auctionArgs{Collateral, modName, c("token1", 20), c("token2", 50), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			[]bidArgs{{buyer, c("token2", 50)}, {buyer, c("token1", 15)}},                                                         // put auction into reverse phase, and add a reverse phase bid
			bidArgs{secondBuyer, c("token1", 14)},
			nil,
			someTime.Add(types.DefaultBidDuration),
			secondBuyer,
			c("token2", 50),
			true,
			false,
		},
		{
			"collateral [reverse]: invalid lot denom",
			auctionArgs{Collateral, modName, c("token1", 20), c("token2", 50), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			[]bidArgs{{buyer, c("token2", 50)}}, // put auction into reverse phase
			bidArgs{buyer, c("badtoken", 15)},
			types.ErrInvalidLotDenom,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 50),
			false,
			false,
		},
		{
			"collateral [reverse]: invalid lot size (greater)",
			auctionArgs{Collateral, modName, c("token1", 20), c("token2", 50), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			[]bidArgs{{buyer, c("token2", 50)}},                                                                                   // put auction into reverse phase
			bidArgs{buyer, c("token1", 21)},
			types.ErrLotTooLarge,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 50),
			false,
			false,
		},
		{
			"collateral [reverse]: invalid lot size (equal)",
			auctionArgs{Collateral, modName, c("token1", 20), c("token2", 50), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			[]bidArgs{{buyer, c("token2", 50)}},                                                                                   // put auction into reverse phase
			bidArgs{buyer, c("token1", 20)},
			types.ErrLotTooLarge,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 50),
			false,
			false,
		},
		{
			"collateral [reverse]: invalid lot size (larger than min increment)",
			auctionArgs{Collateral, modName, c("token1", 60), c("token2", 50), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			[]bidArgs{{buyer, c("token2", 50)}}, // put auction into reverse phase
			bidArgs{buyer, c("token1", 58)},     // max lot at default 5% is 57
			types.ErrLotTooLarge,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 50),
			false,
			false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test
			tApp := app.NewTestApp()
			// Set up seller account
			sellerAcc := supply.NewEmptyModuleAccount(modName, supply.Minter, supply.Burner)
			require.NoError(t, sellerAcc.SetCoins(cs(c("token1", 1000), c("token2", 1000), c("debt", 1000))))
			// Initialize genesis accounts
			tApp.InitializeFromGenesisStates(
				NewAuthGenStateFromAccs(authexported.GenesisAccounts{
					auth.NewBaseAccount(buyer, cs(c("token1", 1000), c("token2", 1000)), nil, 0, 0),
					auth.NewBaseAccount(secondBuyer, cs(c("token1", 1000), c("token2", 1000)), nil, 0, 0),
					auth.NewBaseAccount(collateralAddrs[0], cs(c("token1", 1000), c("token2", 1000)), nil, 0, 0),
					auth.NewBaseAccount(collateralAddrs[1], cs(c("token1", 1000), c("token2", 1000)), nil, 0, 0),
					auth.NewBaseAccount(collateralAddrs[2], cs(c("token1", 1000), c("token2", 1000)), nil, 0, 0),
					sellerAcc,
				}),
			)
			ctx := tApp.NewContext(false, abci.Header{})
			keeper := tApp.GetAuctionKeeper()
			bank := tApp.GetBankKeeper()

			// Start Auction
			var id uint64
			var err error
			switch tc.auctionArgs.auctionType {
			case Surplus:
				if tc.expectPanic {
					require.Panics(t, func() {
						id, err = keeper.StartSurplusAuction(ctx, tc.auctionArgs.seller, tc.auctionArgs.lot, tc.auctionArgs.bid.Denom)
					})
				} else {
					id, err = keeper.StartSurplusAuction(ctx, tc.auctionArgs.seller, tc.auctionArgs.lot, tc.auctionArgs.bid.Denom)
				}
			case Debt:
				id, err = keeper.StartDebtAuction(ctx, tc.auctionArgs.seller, tc.auctionArgs.bid, tc.auctionArgs.lot, tc.auctionArgs.debt)
			case Collateral:
				id, err = keeper.StartCollateralAuction(ctx, tc.auctionArgs.seller, tc.auctionArgs.lot, tc.auctionArgs.bid, tc.auctionArgs.addresses, tc.auctionArgs.weights, tc.auctionArgs.debt) // seller, lot, maxBid, otherPerson
			default:
				t.Fail()
			}

			require.NoError(t, err)

			// Place setup bids
			for _, b := range tc.setupBids {
				require.NoError(t, keeper.PlaceBid(ctx, id, b.bidder, b.amount))
			}

			// Close the auction early to test late bidding (if applicable)
			if strings.Contains(tc.name, "closed") {
				ctx = ctx.WithBlockTime(types.DistantFuture.Add(1))
			}

			// Store some state for use in checks
			oldAuction, found := keeper.GetAuction(ctx, id)
			var oldBidder sdk.AccAddress
			if found {
				oldBidder = oldAuction.GetBidder()
			}
			oldBidderOldCoins := bank.GetCoins(ctx, oldBidder)
			newBidderOldCoins := bank.GetCoins(ctx, tc.bidArgs.bidder)

			// Place bid on auction
			err = keeper.PlaceBid(ctx, id, tc.bidArgs.bidder, tc.bidArgs.amount)

			// Check success/failure
			if tc.expectPass {
				require.Nil(t, err)
				// Check auction was found
				newAuction, found := keeper.GetAuction(ctx, id)
				require.True(t, found)
				// Check auction values
				require.Equal(t, modName, newAuction.GetInitiator())
				require.Equal(t, tc.expectedBidder, newAuction.GetBidder())
				require.Equal(t, tc.expectedBid, newAuction.GetBid())
				require.Equal(t, tc.expectedEndTime, newAuction.GetEndTime())

				// Check coins have moved between bidder and previous bidder
				bidAmt := tc.bidArgs.amount
				switch tc.auctionArgs.auctionType {
				case Debt:
					bidAmt = oldAuction.GetBid()
				case Collateral:
					collatAuction, ok := oldAuction.(types.CollateralAuction)
					require.True(t, ok, tc.name)
					if collatAuction.IsReversePhase() {
						bidAmt = oldAuction.GetBid()
					}
				}
				if oldBidder.Equals(tc.bidArgs.bidder) { // same bidder
					require.Equal(t, newBidderOldCoins.Sub(cs(bidAmt.Sub(oldAuction.GetBid()))), bank.GetCoins(ctx, tc.bidArgs.bidder))
				} else {
					require.Equal(t, newBidderOldCoins.Sub(cs(bidAmt)), bank.GetCoins(ctx, tc.bidArgs.bidder)) // wrapping in cs() to avoid comparing nil and empty coins
					if oldBidder.Equals(supply.NewModuleAddress(oldAuction.GetInitiator())) {                  // handle checking debt coins for case debt auction has had no bids placed yet TODO make this less confusing
						require.Equal(t, oldBidderOldCoins.Add(oldAuction.GetBid()).Add(c("debt", oldAuction.GetBid().Amount.Int64())), bank.GetCoins(ctx, oldBidder))
					} else {
						require.Equal(t, cs(oldBidderOldCoins.Add(oldAuction.GetBid())...), bank.GetCoins(ctx, oldBidder))
					}
				}

			} else {
				// Check expected error code type
				require.NotNil(t, err, "PlaceBid did not return an error") // catch nil values before they cause a panic below
				require.True(t, errors.Is(err, tc.expectedError))

				// Check auction values
				newAuction, found := keeper.GetAuction(ctx, id)
				if found {
					require.Equal(t, modName, newAuction.GetInitiator())
					require.Equal(t, tc.expectedBidder, newAuction.GetBidder())
					require.Equal(t, tc.expectedBid, newAuction.GetBid())
					require.Equal(t, tc.expectedEndTime, newAuction.GetEndTime())
				}

				// Check coins have not moved
				require.Equal(t, newBidderOldCoins, bank.GetCoins(ctx, tc.bidArgs.bidder))
				require.Equal(t, oldBidderOldCoins, bank.GetCoins(ctx, oldBidder))
			}
		})
	}
}
