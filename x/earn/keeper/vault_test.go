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

func (suite *vaultTestSuite) TestGetVaultTotalSupplied() {
	vaultDenom := "usdx"
	startBalance := sdk.NewInt64Coin(vaultDenom, 1000)
	depositAmount := sdk.NewInt64Coin(vaultDenom, 100)

	suite.CreateVault(vaultDenom, types.STRATEGY_TYPE_HARD)

	acc := suite.CreateAccount(sdk.NewCoins(startBalance), 0)

	err := suite.Keeper.Deposit(suite.Ctx, acc.GetAddress(), depositAmount)
	suite.Require().NoError(err)

	vaultTotalSupplied, err := suite.Keeper.GetVaultTotalSupplied(suite.Ctx, vaultDenom)
	suite.Require().NoError(err)

	suite.Equal(depositAmount, vaultTotalSupplied)
}

func (suite *vaultTestSuite) TestGetVaultTotalSupplied_NotFound() {
	vaultDenom := "usdx"

	_, err := suite.Keeper.GetVaultTotalSupplied(suite.Ctx, vaultDenom)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, types.ErrVaultRecordNotFound)
}

func (suite *vaultTestSuite) TestGetVaultTotalValue() {
	// TODO: After strategy implemented GetEstimatedTotalAssets
}

func (suite *vaultTestSuite) TestGetVaultTotalValue_NotFound() {
	vaultDenom := "usdx"

	_, err := suite.Keeper.GetVaultTotalValue(suite.Ctx, vaultDenom)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, types.ErrVaultRecordNotFound)
}

func (suite *vaultTestSuite) TestGetVaultTotalValue_InvalidStrategy() {
	vaultDenom := "usdx"
	suite.CreateVault(vaultDenom, 99999) // not valid strategy type

	_, err := suite.Keeper.GetVaultTotalValue(suite.Ctx, vaultDenom)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, types.ErrInvalidVaultStrategy)
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

	_, err := suite.Keeper.GetVaultAccountSupplied(suite.Ctx, vaultDenom, acc1.GetAddress())
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, types.ErrVaultShareRecordNotFound)

	_, err = suite.Keeper.GetVaultAccountSupplied(suite.Ctx, vaultDenom, acc2.GetAddress())
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, types.ErrVaultShareRecordNotFound)

	// Deposits from both accounts

	err = suite.Keeper.Deposit(suite.Ctx, acc1.GetAddress(), deposit1Amount)
	suite.Require().NoError(err)

	err = suite.Keeper.Deposit(suite.Ctx, acc2.GetAddress(), deposit2Amount)
	suite.Require().NoError(err)

	// Check balances

	vaultAcc1Supplied, err := suite.Keeper.GetVaultAccountSupplied(suite.Ctx, vaultDenom, acc1.GetAddress())
	suite.Require().NoError(err)

	vaultAcc2Supplied, err := suite.Keeper.GetVaultAccountSupplied(suite.Ctx, vaultDenom, acc2.GetAddress())
	suite.Require().NoError(err)

	// Account supply only includes the deposit from respective accounts
	suite.Equal(deposit1Amount, vaultAcc1Supplied)
	suite.Equal(deposit1Amount, vaultAcc2Supplied)
}

func (suite *vaultTestSuite) TestGetVaultAccountValue() {
	vaultDenom := "usdx"
	startBalance := sdk.NewInt64Coin(vaultDenom, 1000)
	depositAmount := sdk.NewInt64Coin(vaultDenom, 100)

	acc := suite.CreateAccount(sdk.NewCoins(startBalance), 0)

	suite.CreateVault(vaultDenom, types.STRATEGY_TYPE_HARD)

	err := suite.Keeper.Deposit(suite.Ctx, acc.GetAddress(), depositAmount)
	suite.Require().NoError(err)

	suite.T().Skip("TODO: After strategy GetEstimatedTotalAssets implemented")

	_, err = suite.Keeper.GetVaultAccountValue(suite.Ctx, vaultDenom, acc.GetAddress())
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, types.ErrVaultShareRecordNotFound)
}

func (suite *vaultTestSuite) TestGetVaultAccountValue_VaultNotFound() {
	vaultDenom := "usdx"
	acc := suite.CreateAccount(sdk.NewCoins(), 0)

	_, err := suite.Keeper.GetVaultAccountValue(suite.Ctx, vaultDenom, acc.GetAddress())
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, types.ErrVaultRecordNotFound)
}

func (suite *vaultTestSuite) TestGetVaultAccountValue_ShareNotFound() {
	vaultDenom := "usdx"
	startBalance := sdk.NewInt64Coin(vaultDenom, 1000)
	depositAmount := sdk.NewInt64Coin(vaultDenom, 100)

	acc1 := suite.CreateAccount(sdk.NewCoins(startBalance), 0)
	acc2 := suite.CreateAccount(sdk.NewCoins(startBalance), 1)

	suite.CreateVault(vaultDenom, types.STRATEGY_TYPE_HARD)

	// Deposit from acc1 so that vault record exists
	err := suite.Keeper.Deposit(suite.Ctx, acc1.GetAddress(), depositAmount)
	suite.Require().NoError(err)

	// Query from acc2 with no share record
	_, err = suite.Keeper.GetVaultAccountValue(suite.Ctx, vaultDenom, acc2.GetAddress())
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, types.ErrVaultShareRecordNotFound)
}

// ----------------------------------------------------------------------------
// State methods

func (suite *vaultTestSuite) TestGetVaultRecord() {
	record := types.NewVaultRecord("usdx")

	_, found := suite.Keeper.GetVaultRecord(suite.Ctx, record.Denom)
	suite.Require().False(found)

	suite.Keeper.SetVaultRecord(suite.Ctx, record)

	stateRecord, found := suite.Keeper.GetVaultRecord(suite.Ctx, record.Denom)
	suite.Require().True(found)
	suite.Require().Equal(record, stateRecord)
}

func (suite *vaultTestSuite) TestUpdateVaultRecord() {
	record := types.NewVaultRecord("usdx")

	record.TotalSupply = sdk.NewInt64Coin("usdx", 100)

	// Update vault
	suite.Keeper.UpdateVaultRecord(suite.Ctx, record)

	stateRecord, found := suite.Keeper.GetVaultRecord(suite.Ctx, record.Denom)
	suite.Require().True(found, "vault record with supply should exist")
	suite.Require().Equal(record, stateRecord)

	// Remove supply
	record.TotalSupply = sdk.NewInt64Coin("usdx", 0)
	suite.Keeper.UpdateVaultRecord(suite.Ctx, record)

	_, found = suite.Keeper.GetVaultRecord(suite.Ctx, record.Denom)
	suite.Require().False(found, "vault record with 0 supply should be deleted")
}

func (suite *vaultTestSuite) TestGetVaultShareRecord() {
	vaultDenom := "usdx"
	startBalance := sdk.NewInt64Coin(vaultDenom, 1000)
	depositAmount := sdk.NewInt64Coin(vaultDenom, 100)
	acc := suite.CreateAccount(sdk.NewCoins(startBalance), 0)

	record := types.NewVaultShareRecord(acc.GetAddress(), vaultDenom)

	// Check share doesn't exist before deposit

	_, found := suite.Keeper.GetVaultShareRecord(suite.Ctx, vaultDenom, acc.GetAddress())
	suite.Require().False(found, "vault share record should not exist before deposit")

	// Update share record
	record.AmountSupplied = depositAmount
	suite.Keeper.SetVaultShareRecord(suite.Ctx, record)

	// Check share exists and matches set value
	stateRecord, found := suite.Keeper.GetVaultShareRecord(suite.Ctx, vaultDenom, acc.GetAddress())
	suite.Require().True(found)
	suite.Require().Equal(record, stateRecord)
}

func (suite *vaultTestSuite) TestUpdateVaultShareRecord() {
	vaultDenom := "usdx"
	startBalance := sdk.NewInt64Coin(vaultDenom, 1000)
	depositAmount := sdk.NewInt64Coin(vaultDenom, 100)
	acc := suite.CreateAccount(sdk.NewCoins(startBalance), 0)

	record := types.NewVaultShareRecord(acc.GetAddress(), vaultDenom)

	record.AmountSupplied = depositAmount

	// Update vault
	suite.Keeper.UpdateVaultShareRecord(suite.Ctx, record)

	stateRecord, found := suite.Keeper.GetVaultShareRecord(suite.Ctx, vaultDenom, acc.GetAddress())
	suite.Require().True(found, "vault share record with supply should exist")
	suite.Require().Equal(record, stateRecord)

	// Remove supply
	record.AmountSupplied = sdk.NewInt64Coin("usdx", 0)
	suite.Keeper.UpdateVaultShareRecord(suite.Ctx, record)

	_, found = suite.Keeper.GetVaultShareRecord(suite.Ctx, vaultDenom, acc.GetAddress())
	suite.Require().False(found, "vault share record with 0 supply should be deleted")
}
