package keeper_test

import (
	"math/rand"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp/keeper"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"
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

func (suite *FeeTestSuite) TestCalculateFeesPrecisionLoss() {
	// Calculates the difference between fees calculated on the total amount of debt,
	// versus iterating over all the 1000 randomly generated cdps.
	// Assumes 7 second block times, runs simulations for 100, 1000, 10000, 100000, and 1000000
	// blocks, where the bulk debt is updated each block, and the cdps are updated once.
	coins := []sdk.Coins{}
	total := sdk.NewCoins()
	for i := 0; i < 1000; i++ {
		ri, err := simulation.RandPositiveInt(rand.New(rand.NewSource(int64(i))), sdk.NewInt(100000000000))
		suite.NoError(err)
		c := sdk.NewCoins(sdk.NewCoin("usdx", ri))
		coins = append(coins, c)
		total = total.Add(cs(sdk.NewCoin("usdx", ri)))
	}

	numBlocks := []int{100, 1000, 10000, 100000}

	for _, nb := range numBlocks {
		bulkFees := sdk.NewCoins()
		individualFees := sdk.NewCoins()
		for x := 0; x < nb; x++ {
			fee := suite.keeper.CalculateFees(suite.ctx, total.Add(bulkFees), i(7), "xrp")
			bulkFees = bulkFees.Add(fee)
		}

		for _, cns := range coins {
			fee := suite.keeper.CalculateFees(suite.ctx, cns, i(int64(nb*7)), "xrp")
			individualFees = individualFees.Add(fee)
		}

		absError := (sdk.OneDec().Sub(sdk.NewDecFromInt(bulkFees[0].Amount).Quo(sdk.NewDecFromInt(individualFees[0].Amount)))).Abs()

		suite.T().Log(bulkFees)
		suite.T().Log(individualFees)
		suite.T().Log(absError)

		suite.True(d("0.00001").GTE(absError))
	}

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
	err := suite.keeper.AddCdp(suite.ctx, addrs[0], cs(c("xrp", 200000000)), cs(c("usdx", 24000000)))
	suite.NoError(err) // check that no error was thrown

	// use the other account to create a cdp that SHOULD NOT have fees updated - 500% collateralization
	// create CDP for the second address
	err = suite.keeper.AddCdp(suite.ctx, addrs[1], cs(c("xrp", 200000000)), cs(c("usdx", 10000000)))
	suite.NoError(err) // check that no error was thrown

}

// UpdateFeesForRiskyCdpsTest tests the functionality for updating the fees for risky CDPs
func (suite *FeeTestSuite) TestUpdateFeesForRiskyCdps() {
	// this helper function creates two CDPs with id 1 and 2 respectively, each with zero fees
	suite.createCdps()

	cdpbefore, _ := suite.keeper.GetCDP(suite.ctx, "xrp", 1)
	// check fees
	suite.T().Log(cdpbefore)

	// move the context forward in time so that cdps will have fees accumulate if CalculateFees is called
	// note - time must be moved forward by a sufficient amount in order for additional
	// fees to accumulate, in this example 60 seconds
	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Second * 60))
	suite.keeper.UpdateFeesForRiskyCdps(suite.ctx, "xrp", "xrp:usd")

	// cdp we expect fees to accumulate for
	cdp1, _ := suite.keeper.GetCDP(suite.ctx, "xrp", 1)
	// check fees are not zero
	suite.T().Log(cdp1)
	// check that the fees have been updated
	suite.False(cdp1.AccumulatedFees.Empty())
	// now check that we have the correct amount of fees overall (2 USDX for this scenario)
	suite.Equal(sdk.NewInt(2), cdp1.AccumulatedFees.AmountOf("usdx"))

	// cdp we expect fees to not accumulate for
	cdp2, _ := suite.keeper.GetCDP(suite.ctx, "xrp", 2)
	suite.T().Log(cdp2)

	// check fees are zero
	suite.True(cdp2.AccumulatedFees.Empty())
}

func (suite *FeeTestSuite) TestGetSetPreviousBlockTime() {
	now := tmtime.Now()

	_, f := suite.keeper.GetPreviousBlockTime(suite.ctx)
	suite.False(f)

	suite.NotPanics(func() { suite.keeper.SetPreviousBlockTime(suite.ctx, now) })

	bpt, f := suite.keeper.GetPreviousBlockTime(suite.ctx)
	suite.True(f)
	suite.Equal(now, bpt)

}

func TestFeeTestSuite(t *testing.T) {
	suite.Run(t, new(FeeTestSuite))
}
