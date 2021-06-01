package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/swap/keeper"
	"github.com/kava-labs/kava/x/swap/types"
)

type QuerierTestSuite struct {
	suite.Suite
	keeper  keeper.Keeper
	app     app.TestApp
	ctx     sdk.Context
	querier sdk.Querier
}

func (suite *QuerierTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})

	tApp.InitializeFromGenesisStates(
		NewSwapGenStateMulti(),
	)

	suite.ctx = ctx
	suite.app = tApp
	suite.keeper = tApp.GetSwapKeeper()
	suite.querier = keeper.NewQuerier(suite.keeper)
}

func (suite *QuerierTestSuite) TestQueryParams() {
	ctx := suite.ctx.WithIsCheckTx(false)
	bz, err := suite.querier(ctx, []string{types.QueryGetParams}, abci.RequestQuery{})
	suite.Nil(err)
	suite.NotNil(bz)

	var p types.Params
	suite.Nil(types.ModuleCdc.UnmarshalJSON(bz, &p))

	swapGenesisState := NewSwapGenStateMulti()
	gs := types.GenesisState{}
	types.ModuleCdc.UnmarshalJSON(swapGenesisState["swap"], &gs)

	suite.Equal(gs.Params, p)
}

func TestQuerierTestSuite(t *testing.T) {
	suite.Run(t, new(QuerierTestSuite))
}
