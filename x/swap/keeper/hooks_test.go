package keeper_test

import (
	"github.com/kava-labs/kava/x/swap/types"
	"github.com/kava-labs/kava/x/swap/types/mocks"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/mock"
)

func (suite *keeperTestSuite) TestHooks_DepositAndWithdraw() {
	suite.Keeper.ClearHooks()
	swapHooks := &mocks.SwapHooks{}
	suite.Keeper.SetHooks(swapHooks)

	pool := types.NewAllowedPool("ukava", "usdx")
	suite.Require().NoError(pool.Validate())
	suite.Keeper.SetParams(suite.Ctx, types.NewParams(types.NewAllowedPools(pool), types.DefaultSwapFee))

	balance := sdk.NewCoins(
		sdk.NewCoin(pool.TokenA, sdk.NewInt(1000e6)),
		sdk.NewCoin(pool.TokenB, sdk.NewInt(1000e6)),
	)
	depositor_1 := suite.CreateAccount(balance)

	depositA := sdk.NewCoin(pool.TokenA, sdk.NewInt(10e6))
	depositB := sdk.NewCoin(pool.TokenB, sdk.NewInt(50e6))
	deposit := sdk.NewCoins(depositA, depositB)

	// expected initial shares - geometric mean
	expectedShares := sdk.NewInt(22360679)

	// first deposit creates pool - calls AfterPoolDepositCreated with initial shares
	swapHooks.On("AfterPoolDepositCreated", suite.Ctx, types.PoolIDFromCoins(deposit), depositor_1.GetAddress(), expectedShares).Once()
	err := suite.Keeper.Deposit(suite.Ctx, depositor_1.GetAddress(), depositA, depositB, sdk.MustNewDecFromStr("0.0015"))
	suite.Require().NoError(err)

	// second deposit adds to pool - calls BeforePoolDepositModified
	// shares given are the initial shares, not the shares added to the pool
	swapHooks.On("BeforePoolDepositModified", suite.Ctx, types.PoolIDFromCoins(deposit), depositor_1.GetAddress(), expectedShares).Once()
	err = suite.Keeper.Deposit(suite.Ctx, depositor_1.GetAddress(), sdk.NewCoin("ukava", sdk.NewInt(5e6)), sdk.NewCoin("usdx", sdk.NewInt(25e6)), sdk.MustNewDecFromStr("0.0015"))
	suite.Require().NoError(err)

	// get the shares from the store from the last deposit
	shareRecord, found := suite.Keeper.GetDepositorShares(suite.Ctx, depositor_1.GetAddress(), types.PoolIDFromCoins(deposit))
	suite.Require().True(found)

	// third deposit adds to pool - calls BeforePoolDepositModified
	// shares given are the shares added in previous deposit, not the shares added to the pool now
	swapHooks.On("BeforePoolDepositModified", suite.Ctx, types.PoolIDFromCoins(deposit), depositor_1.GetAddress(), shareRecord.SharesOwned).Once()
	err = suite.Keeper.Deposit(suite.Ctx, depositor_1.GetAddress(), sdk.NewCoin("ukava", sdk.NewInt(10e6)), sdk.NewCoin("usdx", sdk.NewInt(50e6)), sdk.MustNewDecFromStr("0.0015"))
	suite.Require().NoError(err)

	depositor_2 := suite.NewAccountFromAddr(
		sdk.AccAddress("depositor 2"),
		sdk.NewCoins(
			sdk.NewCoin("ukava", sdk.NewInt(100e6)),
			sdk.NewCoin("usdx", sdk.NewInt(100e6)),
		),
	)

	// first deposit deposit into pool creates the deposit and calls AfterPoolDepositCreated
	expectedShares = sdk.NewInt(2236067)
	swapHooks.On("AfterPoolDepositCreated", suite.Ctx, types.PoolIDFromCoins(deposit), depositor_2.GetAddress(), expectedShares).Once()
	err = suite.Keeper.Deposit(suite.Ctx, depositor_2.GetAddress(), sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.MustNewDecFromStr("0.0015"))
	suite.Require().NoError(err)

	// second deposit into pool calls BeforePoolDepositModified with initial shares given
	swapHooks.On("BeforePoolDepositModified", suite.Ctx, types.PoolIDFromCoins(deposit), depositor_2.GetAddress(), expectedShares).Once()
	err = suite.Keeper.Deposit(suite.Ctx, depositor_2.GetAddress(), sdk.NewCoin("ukava", sdk.NewInt(2e6)), sdk.NewCoin("usdx", sdk.NewInt(10e6)), sdk.MustNewDecFromStr("0.0015"))
	suite.Require().NoError(err)

	// get the shares from the store from the last deposit
	shareRecord, found = suite.Keeper.GetDepositorShares(suite.Ctx, depositor_2.GetAddress(), types.PoolIDFromCoins(deposit))
	suite.Require().True(found)

	// third deposit into pool calls BeforePoolDepositModified with shares from last deposit
	swapHooks.On("BeforePoolDepositModified", suite.Ctx, types.PoolIDFromCoins(deposit), depositor_2.GetAddress(), shareRecord.SharesOwned).Once()
	err = suite.Keeper.Deposit(suite.Ctx, depositor_2.GetAddress(), sdk.NewCoin("ukava", sdk.NewInt(3e6)), sdk.NewCoin("usdx", sdk.NewInt(15e6)), sdk.MustNewDecFromStr("0.0015"))
	suite.Require().NoError(err)

	// test hooks with a full withdraw of all shares
	shareRecord, found = suite.Keeper.GetDepositorShares(suite.Ctx, depositor_1.GetAddress(), types.PoolIDFromCoins(deposit))
	suite.Require().True(found)
	// all shares given to BeforePoolDepositModified
	swapHooks.On("BeforePoolDepositModified", suite.Ctx, types.PoolIDFromCoins(deposit), depositor_1.GetAddress(), shareRecord.SharesOwned).Once()
	err = suite.Keeper.Withdraw(suite.Ctx, depositor_1.GetAddress(), shareRecord.SharesOwned, sdk.NewCoin("ukava", sdk.NewInt(1)), sdk.NewCoin("usdx", sdk.NewInt(1)))
	suite.Require().NoError(err)

	// test hooks on partial withdraw
	shareRecord, found = suite.Keeper.GetDepositorShares(suite.Ctx, depositor_2.GetAddress(), types.PoolIDFromCoins(deposit))
	suite.Require().True(found)
	partialShares := shareRecord.SharesOwned.Quo(sdk.NewInt(3))
	// all shares given to before deposit modified even with partial withdraw
	swapHooks.On("BeforePoolDepositModified", suite.Ctx, types.PoolIDFromCoins(deposit), depositor_2.GetAddress(), shareRecord.SharesOwned).Once()
	err = suite.Keeper.Withdraw(suite.Ctx, depositor_2.GetAddress(), partialShares, sdk.NewCoin("ukava", sdk.NewInt(1)), sdk.NewCoin("usdx", sdk.NewInt(1)))
	suite.Require().NoError(err)

	// test hooks on second partial withdraw
	shareRecord, found = suite.Keeper.GetDepositorShares(suite.Ctx, depositor_2.GetAddress(), types.PoolIDFromCoins(deposit))
	suite.Require().True(found)
	partialShares = shareRecord.SharesOwned.Quo(sdk.NewInt(2))
	// all shares given to before deposit modified even with partial withdraw
	swapHooks.On("BeforePoolDepositModified", suite.Ctx, types.PoolIDFromCoins(deposit), depositor_2.GetAddress(), shareRecord.SharesOwned).Once()
	err = suite.Keeper.Withdraw(suite.Ctx, depositor_2.GetAddress(), partialShares, sdk.NewCoin("ukava", sdk.NewInt(1)), sdk.NewCoin("usdx", sdk.NewInt(1)))
	suite.Require().NoError(err)

	// test hooks withdraw all shares with second depositor
	shareRecord, found = suite.Keeper.GetDepositorShares(suite.Ctx, depositor_2.GetAddress(), types.PoolIDFromCoins(deposit))
	suite.Require().True(found)
	// all shares given to before deposit modified even with partial withdraw
	swapHooks.On("BeforePoolDepositModified", suite.Ctx, types.PoolIDFromCoins(deposit), depositor_2.GetAddress(), shareRecord.SharesOwned).Once()
	err = suite.Keeper.Withdraw(suite.Ctx, depositor_2.GetAddress(), shareRecord.SharesOwned, sdk.NewCoin("ukava", sdk.NewInt(1)), sdk.NewCoin("usdx", sdk.NewInt(1)))
	suite.Require().NoError(err)

	swapHooks.AssertExpectations(suite.T())
}

func (suite *keeperTestSuite) TestHooks_NoPanicsOnNilHooks() {
	suite.Keeper.ClearHooks()

	pool := types.NewAllowedPool("ukava", "usdx")
	suite.Require().NoError(pool.Validate())
	suite.Keeper.SetParams(suite.Ctx, types.NewParams(types.NewAllowedPools(pool), types.DefaultSwapFee))

	balance := sdk.NewCoins(
		sdk.NewCoin(pool.TokenA, sdk.NewInt(1000e6)),
		sdk.NewCoin(pool.TokenB, sdk.NewInt(1000e6)),
	)
	depositor := suite.CreateAccount(balance)

	depositA := sdk.NewCoin(pool.TokenA, sdk.NewInt(10e6))
	depositB := sdk.NewCoin(pool.TokenB, sdk.NewInt(50e6))
	deposit := sdk.NewCoins(depositA, depositB)

	// deposit create pool should not panic when hooks are not set
	err := suite.Keeper.Deposit(suite.Ctx, depositor.GetAddress(), depositA, depositB, sdk.MustNewDecFromStr("0.0015"))
	suite.Require().NoError(err)

	// existing deposit should not panic with hooks are not set
	err = suite.Keeper.Deposit(suite.Ctx, depositor.GetAddress(), sdk.NewCoin("ukava", sdk.NewInt(5e6)), sdk.NewCoin("usdx", sdk.NewInt(25e6)), sdk.MustNewDecFromStr("0.0015"))
	suite.Require().NoError(err)

	// withdraw of shares should not panic when hooks are not set
	shareRecord, found := suite.Keeper.GetDepositorShares(suite.Ctx, depositor.GetAddress(), types.PoolIDFromCoins(deposit))
	suite.Require().True(found)
	err = suite.Keeper.Withdraw(suite.Ctx, depositor.GetAddress(), shareRecord.SharesOwned, sdk.NewCoin("ukava", sdk.NewInt(1)), sdk.NewCoin("usdx", sdk.NewInt(1)))
	suite.Require().NoError(err)
}

func (suite *keeperTestSuite) TestHooks_HookOrdering() {
	suite.Keeper.ClearHooks()
	swapHooks := &mocks.SwapHooks{}
	suite.Keeper.SetHooks(swapHooks)

	pool := types.NewAllowedPool("ukava", "usdx")
	suite.Require().NoError(pool.Validate())
	suite.Keeper.SetParams(suite.Ctx, types.NewParams(types.NewAllowedPools(pool), types.DefaultSwapFee))

	balance := sdk.NewCoins(
		sdk.NewCoin(pool.TokenA, sdk.NewInt(1000e6)),
		sdk.NewCoin(pool.TokenB, sdk.NewInt(1000e6)),
	)
	depositor := suite.CreateAccount(balance)

	depositA := sdk.NewCoin(pool.TokenA, sdk.NewInt(10e6))
	depositB := sdk.NewCoin(pool.TokenB, sdk.NewInt(50e6))
	deposit := sdk.NewCoins(depositA, depositB)

	poolID := types.PoolIDFromCoins(deposit)
	expectedShares := sdk.NewInt(22360679)

	swapHooks.On("AfterPoolDepositCreated", suite.Ctx, poolID, depositor.GetAddress(), expectedShares).Run(func(args mock.Arguments) {
		_, found := suite.Keeper.GetDepositorShares(suite.Ctx, depositor.GetAddress(), poolID)
		suite.Require().True(found, "expected after hook to be called after shares are updated")
	})
	err := suite.Keeper.Deposit(suite.Ctx, depositor.GetAddress(), depositA, depositB, sdk.MustNewDecFromStr("0.0015"))
	suite.Require().NoError(err)

	swapHooks.On("BeforePoolDepositModified", suite.Ctx, poolID, depositor.GetAddress(), expectedShares).Run(func(args mock.Arguments) {
		shareRecord, found := suite.Keeper.GetDepositorShares(suite.Ctx, depositor.GetAddress(), poolID)
		suite.Require().True(found, "expected share record to exist")
		suite.Equal(expectedShares, shareRecord.SharesOwned, "expected hook to be called before shares are updated")
	})
	err = suite.Keeper.Deposit(suite.Ctx, depositor.GetAddress(), depositA, depositB, sdk.MustNewDecFromStr("0.0015"))
	suite.Require().NoError(err)

	existingShareRecord, found := suite.Keeper.GetDepositorShares(suite.Ctx, depositor.GetAddress(), types.PoolIDFromCoins(deposit))
	suite.Require().True(found)
	swapHooks.On("BeforePoolDepositModified", suite.Ctx, poolID, depositor.GetAddress(), existingShareRecord.SharesOwned).Run(func(args mock.Arguments) {
		shareRecord, found := suite.Keeper.GetDepositorShares(suite.Ctx, depositor.GetAddress(), poolID)
		suite.Require().True(found, "expected share record to exist")
		suite.Equal(existingShareRecord.SharesOwned, shareRecord.SharesOwned, "expected hook to be called before shares are updated")
	})
	err = suite.Keeper.Withdraw(suite.Ctx, depositor.GetAddress(), existingShareRecord.SharesOwned.Quo(sdk.NewInt(2)), sdk.NewCoin("ukava", sdk.NewInt(1)), sdk.NewCoin("usdx", sdk.NewInt(1)))
	suite.Require().NoError(err)
}
