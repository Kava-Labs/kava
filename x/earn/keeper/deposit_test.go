package keeper_test

import (
	"os"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/earn/testutil"
	"github.com/kava-labs/kava/x/earn/types"
	"github.com/stretchr/testify/suite"
)

func TestMain(m *testing.M) {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	os.Exit(m.Run())
}

type depositTestSuite struct {
	testutil.Suite
}

func (suite *depositTestSuite) SetupTest() {
	suite.Suite.SetupTest()
	suite.Keeper.SetParams(suite.Ctx, types.DefaultParams())
}

func TestDepositTestSuite(t *testing.T) {
	suite.Run(t, new(depositTestSuite))
}

func (suite *depositTestSuite) TestDeposit_Balances() {
	vaultDenom := "usdx"
	startBalance := sdk.NewInt64Coin(vaultDenom, 1000)
	depositAmount := sdk.NewInt64Coin(vaultDenom, 100)

	suite.CreateVault(vaultDenom, types.STRATEGY_TYPE_LEND)

	acc := suite.CreateAccount(sdk.NewCoins(startBalance), 0)

	err := suite.Keeper.Deposit(suite.Ctx, acc.GetAddress(), depositAmount)
	suite.Require().NoError(err)

	suite.AccountBalanceEqual(
		acc.GetAddress(),
		sdk.NewCoins(startBalance.Sub(depositAmount)), // Account decreases by deposit
	)

	// TODO: Module account balance will be zero when strategies are implemented
	suite.ModuleAccountBalanceEqual(
		sdk.NewCoins(depositAmount),
	)
}

func (suite *depositTestSuite) TestDeposit_Exceed() {
	vaultDenom := "usdx"
	startBalance := sdk.NewInt64Coin(vaultDenom, 1000)
	depositAmount := sdk.NewInt64Coin(vaultDenom, 1001)

	suite.CreateVault(vaultDenom, types.STRATEGY_TYPE_LEND)

	acc := suite.CreateAccount(sdk.NewCoins(startBalance), 0)

	err := suite.Keeper.Deposit(suite.Ctx, acc.GetAddress(), depositAmount)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, sdkerrors.ErrInsufficientFunds)

	// No changes in balances

	suite.AccountBalanceEqual(
		acc.GetAddress(),
		sdk.NewCoins(startBalance),
	)

	suite.ModuleAccountBalanceEqual(
		sdk.NewCoins(),
	)
}

func (suite *depositTestSuite) TestDeposit_Zero() {
	vaultDenom := "usdx"
	startBalance := sdk.NewInt64Coin(vaultDenom, 1000)
	depositAmount := sdk.NewInt64Coin(vaultDenom, 0)

	suite.CreateVault(vaultDenom, types.STRATEGY_TYPE_LEND)

	acc := suite.CreateAccount(sdk.NewCoins(startBalance), 0)

	err := suite.Keeper.Deposit(suite.Ctx, acc.GetAddress(), depositAmount)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, types.ErrInsufficientAmount)

	// No changes in balances

	suite.AccountBalanceEqual(
		acc.GetAddress(),
		sdk.NewCoins(startBalance),
	)

	suite.ModuleAccountBalanceEqual(
		sdk.NewCoins(),
	)
}

func (suite *depositTestSuite) TestDeposit_InvalidVault() {
	vaultDenom := "usdx"
	startBalance := sdk.NewInt64Coin(vaultDenom, 1000)
	depositAmount := sdk.NewInt64Coin(vaultDenom, 1001)

	// Vault not created -- doesn't exist

	acc := suite.CreateAccount(sdk.NewCoins(startBalance), 0)

	err := suite.Keeper.Deposit(suite.Ctx, acc.GetAddress(), depositAmount)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, types.ErrInvalidVaultDenom)

	// No changes in balances

	suite.AccountBalanceEqual(
		acc.GetAddress(),
		sdk.NewCoins(startBalance),
	)

	suite.ModuleAccountBalanceEqual(
		sdk.NewCoins(),
	)
}
