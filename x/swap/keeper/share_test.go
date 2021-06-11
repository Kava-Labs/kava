package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *keeperTestSuite) TestShare_Persistance() {
	poolName := "ukava/usdx"
	depositor := sdk.AccAddress("testAddress1")
	shares := sdk.NewInt(3126432331)
	suite.Keeper.SetDepositorShares(suite.Ctx, depositor, poolName, shares)

	savedShares, ok := suite.Keeper.GetDepositorShares(suite.Ctx, depositor, poolName)
	suite.True(ok)
	suite.Equal(shares, savedShares)

	suite.Keeper.DeleteDepositorShares(suite.Ctx, depositor, poolName)
	deletedShares, ok := suite.Keeper.GetDepositorShares(suite.Ctx, depositor, poolName)
	suite.False(ok)
	suite.Equal(deletedShares, sdk.ZeroInt())
}
