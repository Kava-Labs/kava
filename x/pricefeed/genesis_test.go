package pricefeed_test

import (
	"testing"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/pricefeed"
	"github.com/kava-labs/kava/x/pricefeed/keeper"

	"github.com/stretchr/testify/suite"
)

type GenesisTestSuite struct {
	suite.Suite

	tApp   app.TestApp
	ctx    sdk.Context
	keeper keeper.Keeper
}

func (suite *GenesisTestSuite) SetupTest() {
	suite.tApp = app.NewTestApp()
	suite.ctx = suite.tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})
	suite.keeper = suite.tApp.GetPriceFeedKeeper()
}

func (suite *GenesisTestSuite) TestValidGenState() {
	suite.NotPanics(func() {
		suite.tApp.InitializeFromGenesisStates(
			NewPricefeedGenStateMulti(),
		)
	})
	_, addrs := app.GeneratePrivKeyAddressPairs(10)

	// Must create a new TestApp or InitChain will panic with index already set
	suite.tApp = app.NewTestApp()
	suite.NotPanics(func() {
		suite.tApp.InitializeFromGenesisStates(
			NewPricefeedGenStateWithOracles(addrs),
		)
	})
}

func (suite *GenesisTestSuite) TestInitExportGenState() {
	gs := NewPricefeedGen()

	suite.NotPanics(func() {
		pricefeed.InitGenesis(suite.ctx, suite.keeper, gs)
	})

	exportedGs := pricefeed.ExportGenesis(suite.ctx, suite.keeper)
	suite.NoError(gs.VerboseEqual(exportedGs), "exported genesis should match init genesis")
}

func (suite *GenesisTestSuite) TestParamPricesGenState() {
	gs := NewPricefeedGen()

	suite.NotPanics(func() {
		pricefeed.InitGenesis(suite.ctx, suite.keeper, gs)
	})

	params := suite.keeper.GetParams(suite.ctx)
	suite.NoError(gs.Params.VerboseEqual(params), "params should equal init params")

	pps := suite.keeper.GetRawPrices(suite.ctx, "btc:usd")
	suite.NoError(gs.PostedPrices[0].VerboseEqual(pps[0]), "posted prices should equal init posted prices")
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
