package keeper_test

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/x/kavadist/keeper"
	"github.com/kava-labs/kava/x/kavadist/types"
)

func (suite *KeeperTestSuite) TestQuerier() {
	querier := keeper.NewQuerier(suite.keeper)
	bz, err := querier(suite.ctx, []string{types.QueryGetParams}, abci.RequestQuery{})
	suite.Require().NoError(err)
	suite.NotNil(bz)

	testParams := types.NewParams(true, testPeriods)
	var p types.Params
	suite.Nil(types.ModuleCdc.UnmarshalJSON(bz, &p))
	suite.Require().Equal(testParams, p)
}
