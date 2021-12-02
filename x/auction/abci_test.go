package auction_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/auction"
	"github.com/kava-labs/kava/x/auction/testutil"
	types "github.com/kava-labs/kava/x/auction/types"
)

type abciTestSuite struct {
	testutil.Suite
}

func (suite *abciTestSuite) SetupTest() {
	suite.Suite.SetupTest(4)
}

func TestABCITestSuite(t *testing.T) {
	suite.Run(t, new(abciTestSuite))
}

func (suite *abciTestSuite) TestKeeper_BeginBlocker() {
	buyer := suite.Addrs[0]
	returnAddrs := []sdk.AccAddress{suite.Addrs[1]}
	returnWeights := []sdk.Int{sdk.NewInt(1)}

	suite.AddCoinsToNamedModule(suite.ModAcc.Name, cs(c("token1", 100), c("token2", 100), c("debt", 100)))

	// Start an auction and place a bid
	auctionID, err := suite.Keeper.StartCollateralAuction(suite.Ctx, suite.ModAcc.Name, c("token1", 20), c("token2", 50), returnAddrs, returnWeights, c("debt", 40))
	suite.Require().NoError(err)
	suite.Require().NoError(suite.Keeper.PlaceBid(suite.Ctx, auctionID, buyer, c("token2", 30)))

	// Run the beginblocker, simulating a block time 1ns before auction expiry
	preExpiryTime := suite.Ctx.BlockTime().Add(types.DefaultBidDuration - 1)
	auction.BeginBlocker(suite.Ctx.WithBlockTime(preExpiryTime), suite.Keeper)

	// Check auction has not been closed yet
	_, found := suite.Keeper.GetAuction(suite.Ctx, auctionID)
	suite.True(found)

	// Run the endblocker, simulating a block time equal to auction expiry
	expiryTime := suite.Ctx.BlockTime().Add(types.DefaultBidDuration)
	auction.BeginBlocker(suite.Ctx.WithBlockTime(expiryTime), suite.Keeper)

	// Check auction has been closed
	_, found = suite.Keeper.GetAuction(suite.Ctx, auctionID)
	suite.False(found)
}

func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }
