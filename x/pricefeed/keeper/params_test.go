package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/pricefeed/keeper"
)

type KeeperTestSuite struct {
	suite.Suite

	keeper keeper.Keeper
	addrs  []sdk.AccAddress
	app    app.TestApp
	ctx    sdk.Context
}

func (suite *KeeperTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	_, addrs := app.GeneratePrivKeyAddressPairs(10)
	tApp.InitializeFromGenesisStates(
		NewPricefeedGenStateMulti(),
	)
	suite.keeper = tApp.GetPriceFeedKeeper()
	suite.ctx = ctx
	suite.addrs = addrs
}

func (suite *KeeperTestSuite) TestGetSetOracles() {
	params := suite.keeper.GetParams(suite.ctx)
	suite.Equal([]sdk.AccAddress(nil), params.Markets[0].Oracles)
	params.Markets[0].Oracles = suite.addrs
	suite.NotPanics(func() { suite.keeper.SetParams(suite.ctx, params) })
	params = suite.keeper.GetParams(suite.ctx)
	suite.Equal(suite.addrs, params.Markets[0].Oracles)
	addr, err := suite.keeper.GetOracle(suite.ctx, params.Markets[0].MarketID, suite.addrs[0])
	suite.NoError(err)
	suite.Equal(suite.addrs[0], addr)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
