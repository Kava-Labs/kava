package keeper_test

import (
	"testing"

	"github.com/kava-labs/kava/x/swap/testutil"
	"github.com/kava-labs/kava/x/swap/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

type keeperTestSuite struct {
	testutil.Suite
}

func (suite *keeperTestSuite) SetupTest() {
	suite.Suite.SetupTest()
	suite.Keeper.SetParams(suite.Ctx, types.DefaultParams())
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(keeperTestSuite))
}

func (suite keeperTestSuite) TestParams_Persistance() {
	keeper := suite.Keeper

	params := types.Params{
		AllowedPools: types.AllowedPools{
			types.NewAllowedPool("ukava", "usdx"),
		},
		SwapFee: sdk.MustNewDecFromStr("0.03"),
	}
	keeper.SetParams(suite.Ctx, params)
	suite.Equal(keeper.GetParams(suite.Ctx), params)

	oldParams := params
	params = types.Params{
		AllowedPools: types.AllowedPools{
			types.NewAllowedPool("hard", "ukava"),
		},
		SwapFee: sdk.MustNewDecFromStr("0.01"),
	}
	keeper.SetParams(suite.Ctx, params)
	suite.NotEqual(keeper.GetParams(suite.Ctx), oldParams)
	suite.Equal(keeper.GetParams(suite.Ctx), params)
}

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
