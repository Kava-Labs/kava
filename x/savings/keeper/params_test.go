package keeper_test

import (
	"github.com/kava-labs/kava/x/savings/types"
)

func (suite *KeeperTestSuite) TestGetSetParams_Empty() {
	suite.Run("empty", func() {

		newParams := types.NewParams(nil)
		suite.keeper.SetParams(suite.ctx, newParams)

		fetchedParams := suite.keeper.GetParams(suite.ctx)
		suite.Require().Equal(newParams, fetchedParams)
	})
	suite.Run("non empty", func() {
		newParams := types.NewParams([]string{"btc", "test"})
		suite.keeper.SetParams(suite.ctx, newParams)

		fetchedParams := suite.keeper.GetParams(suite.ctx)
		suite.Require().Equal(newParams, fetchedParams)
	})
}
