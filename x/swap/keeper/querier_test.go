package keeper_test

import (
	"testing"

	"github.com/kava-labs/kava/x/swap/keeper"
	"github.com/kava-labs/kava/x/swap/testutil"
	"github.com/kava-labs/kava/x/swap/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
)

type querierTestSuite struct {
	testutil.Suite
	querier sdk.Querier
}

func (suite *querierTestSuite) SetupTest() {
	suite.Suite.SetupTest()
	suite.App.InitializeFromGenesisStates(
		NewSwapGenStateMulti(),
	)
	suite.querier = keeper.NewQuerier(suite.Keeper)
}

func (suite *querierTestSuite) TestUnkownRequest() {
	ctx := suite.Ctx.WithIsCheckTx(false)
	bz, err := suite.querier(ctx, []string{"invalid-path"}, abci.RequestQuery{})
	suite.Nil(bz)
	suite.EqualError(err, "unknown request: unknown swap query endpoint")
}

func (suite *querierTestSuite) TestQueryParams() {
	ctx := suite.Ctx.WithIsCheckTx(false)
	bz, err := suite.querier(ctx, []string{types.QueryGetParams}, abci.RequestQuery{})
	suite.Nil(err)
	suite.NotNil(bz)

	var p types.Params
	suite.Nil(types.ModuleCdc.UnmarshalJSON(bz, &p))

	swapGenesisState := NewSwapGenStateMulti()
	gs := types.GenesisState{}
	err = types.ModuleCdc.UnmarshalJSON(swapGenesisState["swap"], &gs)
	suite.Require().NoError(err)

	suite.Equal(gs.Params, p)
}

func TestQuerierTestSuite(t *testing.T) {
	suite.Run(t, new(querierTestSuite))
}
