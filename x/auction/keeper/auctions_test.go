package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/kava-labs/kava/x/auction/testutil"
	"github.com/kava-labs/kava/x/auction/types"
)

type auctionTestSuite struct {
	testutil.Suite
}

func (suite *auctionTestSuite) SetupTest() {
	suite.Suite.SetupTest(4)
}

func TestAuctionTestSuite(t *testing.T) {
	suite.Run(t, new(auctionTestSuite))
}

func (suite *auctionTestSuite) TestSurplusAuctionBasic() {
	buyer := suite.Addrs[0]

	// TODO: use cdp.LiquidatorMacc once CDP module is available
	// sellerModName := cdp.LiquidatorMacc
	sellerAddr := authtypes.NewModuleAddress(suite.ModAcc.Name)
	suite.AddCoinsToNamedModule(suite.ModAcc.Name, cs(c("token1", 100), c("token2", 100)))

	// Create an auction (lot: 20 token1, initialBid: 0 token2)
	auctionID, err := suite.Keeper.StartSurplusAuction(suite.Ctx, suite.ModAcc.Name, c("token1", 20), "token2") // lobid denom
	suite.NoError(err)
	// Check seller's coins have decreased
	suite.CheckAccountBalanceEqual(sellerAddr, cs(c("token1", 80), c("token2", 100)))

	// PlaceBid (bid: 10 token, lot: same as starting)
	suite.NoError(suite.Keeper.PlaceBid(suite.Ctx, auctionID, buyer, c("token2", 10)))
	// Check buyer's coins have decreased
	suite.CheckAccountBalanceEqual(buyer, cs(c("token1", 100), c("token2", 90)))
	// Check seller's coins have not increased (because proceeds are burned)
	suite.CheckAccountBalanceEqual(sellerAddr, cs(c("token1", 80), c("token2", 100)))

	// increment bid same bidder
	err = suite.Keeper.PlaceBid(suite.Ctx, auctionID, buyer, c("token2", 20))
	suite.NoError(err)

	// Close auction just at auction expiry time
	suite.Ctx = suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(types.DefaultBidDuration))
	suite.NoError(suite.Keeper.CloseAuction(suite.Ctx, auctionID))
	// Check buyer's coins increased
	suite.CheckAccountBalanceEqual(buyer, cs(c("token1", 120), c("token2", 80)))
}

func (suite *auctionTestSuite) TestDebtAuctionBasic() {
	// Setup
	seller := suite.Addrs[0]
	suite.AddCoinsToNamedModule(suite.ModAcc.Name, cs(c("debt", 100)))

	// Start auction
	auctionID, err := suite.Keeper.StartDebtAuction(suite.Ctx, suite.ModAcc.Name, c("token1", 20), c("token2", 99999), c("debt", 20))
	suite.NoError(err)
	// Check buyer's coins have not decreased (except for debt), as lot is minted at the end
	suite.CheckAccountBalanceEqual(suite.ModAcc.GetAddress(), cs(c("debt", 80)))

	// Place a bid
	suite.NoError(suite.Keeper.PlaceBid(suite.Ctx, auctionID, seller, c("token2", 10)))

	// Check seller's coins have decreased
	suite.CheckAccountBalanceEqual(seller, cs(c("token1", 80), c("token2", 100)))
	// Check buyer's coins have increased
	suite.CheckAccountBalanceEqual(suite.ModAcc.GetAddress(), cs(c("token1", 20), c("debt", 100)))

	// Close auction at just after auction expiry
	ctx := suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(types.DefaultBidDuration))
	suite.NoError(suite.Keeper.CloseAuction(ctx, auctionID))
	// Check seller's coins increased
	suite.CheckAccountBalanceEqual(seller, cs(c("token1", 80), c("token2", 110)))
}

func (suite *auctionTestSuite) TestDebtAuctionDebtRemaining() {
	seller := suite.Addrs[0]

	buyerAddr := authtypes.NewModuleAddress(suite.ModAcc.Name)
	suite.AddCoinsToNamedModule(suite.ModAcc.Name, cs(c("debt", 100)))

	// Start auction
	auctionID, err := suite.Keeper.StartDebtAuction(suite.Ctx, suite.ModAcc.Name, c("token1", 10), c("token2", 99999), c("debt", 20))
	suite.NoError(err)
	// Check buyer's coins have not decreased (except for debt), as lot is minted at the end
	suite.CheckAccountBalanceEqual(buyerAddr, cs(c("debt", 80)))

	// Place a bid
	suite.NoError(suite.Keeper.PlaceBid(suite.Ctx, auctionID, seller, c("token2", 10)))
	// Check seller's coins have decreased
	suite.CheckAccountBalanceEqual(seller, cs(c("token1", 90), c("token2", 100)))
	// Check buyer's coins have increased
	suite.CheckAccountBalanceEqual(buyerAddr, cs(c("token1", 10), c("debt", 90)))

	// Close auction at just after auction expiry
	ctx := suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(types.DefaultBidDuration))
	suite.NoError(suite.Keeper.CloseAuction(ctx, auctionID))
	// Check seller's coins increased
	suite.CheckAccountBalanceEqual(seller, cs(c("token1", 90), c("token2", 110)))
	// check that debt has increased due to corresponding debt being greater than bid
	suite.CheckAccountBalanceEqual(buyerAddr, cs(c("token1", 10), c("debt", 100)))
}

func (suite *auctionTestSuite) TestCollateralAuctionBasic() {
	// Setup
	buyer := suite.Addrs[0]
	returnAddrs := suite.Addrs[1:]
	returnWeights := is(30, 20, 10)
	sellerModName := suite.ModAcc.Name
	sellerAddr := suite.ModAcc.GetAddress()
	suite.AddCoinsToNamedModule(sellerModName, cs(c("token1", 100), c("token2", 100), c("debt", 100)))

	// Start auction
	auctionID, err := suite.Keeper.StartCollateralAuction(suite.Ctx, sellerModName, c("token1", 20), c("token2", 50), returnAddrs, returnWeights, c("debt", 40))
	suite.NoError(err)
	// Check seller's coins have decreased
	suite.CheckAccountBalanceEqual(sellerAddr, cs(c("token1", 80), c("token2", 100), c("debt", 60)))

	// Place a forward bid
	suite.NoError(suite.Keeper.PlaceBid(suite.Ctx, auctionID, buyer, c("token2", 10)))
	// Check bidder's coins have decreased
	suite.CheckAccountBalanceEqual(buyer, cs(c("token1", 100), c("token2", 90)))
	// Check seller's coins have increased
	suite.CheckAccountBalanceEqual(sellerAddr, cs(c("token1", 80), c("token2", 110), c("debt", 70)))
	// Check return addresses have not received coins
	for _, ra := range suite.Addrs[1:] {
		suite.CheckAccountBalanceEqual(ra, cs(c("token1", 100), c("token2", 100)))
	}

	// Place a reverse bid
	suite.NoError(suite.Keeper.PlaceBid(suite.Ctx, auctionID, buyer, c("token2", 50))) // first bid up to max bid to switch phases
	suite.NoError(suite.Keeper.PlaceBid(suite.Ctx, auctionID, buyer, c("token1", 15)))
	// Check bidder's coins have decreased
	suite.CheckAccountBalanceEqual(buyer, cs(c("token1", 100), c("token2", 50)))
	// Check seller's coins have increased
	suite.CheckAccountBalanceEqual(sellerAddr, cs(c("token1", 80), c("token2", 150), c("debt", 100)))
	// Check return addresses have received coins
	suite.CheckAccountBalanceEqual(suite.Addrs[1], cs(c("token1", 102), c("token2", 100)))
	suite.CheckAccountBalanceEqual(suite.Addrs[2], cs(c("token1", 102), c("token2", 100)))
	suite.CheckAccountBalanceEqual(suite.Addrs[3], cs(c("token1", 101), c("token2", 100)))

	// Close auction at just after auction expiry
	ctx := suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(types.DefaultBidDuration))
	suite.NoError(suite.Keeper.CloseAuction(ctx, auctionID))
	// Check buyer's coins increased
	suite.CheckAccountBalanceEqual(buyer, cs(c("token1", 115), c("token2", 50)))
}

func (suite *auctionTestSuite) TestCollateralAuctionDebtRemaining() {
	// Setup
	buyer := suite.Addrs[0]
	returnAddrs := suite.Addrs[1:]
	returnWeights := is(30, 20, 10)
	sellerModName := suite.ModAcc.Name
	sellerAddr := suite.ModAcc.GetAddress()
	suite.AddCoinsToNamedModule(sellerModName, cs(c("token1", 100), c("token2", 100), c("debt", 100)))

	// Start auction
	auctionID, err := suite.Keeper.StartCollateralAuction(suite.Ctx, sellerModName, c("token1", 20), c("token2", 50), returnAddrs, returnWeights, c("debt", 40))
	suite.NoError(err)
	// Check seller's coins have decreased
	suite.CheckAccountBalanceEqual(sellerAddr, cs(c("token1", 80), c("token2", 100), c("debt", 60)))

	// Place a forward bid
	suite.NoError(suite.Keeper.PlaceBid(suite.Ctx, auctionID, buyer, c("token2", 10)))
	// Check bidder's coins have decreased
	suite.CheckAccountBalanceEqual(buyer, cs(c("token1", 100), c("token2", 90)))
	// Check seller's coins have increased
	suite.CheckAccountBalanceEqual(sellerAddr, cs(c("token1", 80), c("token2", 110), c("debt", 70)))
	// Check return addresses have not received coins
	for _, ra := range suite.Addrs[1:] {
		suite.CheckAccountBalanceEqual(ra, cs(c("token1", 100), c("token2", 100)))
	}
	ctx := suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(types.DefaultBidDuration))
	suite.NoError(suite.Keeper.CloseAuction(ctx, auctionID))

	// check that buyers coins have increased
	suite.CheckAccountBalanceEqual(buyer, cs(c("token1", 120), c("token2", 90)))
	// Check return addresses have not received coins
	for _, ra := range suite.Addrs[1:] {
		suite.CheckAccountBalanceEqual(ra, cs(c("token1", 100), c("token2", 100)))
	}
	// check that token2 has increased by 10, debt by 40, for a net debt increase of 30 debt
	suite.CheckAccountBalanceEqual(sellerAddr, cs(c("token1", 80), c("token2", 110), c("debt", 100)))
}

func (suite *auctionTestSuite) TestStartSurplusAuction() {
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
		expPanic   bool
	}{
		{
			"normal",
			someTime,
			args{suite.ModAcc.Name, c("stable", 10), "gov"},
			true, false,
		},
		{
			"no module account",
			someTime,
			args{"nonExistentModule", c("stable", 10), "gov"},
			false, true,
		},
		{
			"not enough coins",
			someTime,
			args{suite.ModAcc.Name, c("stable", 101), "gov"},
			false, false,
		},
		{
			"incorrect denom",
			someTime,
			args{suite.ModAcc.Name, c("notacoin", 10), "gov"},
			false, false,
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			// setup
			initialLiquidatorCoins := cs(c("stable", 100))
			suite.AddCoinsToNamedModule(suite.ModAcc.Name, initialLiquidatorCoins)

			// run function under test
			var (
				id  uint64
				err error
			)
			if tc.expPanic {
				suite.Panics(func() {
					_, _ = suite.Keeper.StartSurplusAuction(suite.Ctx, tc.args.seller, tc.args.lot, tc.args.bidDenom)
				}, tc.name)
			} else {
				id, err = suite.Keeper.StartSurplusAuction(suite.Ctx, tc.args.seller, tc.args.lot, tc.args.bidDenom)
			}

			// check
			liquidatorCoins := suite.BankKeeper.GetAllBalances(suite.Ctx, suite.ModAcc.GetAddress())
			actualAuc, found := suite.Keeper.GetAuction(suite.Ctx, id)
			if tc.expectPass {
				suite.NoError(err, tc.name)
				// check coins moved
				suite.Equal(initialLiquidatorCoins.Sub(cs(tc.args.lot)), liquidatorCoins, tc.name)
				// check auction in store and is correct
				suite.True(found, tc.name)

				surplusAuction := types.SurplusAuction{BaseAuction: types.BaseAuction{
					ID:              id,
					Initiator:       tc.args.seller,
					Lot:             tc.args.lot,
					Bidder:          nil,
					Bid:             c(tc.args.bidDenom, 0),
					HasReceivedBids: false,
					EndTime:         types.DistantFuture,
					MaxEndTime:      types.DistantFuture,
				}}
				suite.Equal(&surplusAuction, actualAuc, tc.name)
			} else if !tc.expPanic && !tc.expectPass {
				suite.Error(err, tc.name)
				// check coins not moved
				suite.Equal(initialLiquidatorCoins, liquidatorCoins, tc.name)
				// check auction not in store
				suite.False(found, tc.name)
			}
		})
	}
}

func (suite *auctionTestSuite) TestCloseAuction() {
	suite.AddCoinsToNamedModule(suite.ModAcc.Name, cs(c("token1", 100), c("token2", 100)))

	// Create an auction (lot: 20 token1, initialBid: 0 token2)
	id, err := suite.Keeper.StartSurplusAuction(suite.Ctx, suite.ModAcc.Name, c("token1", 20), "token2") // lot, bid denom
	suite.NoError(err)

	// Attempt to close the auction before EndTime
	suite.Error(suite.Keeper.CloseAuction(suite.Ctx, id))

	// Attempt to close auction that does not exist
	suite.Error(suite.Keeper.CloseAuction(suite.Ctx, 999))
}

func (suite *auctionTestSuite) TestCloseExpiredAuctions() {
	suite.AddCoinsToNamedModule(suite.ModAcc.Name, cs(c("token1", 100), c("token2", 100)))

	// Start auction 1
	_, err := suite.Keeper.StartSurplusAuction(suite.Ctx, suite.ModAcc.Name, c("token1", 20), "token2") // lot, bid denom
	suite.NoError(err)

	// Start auction 2
	_, err = suite.Keeper.StartSurplusAuction(suite.Ctx, suite.ModAcc.Name, c("token1", 20), "token2") // lot, bid denom
	suite.NoError(err)

	// Fast forward the block time
	ctx := suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(types.DefaultMaxAuctionDuration).Add(1))

	// Close expired auctions
	err = suite.Keeper.CloseExpiredAuctions(ctx)
	suite.NoError(err)
}
