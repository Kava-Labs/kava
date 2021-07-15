package keeper_test

import (
	"testing"

	"github.com/kava-labs/kava/x/swap/testutil"
	"github.com/kava-labs/kava/x/swap/types"
	"github.com/kava-labs/kava/x/swap/types/mocks"

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

func (suite *keeperTestSuite) setupPool(reserves sdk.Coins, totalShares sdk.Int, depositor sdk.AccAddress) string {
	poolID := types.PoolIDFromCoins(reserves)
	suite.AddCoinsToModule(reserves)

	poolRecord := types.PoolRecord{
		PoolID:      poolID,
		ReservesA:   reserves[0],
		ReservesB:   reserves[1],
		TotalShares: totalShares,
	}
	suite.Keeper.SetPool(suite.Ctx, poolRecord)

	shareRecord := types.ShareRecord{
		Depositor:   depositor,
		PoolID:      poolID,
		SharesOwned: totalShares,
	}
	suite.Keeper.SetDepositorShares(suite.Ctx, shareRecord)

	return poolID
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

func (suite keeperTestSuite) TestParams_GetSwapFee() {
	keeper := suite.Keeper

	params := types.Params{
		SwapFee: sdk.MustNewDecFromStr("0.00333"),
	}
	keeper.SetParams(suite.Ctx, params)

	suite.Equal(keeper.GetSwapFee(suite.Ctx), params.SwapFee)
}

func (suite *keeperTestSuite) TestPool_Persistance() {
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
		sdk.NewCoin("usdx", sdk.NewInt(50e6)),
	)

	pool, err := types.NewDenominatedPool(reserves)
	suite.Nil(err)
	record := types.NewPoolRecordFromPool(pool)

	suite.Keeper.SetPool(suite.Ctx, record)

	savedRecord, ok := suite.Keeper.GetPool(suite.Ctx, record.PoolID)
	suite.True(ok)
	suite.Equal(record, savedRecord)

	savedShares, ok := suite.Keeper.GetPoolShares(suite.Ctx, record.PoolID)
	suite.True(ok)
	suite.Equal(record.TotalShares, savedShares)

	suite.Keeper.DeletePool(suite.Ctx, record.PoolID)
	deletedPool, ok := suite.Keeper.GetPool(suite.Ctx, record.PoolID)
	suite.False(ok)
	suite.Equal(deletedPool, types.PoolRecord{})
}

func (suite *keeperTestSuite) TestPool_PanicsWhenInvalid() {
	invalidRecord := types.NewPoolRecord(
		sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(100e6)), sdk.NewCoin("usdx", sdk.NewInt(100e6))),
		i(-1),
	)

	suite.Panics(func() {
		suite.Keeper.SetPool(suite.Ctx, invalidRecord)
	}, "expected set pool to panic with invalid record")
}

func (suite *keeperTestSuite) TestShare_Persistance() {
	poolID := "ukava/usdx"
	depositor := sdk.AccAddress("testAddress1")
	shares := sdk.NewInt(3126432331)

	record := types.NewShareRecord(depositor, poolID, shares)
	suite.Keeper.SetDepositorShares(suite.Ctx, record)

	savedRecord, ok := suite.Keeper.GetDepositorShares(suite.Ctx, depositor, poolID)
	suite.True(ok)
	suite.Equal(record, savedRecord)

	savedShares, ok := suite.Keeper.GetDepositorSharesAmount(suite.Ctx, depositor, poolID)
	suite.True(ok)
	suite.Equal(record.SharesOwned, savedShares)

	suite.Keeper.DeleteDepositorShares(suite.Ctx, depositor, poolID)
	deletedShares, ok := suite.Keeper.GetDepositorShares(suite.Ctx, depositor, poolID)
	suite.False(ok)
	suite.Equal(deletedShares, types.ShareRecord{})
}

func (suite *keeperTestSuite) TestShare_PanicsWhenInvalid() {
	depositor, err := sdk.AccAddressFromBech32("kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w")
	suite.Require().NoError(err)

	invalidRecord := types.NewShareRecord(
		depositor,
		"hard/usdx",
		i(-1),
	)

	suite.Panics(func() {
		suite.Keeper.SetDepositorShares(suite.Ctx, invalidRecord)
	}, "expected set depositor shares to panic with invalid record")
}

func (suite *keeperTestSuite) TestHooks() {
	// ensure no hooks are set
	suite.Keeper.ClearHooks()

	// data
	depositor, err := sdk.AccAddressFromBech32("kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w")
	suite.Require().NoError(err)

	// hooks can be called when not set
	suite.Keeper.AfterPoolDepositCreated(suite.Ctx, "ukava/usdx", depositor, sdk.NewInt(1e6))
	suite.Keeper.BeforePoolDepositModified(suite.Ctx, "ukava/usdx", depositor, sdk.NewInt(1e6))

	// set hooks
	swapHooks := &mocks.SwapHooks{}
	suite.Keeper.SetHooks(swapHooks)

	// test hook calls are correct
	swapHooks.On("AfterPoolDepositCreated", suite.Ctx, "ukava/usdx", depositor, sdk.NewInt(1e6)).Once()
	suite.Keeper.AfterPoolDepositCreated(suite.Ctx, "ukava/usdx", depositor, sdk.NewInt(1e6))
	swapHooks.On("BeforePoolDepositModified", suite.Ctx, "ukava/usdx", depositor, sdk.NewInt(1e6)).Once()
	suite.Keeper.BeforePoolDepositModified(suite.Ctx, "ukava/usdx", depositor, sdk.NewInt(1e6))
	swapHooks.AssertExpectations(suite.T())

	// test second set panics
	suite.PanicsWithValue("cannot set swap hooks twice", func() {
		suite.Keeper.SetHooks(swapHooks)
	}, "expected hooks to panic on second set")
}
