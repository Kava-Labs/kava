package keeper_test

import (
	"testing"

	"github.com/kava-labs/kava/x/earn/testutil"
	"github.com/kava-labs/kava/x/earn/types"
	"github.com/kava-labs/kava/x/earn/types/mocks"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type hookTestSuite struct {
	testutil.Suite
}

func (suite *hookTestSuite) SetupTest() {
	suite.Suite.SetupTest()
	suite.Keeper.SetParams(suite.Ctx, types.DefaultParams())
}

func TestHookTestSuite(t *testing.T) {
	suite.Run(t, new(hookTestSuite))
}

func (suite *hookTestSuite) TestHooks_DepositAndWithdraw() {
	suite.Keeper.ClearHooks()
	earnHooks := &mocks.EarnHooks{}
	suite.Keeper.SetHooks(earnHooks)

	vault1Denom := "usdx"
	vault2Denom := "ukava"
	deposit1Amount := sdk.NewInt64Coin(vault1Denom, 100)
	deposit2Amount := sdk.NewInt64Coin(vault2Denom, 100)

	suite.CreateVault(vault1Denom, types.StrategyTypes{types.STRATEGY_TYPE_HARD}, false, nil)
	suite.CreateVault(vault2Denom, types.StrategyTypes{types.STRATEGY_TYPE_SAVINGS}, false, nil)

	acc := suite.CreateAccount(sdk.NewCoins(
		sdk.NewInt64Coin(vault1Denom, 1000),
		sdk.NewInt64Coin(vault2Denom, 1000),
	), 0)

	// first deposit creates vault - calls AfterVaultDepositCreated with initial shares
	// shares are 1:1
	earnHooks.On(
		"AfterVaultDepositCreated",
		suite.Ctx,
		deposit1Amount.Denom,
		acc.GetAddress(),
		deposit1Amount.Amount.ToDec(),
	).Once()
	err := suite.Keeper.Deposit(
		suite.Ctx,
		acc.GetAddress(),
		deposit1Amount,
		types.STRATEGY_TYPE_HARD,
	)
	suite.Require().NoError(err)

	// second deposit adds to vault - calls BeforeVaultDepositModified
	// shares given are the initial shares, not new the shares added to the vault
	earnHooks.On(
		"BeforeVaultDepositModified",
		suite.Ctx,
		deposit1Amount.Denom,
		acc.GetAddress(),
		deposit1Amount.Amount.ToDec(),
	).Once()
	err = suite.Keeper.Deposit(
		suite.Ctx,
		acc.GetAddress(),
		deposit1Amount,
		types.STRATEGY_TYPE_HARD,
	)
	suite.Require().NoError(err)

	// get the shares from the store from the last deposit
	shareRecord, found := suite.Keeper.GetVaultAccountShares(
		suite.Ctx,
		acc.GetAddress(),
	)
	suite.Require().True(found)

	// third deposit adds to vault - calls BeforeVaultDepositModified
	// shares given are the shares added in previous deposit, not the shares added to the vault now
	earnHooks.On(
		"BeforeVaultDepositModified",
		suite.Ctx,
		deposit1Amount.Denom,
		acc.GetAddress(),
		shareRecord.AmountOf(deposit1Amount.Denom),
	).Once()
	err = suite.Keeper.Deposit(
		suite.Ctx,
		acc.GetAddress(),
		deposit1Amount,
		types.STRATEGY_TYPE_HARD,
	)
	suite.Require().NoError(err)

	// new deposit denom into vault creates the deposit and calls AfterVaultDepositCreated
	earnHooks.On(
		"AfterVaultDepositCreated",
		suite.Ctx,
		deposit2Amount.Denom,
		acc.GetAddress(),
		deposit2Amount.Amount.ToDec(),
	).Once()
	err = suite.Keeper.Deposit(
		suite.Ctx,
		acc.GetAddress(),
		deposit2Amount,
		types.STRATEGY_TYPE_SAVINGS,
	)
	suite.Require().NoError(err)

	// second deposit into vault calls BeforeVaultDepositModified with initial shares given
	earnHooks.On(
		"BeforeVaultDepositModified",
		suite.Ctx,
		deposit2Amount.Denom,
		acc.GetAddress(),
		deposit2Amount.Amount.ToDec(),
	).Once()
	err = suite.Keeper.Deposit(
		suite.Ctx,
		acc.GetAddress(),
		deposit2Amount,
		types.STRATEGY_TYPE_SAVINGS,
	)
	suite.Require().NoError(err)

	// get the shares from the store from the last deposit
	shareRecord, found = suite.Keeper.GetVaultAccountShares(
		suite.Ctx,
		acc.GetAddress(),
	)
	suite.Require().True(found)

	// third deposit into vault calls BeforeVaultDepositModified with shares from last deposit
	earnHooks.On(
		"BeforeVaultDepositModified",
		suite.Ctx,
		deposit2Amount.Denom,
		acc.GetAddress(),
		shareRecord.AmountOf(deposit2Amount.Denom),
	).Once()
	err = suite.Keeper.Deposit(
		suite.Ctx,
		acc.GetAddress(),
		deposit2Amount,
		types.STRATEGY_TYPE_SAVINGS,
	)
	suite.Require().NoError(err)

	// ------------------------------------------------------------
	// test hooks with a full withdraw of all shares deposit 1 denom
	shareRecord, found = suite.Keeper.GetVaultAccountShares(
		suite.Ctx,
		acc.GetAddress(),
	)
	suite.Require().True(found)

	// all shares given to BeforeVaultDepositModified
	earnHooks.On(
		"BeforeVaultDepositModified",
		suite.Ctx,
		deposit1Amount.Denom,
		acc.GetAddress(),
		shareRecord.AmountOf(deposit1Amount.Denom),
	).Once()
	err = suite.Keeper.Withdraw(
		suite.Ctx,
		acc.GetAddress(),
		// 3 deposits, multiply original deposit amount by 3
		sdk.NewCoin(deposit1Amount.Denom, deposit1Amount.Amount.MulRaw(3)),
		types.STRATEGY_TYPE_HARD,
	)
	suite.Require().NoError(err)

	// test hooks on partial withdraw
	shareRecord, found = suite.Keeper.GetVaultAccountShares(
		suite.Ctx,
		acc.GetAddress(),
	)
	suite.Require().True(found)

	// all shares given to before deposit modified even with partial withdraw
	earnHooks.On(
		"BeforeVaultDepositModified",
		suite.Ctx,
		deposit2Amount.Denom,
		acc.GetAddress(),
		shareRecord.AmountOf(deposit2Amount.Denom),
	).Once()
	err = suite.Keeper.Withdraw(
		suite.Ctx,
		acc.GetAddress(),
		deposit2Amount,
		types.STRATEGY_TYPE_SAVINGS,
	)
	suite.Require().NoError(err)

	// test hooks on second partial withdraw
	shareRecord, found = suite.Keeper.GetVaultAccountShares(
		suite.Ctx,
		acc.GetAddress(),
	)
	suite.Require().True(found)

	// all shares given to before deposit modified even with partial withdraw
	earnHooks.On(
		"BeforeVaultDepositModified",
		suite.Ctx,
		deposit2Amount.Denom,
		acc.GetAddress(),
		shareRecord.AmountOf(deposit2Amount.Denom),
	).Once()
	err = suite.Keeper.Withdraw(
		suite.Ctx,
		acc.GetAddress(),
		deposit2Amount,
		types.STRATEGY_TYPE_SAVINGS,
	)
	suite.Require().NoError(err)

	// test hooks withdraw all remaining shares
	shareRecord, found = suite.Keeper.GetVaultAccountShares(
		suite.Ctx,
		acc.GetAddress(),
	)
	suite.Require().True(found)

	// all shares given to before deposit modified even with partial withdraw
	earnHooks.On(
		"BeforeVaultDepositModified",
		suite.Ctx,
		deposit2Amount.Denom,
		acc.GetAddress(),
		shareRecord.AmountOf(deposit2Amount.Denom),
	).Once()
	err = suite.Keeper.Withdraw(
		suite.Ctx,
		acc.GetAddress(),
		deposit2Amount,
		types.STRATEGY_TYPE_SAVINGS,
	)
	suite.Require().NoError(err)

	earnHooks.AssertExpectations(suite.T())
}

func (suite *hookTestSuite) TestHooks_NoPanicsOnNilHooks() {
	suite.Keeper.ClearHooks()

	vaultDenom := "usdx"
	startBalance := sdk.NewInt64Coin(vaultDenom, 1000)
	depositAmount := sdk.NewInt64Coin(vaultDenom, 100)
	withdrawAmount := sdk.NewInt64Coin(vaultDenom, 100)

	suite.CreateVault(vaultDenom, types.StrategyTypes{types.STRATEGY_TYPE_HARD}, false, nil)

	acc := suite.CreateAccount(sdk.NewCoins(startBalance), 0)

	// AfterVaultDepositModified should not panic if no hooks are registered
	err := suite.Keeper.Deposit(suite.Ctx, acc.GetAddress(), depositAmount, types.STRATEGY_TYPE_HARD)
	suite.Require().NoError(err)

	// BeforeVaultDepositModified should not panic if no hooks are registered
	err = suite.Keeper.Deposit(suite.Ctx, acc.GetAddress(), depositAmount, types.STRATEGY_TYPE_HARD)
	suite.Require().NoError(err)

	// BeforeVaultDepositModified should not panic if no hooks are registered
	err = suite.Keeper.Withdraw(suite.Ctx, acc.GetAddress(), withdrawAmount, types.STRATEGY_TYPE_HARD)
	suite.Require().NoError(err)
}

func (suite *hookTestSuite) TestHooks_HookOrdering() {
	suite.Keeper.ClearHooks()
	earnHooks := &mocks.EarnHooks{}
	suite.Keeper.SetHooks(earnHooks)

	vaultDenom := "usdx"
	startBalance := sdk.NewInt64Coin(vaultDenom, 1000)
	depositAmount := sdk.NewInt64Coin(vaultDenom, 100)

	suite.CreateVault(vaultDenom, types.StrategyTypes{types.STRATEGY_TYPE_HARD}, false, nil)

	acc := suite.CreateAccount(sdk.NewCoins(startBalance), 0)

	earnHooks.On("AfterVaultDepositCreated", suite.Ctx, depositAmount.Denom, acc.GetAddress(), depositAmount.Amount.ToDec()).
		Run(func(args mock.Arguments) {
			shares, found := suite.Keeper.GetVaultAccountShares(suite.Ctx, acc.GetAddress())
			suite.Require().True(found, "expected after hook to be called after shares are updated")
			suite.Require().Equal(depositAmount.Amount.ToDec(), shares.AmountOf(depositAmount.Denom))
		})
	err := suite.Keeper.Deposit(suite.Ctx, acc.GetAddress(), depositAmount, types.STRATEGY_TYPE_HARD)
	suite.Require().NoError(err)

	earnHooks.On("BeforeVaultDepositModified", suite.Ctx, depositAmount.Denom, acc.GetAddress(), depositAmount.Amount.ToDec()).
		Run(func(args mock.Arguments) {
			shares, found := suite.Keeper.GetVaultAccountShares(suite.Ctx, acc.GetAddress())
			suite.Require().True(found, "expected after hook to be called after shares are updated")
			suite.Require().Equal(depositAmount.Amount.ToDec(), shares.AmountOf(depositAmount.Denom))
		})
	err = suite.Keeper.Deposit(suite.Ctx, acc.GetAddress(), depositAmount, types.STRATEGY_TYPE_HARD)
	suite.Require().NoError(err)

	existingShares, found := suite.Keeper.GetVaultAccountShares(suite.Ctx, acc.GetAddress())
	suite.Require().True(found)
	earnHooks.On("BeforeVaultDepositModified", suite.Ctx, depositAmount.Denom, acc.GetAddress(), existingShares.AmountOf(depositAmount.Denom)).
		Run(func(args mock.Arguments) {
			shares, found := suite.Keeper.GetVaultAccountShares(suite.Ctx, acc.GetAddress())
			suite.Require().True(found, "expected after hook to be called after shares are updated")
			suite.Require().Equal(depositAmount.Amount.MulRaw(2).ToDec(), shares.AmountOf(depositAmount.Denom))
		})
	err = suite.Keeper.Withdraw(suite.Ctx, acc.GetAddress(), depositAmount, types.STRATEGY_TYPE_HARD)
	suite.Require().NoError(err)
}
