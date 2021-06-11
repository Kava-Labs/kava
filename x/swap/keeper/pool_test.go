package keeper_test

import (
	"github.com/kava-labs/kava/x/swap/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *keeperTestSuite) TestPool_Persistance() {
	reserveA := sdk.NewCoin("ukava", sdk.NewInt(10e6))
	reserveB := sdk.NewCoin("usdx", sdk.NewInt(50e6))

	pool, err := types.NewPool(reserveA, reserveB)
	suite.Nil(err)

	suite.Keeper.SetPool(suite.Ctx, pool)

	savedPool, ok := suite.Keeper.GetPool(suite.Ctx, pool.Name())
	suite.True(ok)
	suite.Equal(pool, savedPool)

	suite.Keeper.DeletePool(suite.Ctx, pool.Name())
	deletedPool, ok := suite.Keeper.GetPool(suite.Ctx, pool.Name())
	suite.False(ok)
	suite.Equal(deletedPool, types.Pool{})
}
