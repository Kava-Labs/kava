package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp/keeper"
)

type FeeTestSuite struct {
	suite.Suite

	keeper keeper.Keeper
	app    app.TestApp
	ctx    sdk.Context
}

func (suite *FeeTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	tApp.InitializeFromGenesisStates(
		NewPricefeedGenStateMulti(),
		NewCDPGenStateMulti(),
	)
	keeper := tApp.GetCDPKeeper()
	suite.app = tApp
	suite.ctx = ctx
	suite.keeper = keeper
}

// createCdps is a helper function to create two CDPs each with zero fees
func (suite *FeeTestSuite) createCdps() {
	// create 2 accounts in the state and give them some coins
	// create two private key pair addresses
	_, addrs := app.GeneratePrivKeyAddressPairs(2)
	ak := suite.app.GetAccountKeeper()
	// setup the first account
	acc := ak.NewAccountWithAddress(suite.ctx, addrs[0])
	acc.SetCoins(cs(c("xrp", 200000000), c("btc", 500000000)))

	ak.SetAccount(suite.ctx, acc)
	// now setup the second account
	acc2 := ak.NewAccountWithAddress(suite.ctx, addrs[1])
	acc2.SetCoins(cs(c("xrp", 200000000), c("btc", 500000000)))
	ak.SetAccount(suite.ctx, acc2)

	// now create two cdps with the addresses we just created
	// use the created account to create a cdp that SHOULD have fees updated
	// to get a ratio between 100 - 110% of liquidation ratio we can use 200xrp ($50) and 24 usdx (208% collateralization with liquidation ratio of 200%)
	// create CDP for the first address
	err := suite.keeper.AddCdp(suite.ctx, addrs[0], c("xrp", 200000000), c("usdx", 24000000))
	suite.NoError(err) // check that no error was thrown

	// use the other account to create a cdp that SHOULD NOT have fees updated - 500% collateralization
	// create CDP for the second address
	err = suite.keeper.AddCdp(suite.ctx, addrs[1], c("xrp", 200000000), c("usdx", 10000000))
	suite.NoError(err) // check that no error was thrown

}

// TestUpdateFees tests the functionality for updating the fees for CDPs
func (suite *FeeTestSuite) TestUpdateFees() {
	// this helper function creates two CDPs with id 1 and 2 respectively, each with zero fees
	suite.createCdps()

	// move the context forward in time so that cdps will have fees accumulate if CalculateFees is called
	// note - time must be moved forward by a sufficient amount in order for additional
	// fees to accumulate, in this example 600 seconds
	oldtime := suite.ctx.BlockTime()
	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Second * 600))
	err := suite.keeper.UpdateFeesForAllCdps(suite.ctx, "xrp")
	suite.NoError(err) // check that we don't have any error

	// cdp we expect fees to accumulate for
	cdp1, found := suite.keeper.GetCDP(suite.ctx, "xrp", 1)
	suite.True(found)
	// check fees are not zero
	// check that the fees have been updated
	suite.False(cdp1.AccumulatedFees.IsZero())
	// now check that we have the correct amount of fees overall (22 USDX for this scenario)
	suite.Equal(sdk.NewInt(22), cdp1.AccumulatedFees.Amount)
	suite.Equal(suite.ctx.BlockTime(), cdp1.FeesUpdated)
	// cdp we expect fees to not accumulate for because of rounding to zero
	cdp2, found := suite.keeper.GetCDP(suite.ctx, "xrp", 2)
	suite.True(found)
	// check fees are zero
	suite.True(cdp2.AccumulatedFees.IsZero())
	suite.Equal(oldtime, cdp2.FeesUpdated)
}

func TestFeeTestSuite(t *testing.T) {
	suite.Run(t, new(FeeTestSuite))
}
