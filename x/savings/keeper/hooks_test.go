package keeper_test

import (
	"github.com/kava-labs/kava/x/savings/types"
	"github.com/kava-labs/kava/x/savings/types/mocks"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/mock"
)

func (suite *KeeperTestSuite) TestHooks_DepositAndWithdraw() {
	suite.keeper.ClearHooks()
	savingsHooks := &mocks.SavingsHooks{}
	suite.keeper.SetHooks(savingsHooks)

	denom0 := "ukava"
	denom1 := "usdx"
	suite.keeper.SetParams(suite.ctx, types.NewParams([]string{denom0, denom1}))

	balance := sdk.NewCoins(
		sdk.NewCoin(denom0, sdk.NewInt(1000e6)),
		sdk.NewCoin(denom1, sdk.NewInt(1000e6)),
	)
	depositor_1 := suite.CreateAccount(balance)

	deposit0 := sdk.NewCoin(denom0, sdk.NewInt(10e6))
	deposit1 := sdk.NewCoin(denom1, sdk.NewInt(50e6))
	deposit2 := sdk.NewCoin(denom0, sdk.NewInt(5e6))

	// first deposit creates deposit - calls AfterSavingsDepositCreated with initial shares
	savingsHooks.On(
		"AfterSavingsDepositCreated",
		suite.ctx,
		depositor_1.GetAddress(),
		cs(deposit0),
	).Once()
	err := suite.keeper.Deposit(suite.ctx, depositor_1.GetAddress(), cs(deposit0))
	suite.Require().NoError(err)

	// second deposit adds to pool - calls BeforeSavingsDepositModified
	// shares given are the initial shares, not the shares added to the pool
	savingsHooks.On(
		"BeforeSavingsDepositModified",
		suite.ctx,
		depositor_1.GetAddress(),
		cs(deposit0),
		[]string{deposit1.Denom},
	).Once()
	err = suite.keeper.Deposit(
		suite.ctx,
		depositor_1.GetAddress(),
		cs(deposit1),
	)
	suite.Require().NoError(err)

	// get the shares from the store from the last deposit
	accDeposit, found := suite.keeper.GetDeposit(
		suite.ctx,
		depositor_1.GetAddress(),
	)
	suite.Require().True(found)

	// third deposit adds to pool - calls BeforeSavingsDepositModified
	// shares given are the shares added in previous deposit, not the shares added to the pool now
	savingsHooks.On(
		"BeforeSavingsDepositModified",
		suite.ctx,
		depositor_1.GetAddress(),
		cs(deposit0, deposit1),
		[]string(nil), // no new denoms
	).Once()
	err = suite.keeper.Deposit(
		suite.ctx,
		depositor_1.GetAddress(),
		cs(deposit2),
	)
	suite.Require().NoError(err)

	depositor_2 := suite.CreateAccountWithAddress(
		sdk.AccAddress("depositor 2---------"),
		sdk.NewCoins(
			sdk.NewCoin("ukava", sdk.NewInt(100e6)),
			sdk.NewCoin("usdx", sdk.NewInt(100e6)),
		),
	)

	// first deposit deposit into pool creates the deposit and calls AfterSavingsDepositCreated
	savingsHooks.On(
		"AfterSavingsDepositCreated",
		suite.ctx,
		depositor_2.GetAddress(),
		cs(deposit0),
	).Once()
	err = suite.keeper.Deposit(
		suite.ctx,
		depositor_2.GetAddress(),
		cs(deposit0),
	)
	suite.Require().NoError(err)

	// second deposit into pool calls BeforeSavingsDepositModified with initial shares given
	savingsHooks.On(
		"BeforeSavingsDepositModified",
		suite.ctx,
		depositor_2.GetAddress(),
		cs(deposit0),
		[]string{deposit1.Denom},
	).Once()
	err = suite.keeper.Deposit(
		suite.ctx,
		depositor_2.GetAddress(),
		cs(deposit1),
	)
	suite.Require().NoError(err)

	// get the shares from the store from the last deposit
	accDeposit, found = suite.keeper.GetDeposit(suite.ctx, depositor_2.GetAddress())
	suite.Require().True(found)

	// third deposit into pool calls BeforeSavingsDepositModified with shares from last deposit
	savingsHooks.On(
		"BeforeSavingsDepositModified",
		suite.ctx,
		accDeposit.Depositor,
		accDeposit.Amount,
		[]string(nil),
	).Once()
	err = suite.keeper.Deposit(
		suite.ctx,
		depositor_2.GetAddress(),
		cs(deposit2),
	)
	suite.Require().NoError(err)

	// test hooks with a full withdraw of all shares
	accDeposit, found = suite.keeper.GetDeposit(suite.ctx, depositor_1.GetAddress())
	suite.Require().True(found)
	// all shares given to BeforeSavingsDepositModified
	savingsHooks.On(
		"BeforeSavingsDepositModified",
		suite.ctx,
		accDeposit.Depositor,
		accDeposit.Amount,
		[]string(nil),
	).Once()
	err = suite.keeper.Withdraw(
		suite.ctx,
		depositor_1.GetAddress(),
		accDeposit.Amount,
	)
	suite.Require().NoError(err)

	// test hooks on partial withdraw
	accDeposit, found = suite.keeper.GetDeposit(suite.ctx, depositor_2.GetAddress())
	suite.Require().True(found)

	partialShares := accDeposit.Amount.AmountOf(denom0).Quo(sdk.NewInt(3))
	partialWithdraw1 := sdk.NewCoin(denom0, partialShares)
	// all shares given to before deposit modified even with partial withdraw
	savingsHooks.On(
		"BeforeSavingsDepositModified",
		suite.ctx,
		accDeposit.Depositor,
		accDeposit.Amount,
		[]string(nil),
	).Once()
	err = suite.keeper.Withdraw(suite.ctx, depositor_2.GetAddress(), cs(partialWithdraw1))
	suite.Require().NoError(err)

	// test hooks on second partial withdraw
	accDeposit, found = suite.keeper.GetDeposit(suite.ctx, depositor_2.GetAddress())
	suite.Require().True(found)
	partialShares = accDeposit.Amount.AmountOf(denom0).Quo(sdk.NewInt(2))
	partialWithdraw2 := sdk.NewCoin(denom0, partialShares)
	// all shares given to before deposit modified even with partial withdraw
	savingsHooks.On(
		"BeforeSavingsDepositModified",
		suite.ctx,
		accDeposit.Depositor,
		accDeposit.Amount,
		[]string(nil),
	).Once()
	err = suite.keeper.Withdraw(
		suite.ctx,
		depositor_2.GetAddress(),
		cs(partialWithdraw2),
	)
	suite.Require().NoError(err)

	// test hooks withdraw all shares with second depositor
	accDeposit, found = suite.keeper.GetDeposit(suite.ctx, depositor_2.GetAddress())
	suite.Require().True(found)
	// all shares given to before deposit modified even with partial withdraw
	savingsHooks.On(
		"BeforeSavingsDepositModified",
		suite.ctx,
		accDeposit.Depositor,
		accDeposit.Amount,
		[]string(nil),
	).Once()
	err = suite.keeper.Withdraw(
		suite.ctx,
		depositor_2.GetAddress(),
		accDeposit.Amount,
	)
	suite.Require().NoError(err)

	savingsHooks.AssertExpectations(suite.T())
}

func (suite *KeeperTestSuite) TestHooks_NoPanicsOnNilHooks() {
	suite.keeper.ClearHooks()

	denom0 := "ukava"
	denom1 := "usdx"
	suite.keeper.SetParams(suite.ctx, types.NewParams([]string{denom0, denom1}))

	balance := sdk.NewCoins(
		sdk.NewCoin(denom0, sdk.NewInt(1000e6)),
		sdk.NewCoin(denom1, sdk.NewInt(1000e6)),
	)
	depositor := suite.CreateAccount(balance)

	depositA := sdk.NewCoin(denom0, sdk.NewInt(10e6))
	depositB := sdk.NewCoin(denom1, sdk.NewInt(50e6))

	// deposit create pool should not panic when hooks are not set
	err := suite.keeper.Deposit(
		suite.ctx,
		depositor.GetAddress(),
		cs(depositA),
	)
	suite.Require().NoError(err)

	// existing deposit should not panic with hooks are not set
	err = suite.keeper.Deposit(
		suite.ctx,
		depositor.GetAddress(),
		cs(depositB),
	)
	suite.Require().NoError(err)

	// withdraw of shares should not panic when hooks are not set
	accDeposit, found := suite.keeper.GetDeposit(
		suite.ctx,
		depositor.GetAddress(),
	)
	suite.Require().True(found)
	err = suite.keeper.Withdraw(
		suite.ctx,
		depositor.GetAddress(),
		accDeposit.Amount,
	)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestHooks_HookOrdering() {
	suite.keeper.ClearHooks()
	savingsHooks := &mocks.SavingsHooks{}
	suite.keeper.SetHooks(savingsHooks)

	denom0 := "ukava"
	denom1 := "usdx"
	suite.keeper.SetParams(suite.ctx, types.NewParams([]string{denom0, denom1}))

	balance := sdk.NewCoins(
		sdk.NewCoin(denom0, sdk.NewInt(1000e6)),
		sdk.NewCoin(denom1, sdk.NewInt(1000e6)),
	)
	depositor := suite.CreateAccount(balance)

	depositA := sdk.NewCoin(denom0, sdk.NewInt(10e6))
	depositB := sdk.NewCoin(denom1, sdk.NewInt(50e6))

	savingsHooks.On(
		"AfterSavingsDepositCreated",
		suite.ctx,
		depositor.GetAddress(),
		cs(depositA),
	).Run(func(args mock.Arguments) {
		_, found := suite.keeper.GetDeposit(suite.ctx, depositor.GetAddress())
		suite.Require().True(found, "expected after hook to be called after shares are updated")
	})
	err := suite.keeper.Deposit(
		suite.ctx,
		depositor.GetAddress(),
		cs(depositA),
	)
	suite.Require().NoError(err)

	savingsHooks.On(
		"BeforeSavingsDepositModified",
		suite.ctx,
		depositor.GetAddress(),
		cs(depositA),
		[]string{depositB.Denom},
	).Run(func(args mock.Arguments) {
		accDeposit, found := suite.keeper.GetDeposit(suite.ctx, depositor.GetAddress())
		suite.Require().True(found, "expected share record to exist")
		suite.Equal(depositA.Amount, accDeposit.Amount.AmountOf(denom0), "expected hook to be called before shares are updated")
	})
	err = suite.keeper.Deposit(suite.ctx, depositor.GetAddress(), cs(depositB))
	suite.Require().NoError(err)

	existingaccDeposit, found := suite.keeper.GetDeposit(suite.ctx, depositor.GetAddress())
	suite.Require().True(found)
	savingsHooks.On(
		"BeforeSavingsDepositModified",
		suite.ctx,
		depositor.GetAddress(),
		cs(depositA, depositB),
		[]string(nil),
	).Run(func(args mock.Arguments) {
		accDeposit, found := suite.keeper.GetDeposit(suite.ctx, depositor.GetAddress())
		suite.Require().True(found, "expected share record to exist")
		suite.Equal(existingaccDeposit.Amount, accDeposit.Amount, "expected hook to be called before shares are updated")
	})
	err = suite.keeper.Withdraw(suite.ctx, depositor.GetAddress(), cs(depositA))
	suite.Require().NoError(err)
}
