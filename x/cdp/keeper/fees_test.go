package keeper_test

import (
	"math/rand"
	"testing"

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

	numBlocks := []int{100, 1000, 10000, 100000, 1000000}

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

func TestFeeTestSuite(t *testing.T) {
	suite.Run(t, new(FeeTestSuite))
}
