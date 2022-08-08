package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/earn/testutil"
	"github.com/kava-labs/kava/x/earn/types"

	"github.com/stretchr/testify/suite"
)

const savingsVaultDenom = "ukava"

type strategySavingsTestSuite struct {
	testutil.Suite
}

func (suite *strategySavingsTestSuite) SetupTest() {
	suite.Suite.SetupTest()
	suite.Keeper.SetParams(suite.Ctx, types.DefaultParams())
}

func TestStrategySavingsTestSuite(t *testing.T) {
	suite.Run(t, new(strategySavingsTestSuite))
}

func (suite *strategySavingsTestSuite) TestGetStrategyType() {
	strategy, err := suite.Keeper.GetStrategy(types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	suite.Equal(types.STRATEGY_TYPE_SAVINGS, strategy.GetStrategyType())
}

func (suite *strategySavingsTestSuite) TestDeposit_SingleAcc() {
	startBalance := sdk.NewInt64Coin(savingsVaultDenom, 1000)
	depositAmount := sdk.NewInt64Coin(savingsVaultDenom, 100)

	suite.CreateVault(savingsVaultDenom, types.StrategyTypes{types.STRATEGY_TYPE_SAVINGS}, false, nil)

	acc := suite.CreateAccount(sdk.NewCoins(startBalance), 0)

	err := suite.Keeper.Deposit(suite.Ctx, acc.GetAddress(), depositAmount, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	suite.SavingsDepositAmountEqual(sdk.NewCoins(depositAmount))
	suite.VaultTotalValuesEqual(sdk.NewCoins(depositAmount))
	suite.VaultTotalSharesEqual(types.NewVaultShares(
		types.NewVaultShare(depositAmount.Denom, depositAmount.Amount.ToDec()),
	))

	// Query vault total
	totalValue, err := suite.Keeper.GetVaultTotalValue(suite.Ctx, savingsVaultDenom)
	suite.Require().NoError(err)

	suite.Equal(depositAmount, totalValue)
}

func (suite *strategySavingsTestSuite) TestDeposit_SingleAcc_MultipleDeposits() {
	startBalance := sdk.NewInt64Coin(savingsVaultDenom, 1000)
	depositAmount := sdk.NewInt64Coin(savingsVaultDenom, 100)

	suite.CreateVault(savingsVaultDenom, types.StrategyTypes{types.STRATEGY_TYPE_SAVINGS}, false, nil)

	acc := suite.CreateAccount(sdk.NewCoins(startBalance), 0)

	err := suite.Keeper.Deposit(suite.Ctx, acc.GetAddress(), depositAmount, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	// Second deposit
	err = suite.Keeper.Deposit(suite.Ctx, acc.GetAddress(), depositAmount, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	expectedVaultBalance := depositAmount.Add(depositAmount)
	suite.SavingsDepositAmountEqual(sdk.NewCoins(expectedVaultBalance))
	suite.VaultTotalValuesEqual(sdk.NewCoins(expectedVaultBalance))
	suite.VaultTotalSharesEqual(types.NewVaultShares(
		types.NewVaultShare(expectedVaultBalance.Denom, expectedVaultBalance.Amount.ToDec()),
	))

	// Query vault total
	totalValue, err := suite.Keeper.GetVaultTotalValue(suite.Ctx, savingsVaultDenom)
	suite.Require().NoError(err)

	suite.Equal(depositAmount.Add(depositAmount), totalValue)
}

func (suite *strategySavingsTestSuite) TestDeposit_MultipleAcc_MultipleDeposits() {
	startBalance := sdk.NewInt64Coin(savingsVaultDenom, 1000)
	depositAmount := sdk.NewInt64Coin(savingsVaultDenom, 100)

	expectedTotalValue := sdk.NewCoin(savingsVaultDenom, depositAmount.Amount.MulRaw(4))

	suite.CreateVault(savingsVaultDenom, types.StrategyTypes{types.STRATEGY_TYPE_SAVINGS}, false, nil)

	acc1 := suite.CreateAccount(sdk.NewCoins(startBalance), 0)
	acc2 := suite.CreateAccount(sdk.NewCoins(startBalance), 0)

	// 2 deposits each account
	for i := 0; i < 2; i++ {
		// Deposit from acc1
		err := suite.Keeper.Deposit(suite.Ctx, acc1.GetAddress(), depositAmount, types.STRATEGY_TYPE_SAVINGS)
		suite.Require().NoError(err)

		// Deposit from acc2
		err = suite.Keeper.Deposit(suite.Ctx, acc2.GetAddress(), depositAmount, types.STRATEGY_TYPE_SAVINGS)
		suite.Require().NoError(err)
	}

	suite.SavingsDepositAmountEqual(sdk.NewCoins(expectedTotalValue))
	suite.VaultTotalValuesEqual(sdk.NewCoins(expectedTotalValue))
	suite.VaultTotalSharesEqual(types.NewVaultShares(
		types.NewVaultShare(expectedTotalValue.Denom, expectedTotalValue.Amount.ToDec()),
	))

	// Query vault total
	totalValue, err := suite.Keeper.GetVaultTotalValue(suite.Ctx, savingsVaultDenom)
	suite.Require().NoError(err)

	suite.Equal(expectedTotalValue, totalValue)
}

func (suite *strategySavingsTestSuite) TestGetVaultTotalValue_Empty() {
	suite.CreateVault(savingsVaultDenom, types.StrategyTypes{types.STRATEGY_TYPE_SAVINGS}, false, nil)

	// Query vault total
	totalValue, err := suite.Keeper.GetVaultTotalValue(suite.Ctx, savingsVaultDenom)
	suite.Require().NoError(err)

	suite.Equal(sdk.NewCoin(savingsVaultDenom, sdk.ZeroInt()), totalValue)
}

func (suite *strategySavingsTestSuite) TestGetVaultTotalValue_NoDenomDeposit() {
	// 2 Vaults usdx, busd
	// 1st vault has deposits
	// 2nd vault has no deposits

	vaultDenomBusd := "busd"

	suite.CreateVault(savingsVaultDenom, types.StrategyTypes{types.STRATEGY_TYPE_SAVINGS}, false, nil)
	suite.CreateVault(vaultDenomBusd, types.StrategyTypes{types.STRATEGY_TYPE_SAVINGS}, false, nil)

	startBalance := sdk.NewInt64Coin(savingsVaultDenom, 1000)
	depositAmount := sdk.NewInt64Coin(savingsVaultDenom, 100)

	acc := suite.CreateAccount(sdk.NewCoins(startBalance), 0)

	// Deposit vault1
	err := suite.Keeper.Deposit(suite.Ctx, acc.GetAddress(), depositAmount, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	// Query vault total, savings deposit exists for account, but amount in busd does not
	// Vault2 does not have any value, only returns amount for the correct denom
	// if a savings deposit already exists
	totalValueBusd, err := suite.Keeper.GetVaultTotalValue(suite.Ctx, vaultDenomBusd)
	suite.Require().NoError(err)

	suite.Equal(sdk.NewCoin(vaultDenomBusd, sdk.ZeroInt()), totalValueBusd)
}

// ----------------------------------------------------------------------------
// Withdraw

func (suite *strategySavingsTestSuite) TestWithdraw() {
	startBalance := sdk.NewInt64Coin(savingsVaultDenom, 1000)
	depositAmount := sdk.NewInt64Coin(savingsVaultDenom, 100)

	suite.CreateVault(savingsVaultDenom, types.StrategyTypes{types.STRATEGY_TYPE_SAVINGS}, false, nil)

	acc := suite.CreateAccount(sdk.NewCoins(startBalance), 0)
	err := suite.Keeper.Deposit(suite.Ctx, acc.GetAddress(), depositAmount, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	suite.SavingsDepositAmountEqual(sdk.NewCoins(depositAmount))

	// Query vault total
	totalValue, err := suite.Keeper.GetVaultTotalValue(suite.Ctx, savingsVaultDenom)
	suite.Require().NoError(err)
	suite.Equal(depositAmount, totalValue)

	// Withdraw
	err = suite.Keeper.Withdraw(suite.Ctx, acc.GetAddress(), depositAmount, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	suite.SavingsDepositAmountEqual(sdk.NewCoins())
	suite.VaultTotalValuesEqual(sdk.NewCoins())
	suite.VaultTotalSharesEqual(types.NewVaultShares())

	totalValue, err = suite.Keeper.GetVaultTotalValue(suite.Ctx, savingsVaultDenom)
	suite.Require().NoError(err)
	suite.Equal(sdk.NewInt64Coin(savingsVaultDenom, 0), totalValue)

	// Withdraw again
	err = suite.Keeper.Withdraw(suite.Ctx, acc.GetAddress(), depositAmount, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, types.ErrVaultRecordNotFound, "vault should be deleted when no more supply")
}

func (suite *strategySavingsTestSuite) TestWithdraw_OnlyWithdrawOwnSupply() {
	startBalance := sdk.NewInt64Coin(savingsVaultDenom, 1000)
	depositAmount := sdk.NewInt64Coin(savingsVaultDenom, 100)

	suite.CreateVault(savingsVaultDenom, types.StrategyTypes{types.STRATEGY_TYPE_SAVINGS}, false, nil)

	// Deposits from 2 accounts
	acc1 := suite.CreateAccount(sdk.NewCoins(startBalance), 0).GetAddress()
	acc2 := suite.CreateAccount(sdk.NewCoins(startBalance), 1).GetAddress()
	err := suite.Keeper.Deposit(suite.Ctx, acc1, depositAmount, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	err = suite.Keeper.Deposit(suite.Ctx, acc2, depositAmount, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	// Withdraw
	err = suite.Keeper.Withdraw(suite.Ctx, acc1, depositAmount, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	// Withdraw again
	err = suite.Keeper.Withdraw(suite.Ctx, acc1, depositAmount, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().Error(err)
	suite.Require().ErrorIs(
		err,
		types.ErrVaultShareRecordNotFound,
		"should only be able to withdraw the account's own supply",
	)
}

func (suite *strategySavingsTestSuite) TestWithdraw_WithAccumulatedSavings() {
	startBalance := sdk.NewInt64Coin(savingsVaultDenom, 1000)
	depositAmount := sdk.NewInt64Coin(savingsVaultDenom, 100)

	suite.CreateVault(savingsVaultDenom, types.StrategyTypes{types.STRATEGY_TYPE_SAVINGS}, false, nil)

	// Deposits accounts
	acc := suite.CreateAccount(sdk.NewCoins(startBalance), 0).GetAddress()
	acc2 := suite.CreateAccount(sdk.NewCoins(startBalance), 1).GetAddress()

	err := suite.Keeper.Deposit(suite.Ctx, acc, depositAmount, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	// Deposit from acc2 so the vault doesn't get deleted when withdrawing
	err = suite.Keeper.Deposit(suite.Ctx, acc2, depositAmount, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	// Direct savings deposit from module account to increase vault value
	suite.App.FundModuleAccount(suite.Ctx, types.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(savingsVaultDenom, 20)))
	macc := suite.AccountKeeper.GetModuleAccount(suite.Ctx, types.ModuleName)
	suite.SavingsKeeper.Deposit(suite.Ctx, macc.GetAddress(), sdk.NewCoins(sdk.NewInt64Coin(savingsVaultDenom, 20)))

	// Query account value
	accValue, err := suite.Keeper.GetVaultAccountValue(suite.Ctx, savingsVaultDenom, acc)
	suite.Require().NoError(err)
	suite.Equal(depositAmount.AddAmount(sdk.NewInt(10)), accValue)

	// Withdraw 100, 10 remaining
	err = suite.Keeper.Withdraw(suite.Ctx, acc, depositAmount, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	// Withdraw 100 again -- too much
	err = suite.Keeper.Withdraw(suite.Ctx, acc, depositAmount, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().Error(err)
	suite.Require().ErrorIs(
		err,
		types.ErrInsufficientValue,
		"cannot withdraw more than account value",
	)

	// Half of remaining 10, 5 remaining
	err = suite.Keeper.Withdraw(suite.Ctx, acc, sdk.NewCoin(savingsVaultDenom, sdk.NewInt(5)), types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	// Withdraw all
	err = suite.Keeper.Withdraw(suite.Ctx, acc, sdk.NewCoin(savingsVaultDenom, sdk.NewInt(5)), types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	accValue, err = suite.Keeper.GetVaultAccountValue(suite.Ctx, savingsVaultDenom, acc)
	suite.Require().Errorf(
		err,
		"account should be deleted when all shares withdrawn but has %s value still",
		accValue,
	)
	suite.Require().Equal("account vault share record for ukava not found", err.Error())
}

func (suite *strategySavingsTestSuite) TestAccountShares() {
	startBalance := sdk.NewInt64Coin(savingsVaultDenom, 1000)
	depositAmount := sdk.NewInt64Coin(savingsVaultDenom, 100)
	suite.App.FundModuleAccount(suite.Ctx, types.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(savingsVaultDenom, 1000)))

	suite.CreateVault(savingsVaultDenom, types.StrategyTypes{types.STRATEGY_TYPE_SAVINGS}, false, nil)

	// Deposit from account1
	acc1 := suite.CreateAccount(sdk.NewCoins(startBalance), 0).GetAddress()
	acc2 := suite.CreateAccount(sdk.NewCoins(startBalance), 1).GetAddress()

	// 1. acc1 deposit 100
	err := suite.Keeper.Deposit(suite.Ctx, acc1, depositAmount, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	acc1Shares, found := suite.Keeper.GetVaultAccountShares(suite.Ctx, acc1)
	suite.Require().True(found)
	suite.Equal(sdk.NewDec(100), acc1Shares.AmountOf(savingsVaultDenom), "initial deposit 1:1 shares")

	// 2. Direct savings deposit from module account to increase vault value
	// Total value: 100 -> 110
	macc := suite.AccountKeeper.GetModuleAccount(suite.Ctx, types.ModuleName)
	err = suite.SavingsKeeper.Deposit(suite.Ctx, macc.GetAddress(), sdk.NewCoins(sdk.NewInt64Coin(savingsVaultDenom, 10)))
	suite.Require().NoError(err)

	// 2. acc2 deposit 100
	// share price is 10% more expensive now
	// savings 110 -> 210
	err = suite.Keeper.Deposit(suite.Ctx, acc2, depositAmount, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	// 100 * 100 / 210 = 47.619047619 shares
	// 2.1 price * 47.619047619 = 99.9999999999
	acc2Value, err := suite.Keeper.GetVaultAccountValue(suite.Ctx, savingsVaultDenom, acc2)
	suite.Require().NoError(err)
	suite.Equal(
		sdk.NewInt(99),
		acc2Value.Amount,
		"value 1 less than deposit amount with different share price, decimals truncated",
	)

	acc2Shares, found := suite.Keeper.GetVaultAccountShares(suite.Ctx, acc2)
	suite.Require().True(found)
	// 100 * 100 / 110 = 190.909090909090909091
	// QuoInt64() truncates
	expectedAcc2Shares := sdk.NewDec(100).MulInt64(100).QuoInt64(110)
	suite.Equal(expectedAcc2Shares, acc2Shares.AmountOf(savingsVaultDenom))

	vaultTotalShares, found := suite.Keeper.GetVaultTotalShares(suite.Ctx, savingsVaultDenom)
	suite.Require().True(found)
	suite.Equal(sdk.NewDec(100).Add(expectedAcc2Shares), vaultTotalShares.Amount)

	// Savings deposit again from module account to triple original value
	// 210 -> 300
	suite.SavingsKeeper.Deposit(suite.Ctx, macc.GetAddress(), sdk.NewCoins(sdk.NewInt64Coin(savingsVaultDenom, 90)))

	// Deposit again from acc1
	err = suite.Keeper.Deposit(suite.Ctx, acc1, depositAmount, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	acc1Shares, found = suite.Keeper.GetVaultAccountShares(suite.Ctx, acc1)
	suite.Require().True(found)
	// totalShares = 100 + 90            = 190
	// totalValue  = 100 + 10 + 100 + 90 = 300
	// sharesIssued = assetAmount * (shareCount / totalTokens)
	// sharedIssued = 100 * 190 / 300 = 63.3 = 63
	// total shares = 100 + 63 = 163
	suite.Equal(
		sdk.NewDec(100).Add(sdk.NewDec(100).Mul(vaultTotalShares.Amount).Quo(sdk.NewDec(300))),
		acc1Shares.AmountOf(savingsVaultDenom),
		"shares should consist of 100 of 1x share price and 63 of 3x share price",
	)
}

func (suite *strategySavingsTestSuite) TestWithdraw_AccumulatedAmount() {
	startBalance := sdk.NewInt64Coin(savingsVaultDenom, 1000)
	depositAmount := sdk.NewInt64Coin(savingsVaultDenom, 100)
	suite.App.FundModuleAccount(suite.Ctx, types.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(savingsVaultDenom, 1000)))

	suite.CreateVault(savingsVaultDenom, types.StrategyTypes{types.STRATEGY_TYPE_SAVINGS}, false, nil)

	// Deposit from account1
	acc1 := suite.CreateAccount(sdk.NewCoins(startBalance), 0).GetAddress()
	acc2 := suite.CreateAccount(sdk.NewCoins(startBalance), 1).GetAddress()

	// 1. acc1 deposit 100
	err := suite.Keeper.Deposit(suite.Ctx, acc1, depositAmount, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	// acc2 deposit 100, just to make sure other deposits do not affect acc1
	err = suite.Keeper.Deposit(suite.Ctx, acc2, depositAmount, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	acc1Shares, found := suite.Keeper.GetVaultAccountShares(suite.Ctx, acc1)
	suite.Require().True(found)
	suite.Equal(sdk.NewDec(100), acc1Shares.AmountOf(savingsVaultDenom), "initial deposit 1:1 shares")

	// 2. Direct savings deposit from module account to increase vault value
	// Total value: 200 -> 220, 110 each account
	macc := suite.AccountKeeper.GetModuleAccount(suite.Ctx, types.ModuleName)
	err = suite.SavingsKeeper.Deposit(suite.Ctx, macc.GetAddress(), sdk.NewCoins(sdk.NewInt64Coin(savingsVaultDenom, 20)))
	suite.Require().NoError(err)

	// 3. Withdraw all from acc1 - including accumulated amount
	err = suite.Keeper.Withdraw(suite.Ctx, acc1, depositAmount.AddAmount(sdk.NewInt(10)), types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	_, found = suite.Keeper.GetVaultAccountShares(suite.Ctx, acc1)
	suite.Require().False(found, "should have withdrawn entire shares")
}

func (suite *strategySavingsTestSuite) TestWithdraw_AccumulatedTruncated() {
	startBalance := sdk.NewInt64Coin(savingsVaultDenom, 1000)
	depositAmount := sdk.NewInt64Coin(savingsVaultDenom, 100)
	suite.App.FundModuleAccount(suite.Ctx, types.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(savingsVaultDenom, 1000)))

	suite.CreateVault(savingsVaultDenom, types.StrategyTypes{types.STRATEGY_TYPE_SAVINGS}, false, nil)

	// Deposit from account1
	acc1 := suite.CreateAccount(sdk.NewCoins(startBalance), 0).GetAddress()
	acc2 := suite.CreateAccount(sdk.NewCoins(startBalance), 1).GetAddress()

	// 1. acc1 deposit 100
	err := suite.Keeper.Deposit(suite.Ctx, acc1, depositAmount, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	// acc2 deposit 100, just to make sure other deposits do not affect acc1
	err = suite.Keeper.Deposit(suite.Ctx, acc2, depositAmount, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	acc1Shares, found := suite.Keeper.GetVaultAccountShares(suite.Ctx, acc1)
	suite.Require().True(found)
	suite.Equal(sdk.NewDec(100), acc1Shares.AmountOf(savingsVaultDenom), "initial deposit 1:1 shares")

	// 2. Direct savings deposit from module account to increase vault value
	// Total value: 200 -> 211, 105.5 each account
	macc := suite.AccountKeeper.GetModuleAccount(suite.Ctx, types.ModuleName)
	err = suite.SavingsKeeper.Deposit(suite.Ctx, macc.GetAddress(), sdk.NewCoins(sdk.NewInt64Coin(savingsVaultDenom, 11)))
	suite.Require().NoError(err)

	accBal, err := suite.Keeper.GetVaultAccountValue(suite.Ctx, savingsVaultDenom, acc1)
	suite.Require().NoError(err)
	suite.Equal(depositAmount.AddAmount(sdk.NewInt(5)), accBal, "acc1 should have 105 usdx")

	// 3. Withdraw all from acc1 - including accumulated amount
	err = suite.Keeper.Withdraw(suite.Ctx, acc1, depositAmount.AddAmount(sdk.NewInt(5)), types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	acc1Shares, found = suite.Keeper.GetVaultAccountShares(suite.Ctx, acc1)
	suite.Require().Falsef(found, "should have withdrawn entire shares but has %s", acc1Shares)

	_, err = suite.Keeper.GetVaultAccountValue(suite.Ctx, savingsVaultDenom, acc1)
	suite.Require().Error(err)
}

func (suite *strategySavingsTestSuite) TestWithdraw_ExpensiveShares() {
	startBalance := sdk.NewInt64Coin(savingsVaultDenom, 1000)
	depositAmount := sdk.NewInt64Coin(savingsVaultDenom, 100)
	suite.App.FundModuleAccount(suite.Ctx, types.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(savingsVaultDenom, 2000)))

	suite.CreateVault(savingsVaultDenom, types.StrategyTypes{types.STRATEGY_TYPE_SAVINGS}, false, nil)

	// Deposit from account1
	acc1 := suite.CreateAccount(sdk.NewCoins(startBalance), 0).GetAddress()

	// 1. acc1 deposit 100
	err := suite.Keeper.Deposit(suite.Ctx, acc1, depositAmount, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	acc1Shares, found := suite.Keeper.GetVaultAccountShares(suite.Ctx, acc1)
	suite.Require().True(found)
	suite.Equal(sdk.NewDec(100), acc1Shares.AmountOf(savingsVaultDenom), "initial deposit 1:1 shares")

	// 2. Direct savings deposit from module account to increase vault value
	// Total value: 100 -> 2000, shares now 10usdx each
	macc := suite.AccountKeeper.GetModuleAccount(suite.Ctx, types.ModuleName)
	err = suite.SavingsKeeper.Deposit(suite.Ctx, macc.GetAddress(), sdk.NewCoins(sdk.NewInt64Coin(savingsVaultDenom, 1900)))
	suite.Require().NoError(err)

	accBal, err := suite.Keeper.GetVaultAccountValue(suite.Ctx, savingsVaultDenom, acc1)
	suite.Require().NoError(err)
	suite.Equal(sdk.NewInt(2000), accBal.Amount, "acc1 should have 2000 usdx")

	// 3. Withdraw all from acc1 - including accumulated amount
	err = suite.Keeper.Withdraw(suite.Ctx, acc1, sdk.NewInt64Coin(savingsVaultDenom, 2000), types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	acc1Shares, found = suite.Keeper.GetVaultAccountShares(suite.Ctx, acc1)
	suite.Require().Falsef(found, "should have withdrawn entire shares but has %s", acc1Shares)

	_, err = suite.Keeper.GetVaultAccountValue(suite.Ctx, savingsVaultDenom, acc1)
	suite.Require().Error(err)
}
