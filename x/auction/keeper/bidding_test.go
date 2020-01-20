package keeper_test

import (
	"strings"
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
)

type AuctionType int

const (
	Invalid          AuctionType = 0
	Surplus          AuctionType = 1
	Debt             AuctionType = 2
	CollateralPhase1 AuctionType = 3
	CollateralPhase2 AuctionType = 4
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
		bidder       sdk.AccAddress
		amount       sdk.Coin
		secondBidder sdk.AccAddress
	}

	tests := []struct {
		name            string
		auctionArgs     auctionArgs
		bidArgs         bidArgs
		expectedError   sdk.CodeType
		expectedEndTime time.Time
		expectedBidder  sdk.AccAddress
		expectedBid     sdk.Coin
		expectpass      bool
	}{
		{
			"basic: auction doesn't exist",
			auctionArgs{Surplus, "", c("token1", 1), c("token2", 1), sdk.Coin{}, []sdk.AccAddress{}, []sdk.Int{}},
			bidArgs{buyer, c("token2", 10), nil},
			types.CodeAuctionNotFound,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 10),
			false,
		},
		{
			"surplus: normal",
			auctionArgs{Surplus, modName, c("token1", 100), c("token2", 10), sdk.Coin{}, []sdk.AccAddress{}, []sdk.Int{}},
			bidArgs{buyer, c("token2", 10), nil},
			sdk.CodeType(0),
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 10),
			true,
		},
		{
			"surplus: second bidder",
			auctionArgs{Surplus, modName, c("token1", 100), c("token2", 10), sdk.Coin{}, []sdk.AccAddress{}, []sdk.Int{}},
			bidArgs{buyer, c("token2", 10), secondBuyer},
			sdk.CodeType(0),
			someTime.Add(types.DefaultBidDuration),
			secondBuyer,
			c("token2", 11),
			true,
		},
		{
			"surplus: invalid bid denom",
			auctionArgs{Surplus, modName, c("token1", 100), c("token2", 10), sdk.Coin{}, []sdk.AccAddress{}, []sdk.Int{}},
			bidArgs{buyer, c("badtoken", 10), nil},
			types.CodeInvalidBidDenom,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 10),
			false,
		},
		{
			"surplus: invalid bid (equal)",
			auctionArgs{Surplus, modName, c("token1", 100), c("token2", 0), sdk.Coin{}, []sdk.AccAddress{}, []sdk.Int{}},
			bidArgs{buyer, c("token2", 0), nil},
			types.CodeBidTooSmall,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 10),
			false,
		},
		{
			"debt: normal",
			auctionArgs{Debt, modName, c("token1", 20), c("token2", 100), c("debt", 20), []sdk.AccAddress{}, []sdk.Int{}}, // initial bid, lot
			bidArgs{buyer, c("token1", 10), nil},
			sdk.CodeType(0),
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 100),
			true,
		},
		{
			"debt: second bidder",
			auctionArgs{Debt, modName, c("token1", 20), c("token2", 100), c("debt", 20), []sdk.AccAddress{}, []sdk.Int{}}, // initial bid, lot
			bidArgs{buyer, c("token1", 10), secondBuyer},
			sdk.CodeType(0),
			someTime.Add(types.DefaultBidDuration),
			secondBuyer,
			c("token2", 100),
			true,
		},
		{
			"debt: invalid lot denom",
			auctionArgs{Debt, modName, c("token1", 20), c("token2", 100), c("debt", 20), []sdk.AccAddress{}, []sdk.Int{}}, // initial bid, lot
			bidArgs{buyer, c("badtoken", 10), nil},
			types.CodeInvalidLotDenom,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token1", 20),
			false,
		},
		{
			"debt: invalid lot size (larger)",
			auctionArgs{Debt, modName, c("token1", 20), c("token2", 100), c("debt", 20), []sdk.AccAddress{}, []sdk.Int{}},
			bidArgs{buyer, c("token1", 21), nil},
			types.CodeLotTooLarge,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token1", 20),
			false,
		},
		{
			"collateral [forward]: normal",
			auctionArgs{CollateralPhase1, modName, c("token1", 20), c("token2", 100), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			bidArgs{buyer, c("token2", 10), nil},
			sdk.CodeType(0),
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 10),
			true,
		},
		{
			"collateral [forward]: second bidder",
			auctionArgs{CollateralPhase1, modName, c("token1", 20), c("token2", 100), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			bidArgs{buyer, c("token2", 10), secondBuyer},
			sdk.CodeType(0),
			someTime.Add(types.DefaultBidDuration),
			secondBuyer,
			c("token2", 11),
			true,
		},
		{
			"collateral [forward]: invalid bid denom",
			auctionArgs{CollateralPhase1, modName, c("token1", 20), c("token2", 100), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			bidArgs{buyer, c("badtoken", 10), nil},
			types.CodeInvalidBidDenom,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 10),
			false,
		},
		{
			"collateral [forward]: invalid bid size (smaller)",
			auctionArgs{CollateralPhase1, modName, c("token1", 20), c("token2", 100), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			bidArgs{buyer, c("token2", 0), nil},                                                                                          // lot, bid
			types.CodeBidTooSmall,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 10),
			false,
		},
		{
			"collateral [forward]: invalid bid size (greater than max)",
			auctionArgs{CollateralPhase1, modName, c("token1", 20), c("token2", 100), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			bidArgs{buyer, c("token2", 101), nil},                                                                                        // lot, bid
			types.CodeBidTooLarge,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 10),
			false,
		},
		{
			"collateral [reverse]: normal",
			auctionArgs{CollateralPhase2, modName, c("token1", 20), c("token2", 50), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			bidArgs{buyer, c("token1", 15), nil},
			sdk.CodeType(0),
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 50),
			true,
		},
		{
			"collateral [reverse]: second bidder",
			auctionArgs{CollateralPhase2, modName, c("token1", 20), c("token2", 50), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			bidArgs{buyer, c("token1", 15), secondBuyer},
			sdk.CodeType(0),
			someTime.Add(types.DefaultBidDuration),
			secondBuyer,
			c("token2", 50),
			true,
		},
		{
			"collateral [reverse]: invalid lot denom",
			auctionArgs{CollateralPhase2, modName, c("token1", 20), c("token2", 50), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			bidArgs{buyer, c("badtoken", 15), nil},
			types.CodeInvalidLotDenom,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 50),
			false,
		},
		{
			"collateral [reverse]: invalid lot size (equal)",
			auctionArgs{CollateralPhase2, modName, c("token1", 20), c("token2", 50), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			bidArgs{buyer, c("token1", 20), nil},
			types.CodeLotTooLarge,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 50),
			false,
		},
		{
			"collateral [reverse]: invalid lot size (greater)",
			auctionArgs{CollateralPhase2, modName, c("token1", 20), c("token2", 50), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			bidArgs{buyer, c("token1", 21), nil},
			types.CodeLotTooLarge,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 50),
			false,
		},
		{
			"basic: closed auction",
			auctionArgs{Surplus, modName, c("token1", 100), c("token2", 10), sdk.Coin{}, []sdk.AccAddress{}, []sdk.Int{}},
			bidArgs{buyer, c("token2", 10), nil},
			types.CodeAuctionHasExpired,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 10),
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

			// Start Auction
			var id uint64
			var err sdk.Error
			switch tc.auctionArgs.auctionType {
			case Surplus:
				id, _ = keeper.StartSurplusAuction(ctx, tc.auctionArgs.seller, tc.auctionArgs.lot, tc.auctionArgs.bid.Denom)
			case Debt:
				id, _ = keeper.StartDebtAuction(ctx, tc.auctionArgs.seller, tc.auctionArgs.bid, tc.auctionArgs.lot, tc.auctionArgs.debt)
			case CollateralPhase1, CollateralPhase2:
				id, _ = keeper.StartCollateralAuction(ctx, tc.auctionArgs.seller, tc.auctionArgs.lot, tc.auctionArgs.bid, tc.auctionArgs.addresses, tc.auctionArgs.weights, tc.auctionArgs.debt) // seller, lot, maxBid, otherPerson
				// Move CollateralAuction to debt phase by placing max bid
				if tc.auctionArgs.auctionType == CollateralPhase2 {
					err = keeper.PlaceBid(ctx, id, tc.bidArgs.bidder, tc.auctionArgs.bid)
					require.NoError(t, err)
				}
			default:
				t.Fail()
			}

			// Close the auction early to test late bidding (if applicable)
			if strings.Contains(tc.name, "closed") {
				ctx = ctx.WithBlockTime(types.DistantFuture.Add(1))
			}

			// Place bid on auction
			err = keeper.PlaceBid(ctx, id, tc.bidArgs.bidder, tc.bidArgs.amount)

			// Place second bid from new bidder
			if tc.bidArgs.secondBidder != nil {
				// Set bid increase/decrease based on auction type, phase
				var secondBid sdk.Coin
				switch tc.auctionArgs.auctionType {
				case Surplus, CollateralPhase1:
					secondBid = tc.bidArgs.amount.Add(c(tc.bidArgs.amount.Denom, 1))
				case Debt, CollateralPhase2:
					secondBid = tc.bidArgs.amount.Sub(c(tc.bidArgs.amount.Denom, 1))
				default:
					t.Fail()
				}
				// Place the second bid
				err2 := keeper.PlaceBid(ctx, id, tc.bidArgs.secondBidder, secondBid)
				require.NoError(t, err2)
			}

			// Check success/failure
			if tc.expectpass {
				require.Nil(t, err)
				// Get auction from store
				auction, found := keeper.GetAuction(ctx, id)
				require.True(t, found)
				// Check auction values
				require.Equal(t, modName, auction.GetInitiator())
				require.Equal(t, tc.expectedBidder, auction.GetBidder())
				require.Equal(t, tc.expectedBid, auction.GetBid())
				require.Equal(t, tc.expectedEndTime, auction.GetEndTime())
			} else {
				// Check expected error code type
				require.Equal(t, tc.expectedError, err.Result().Code)
			}
		})
	}
}
