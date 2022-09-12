package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/earn/testutil"
	"github.com/kava-labs/kava/x/earn/types"
	"github.com/stretchr/testify/suite"
)

type vaultTestSuite struct {
	testutil.Suite
}

func (suite *vaultTestSuite) SetupTest() {
	suite.Suite.SetupTest()
	suite.Keeper.SetParams(suite.Ctx, types.DefaultParams())
}

func TestVaultTestSuite(t *testing.T) {
	suite.Run(t, new(vaultTestSuite))
}

func (suite *vaultTestSuite) TestGetVaultTotalShares() {
	vaultDenom := "usdx"
	startBalance := sdk.NewInt64Coin(vaultDenom, 1000)
	depositAmount := sdk.NewInt64Coin(vaultDenom, 100)

	suite.CreateVault(vaultDenom, types.STRATEGY_TYPE_HARD)

	acc := suite.CreateAccount(sdk.NewCoins(startBalance), 0)

	err := suite.Keeper.Deposit(suite.Ctx, acc.GetAddress(), depositAmount, types.STRATEGY_TYPE_HARD)
	suite.Require().NoError(err)

	vaultTotalShares, found := suite.Keeper.GetVaultTotalShares(suite.Ctx, vaultDenom)
	suite.Require().True(found)

	suite.Equal(depositAmount.Amount.ToDec(), vaultTotalShares.Amount)
}

func (suite *vaultTestSuite) TestGetVaultTotalShares_NotFound() {
	vaultDenom := "usdx"

	_, found := suite.Keeper.GetVaultTotalShares(suite.Ctx, vaultDenom)
	suite.Require().False(found)
}

func (suite *vaultTestSuite) TestGetVaultTotalValue() {
	vaultDenom := "usdx"

	suite.CreateVault(vaultDenom, types.STRATEGY_TYPE_HARD)

	totalValue, err := suite.Keeper.GetVaultTotalValue(suite.Ctx, vaultDenom)
	suite.Require().NoError(err)
	suite.Equal(sdk.NewInt(0), totalValue.Amount)
}

func (suite *vaultTestSuite) TestGetVaultTotalValue_NotFound() {
	vaultDenom := "usdx"

	_, err := suite.Keeper.GetVaultTotalValue(suite.Ctx, vaultDenom)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, types.ErrVaultRecordNotFound)
}

func (suite *vaultTestSuite) TestInvalidVaultStrategy() {
	vaultDenom := "usdx"

	suite.PanicsWithValue("value from ParamSetPair is invalid: invalid strategy 99999", func() {
		suite.CreateVault(vaultDenom, 99999) // not valid strategy type
	})
}

func (suite *vaultTestSuite) TestGetVaultAccountSupplied() {
	vaultDenom := "usdx"
	startBalance := sdk.NewInt64Coin(vaultDenom, 1000)
	deposit1Amount := sdk.NewInt64Coin(vaultDenom, 100)
	deposit2Amount := sdk.NewInt64Coin(vaultDenom, 100)

	suite.CreateVault(vaultDenom, types.STRATEGY_TYPE_HARD)

	acc1 := suite.CreateAccount(sdk.NewCoins(startBalance), 0)
	acc2 := suite.CreateAccount(sdk.NewCoins(startBalance), 1)

	// Before deposit, account supplied is 0

	_, found := suite.Keeper.GetVaultShareRecord(suite.Ctx, acc1.GetAddress())
	suite.Require().False(found)

	_, found = suite.Keeper.GetVaultShareRecord(suite.Ctx, acc2.GetAddress())
	suite.Require().False(found)

	// Deposits from both accounts
	err := suite.Keeper.Deposit(suite.Ctx, acc1.GetAddress(), deposit1Amount, types.STRATEGY_TYPE_HARD)
	suite.Require().NoError(err)

	err = suite.Keeper.Deposit(suite.Ctx, acc2.GetAddress(), deposit2Amount, types.STRATEGY_TYPE_HARD)
	suite.Require().NoError(err)

	// Check balances

	vaultAcc1Supplied, found := suite.Keeper.GetVaultShareRecord(suite.Ctx, acc1.GetAddress())
	suite.Require().True(found)

	vaultAcc2Supplied, found := suite.Keeper.GetVaultShareRecord(suite.Ctx, acc2.GetAddress())
	suite.Require().True(found)

	// Account supply only includes the deposit from respective accounts
	suite.Equal(deposit1Amount.Amount.ToDec(), vaultAcc1Supplied.Shares.AmountOf(vaultDenom))
	suite.Equal(deposit1Amount.Amount.ToDec(), vaultAcc2Supplied.Shares.AmountOf(vaultDenom))
}

func (suite *vaultTestSuite) TestGetVaultAccountValue() {
	vaultDenom := "usdx"
	startBalance := sdk.NewInt64Coin(vaultDenom, 1000)
	depositAmount := sdk.NewInt64Coin(vaultDenom, 100)

	acc := suite.CreateAccount(sdk.NewCoins(startBalance), 0)

	suite.CreateVault(vaultDenom, types.STRATEGY_TYPE_HARD)

	err := suite.Keeper.Deposit(suite.Ctx, acc.GetAddress(), depositAmount, types.STRATEGY_TYPE_HARD)
	suite.Require().NoError(err)

	accValue, err := suite.Keeper.GetVaultAccountValue(suite.Ctx, vaultDenom, acc.GetAddress())
	suite.Require().NoError(err)
	suite.Equal(depositAmount, accValue, "value should be same as deposit amount")
}

func (suite *vaultTestSuite) TestGetVaultAccountValue_VaultNotFound() {
	vaultDenom := "usdx"
	acc := suite.CreateAccount(sdk.NewCoins(), 0)

	_, err := suite.Keeper.GetVaultAccountValue(suite.Ctx, vaultDenom, acc.GetAddress())
	suite.Require().Error(err)
	suite.Require().Equal("account vault share record for usdx not found", err.Error())
}

func (suite *vaultTestSuite) TestGetVaultAccountValue_ShareNotFound() {
	vaultDenom := "usdx"
	startBalance := sdk.NewInt64Coin(vaultDenom, 1000)
	depositAmount := sdk.NewInt64Coin(vaultDenom, 100)

	acc1 := suite.CreateAccount(sdk.NewCoins(startBalance), 0)
	acc2 := suite.CreateAccount(sdk.NewCoins(startBalance), 1)

	suite.CreateVault(vaultDenom, types.STRATEGY_TYPE_HARD)

	// Deposit from acc1 so that vault record exists
	err := suite.Keeper.Deposit(suite.Ctx, acc1.GetAddress(), depositAmount, types.STRATEGY_TYPE_HARD)
	suite.Require().NoError(err)

	// Query from acc2 with no share record
	_, err = suite.Keeper.GetVaultAccountValue(suite.Ctx, vaultDenom, acc2.GetAddress())
	suite.Require().Error(err)
	suite.Require().Equal("account vault share record for usdx not found", err.Error())
}
