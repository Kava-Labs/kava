package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp/keeper"
)

type KeeperTestSuite struct {
	suite.Suite

	keeper keeper.Keeper
	app    app.TestApp
	ctx    sdk.Context
}

func (suite *KeeperTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
	suite.ResetChain()
	return
}

func (suite *KeeperTestSuite) ResetChain() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	keeper := tApp.GetCDPKeeper()

	suite.app = tApp
	suite.ctx = ctx
	suite.keeper = keeper
}

func (suite *KeeperTestSuite) TestGetSetSavingsRateDistributed() {
	suite.ResetChain()

	// Set savings rate distributed value
	savingsRateDist := sdk.NewInt(555000555000)
	suite.keeper.SetSavingsRateDistributed(suite.ctx, savingsRateDist)

	// Check store's savings rate distributed value
	s := suite.keeper.GetSavingsRateDistributed(suite.ctx)
	suite.Equal(savingsRateDist, s)
}
