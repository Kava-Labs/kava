package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/earn/types"
)

// ----------------------------------------------------------------------------
// State methods

func (suite *vaultTestSuite) TestGetVaultRecord() {
	record := types.NewVaultRecord("usdx", sdk.ZeroDec())

	_, found := suite.Keeper.GetVaultRecord(suite.Ctx, record.TotalShares.Denom)
	suite.Require().False(found)

	suite.Keeper.SetVaultRecord(suite.Ctx, record)

	stateRecord, found := suite.Keeper.GetVaultRecord(suite.Ctx, record.TotalShares.Denom)
	suite.Require().True(found)
	suite.Require().Equal(record, stateRecord)
}

func (suite *vaultTestSuite) TestUpdateVaultRecord() {
	record := types.NewVaultRecord("usdx", sdk.ZeroDec())

	record.TotalShares = types.NewVaultShare("usdx", sdk.NewDec(100))

	// Update vault
	suite.Keeper.UpdateVaultRecord(suite.Ctx, record)

	stateRecord, found := suite.Keeper.GetVaultRecord(suite.Ctx, record.TotalShares.Denom)
	suite.Require().True(found, "vault record with supply should exist")
	suite.Require().Equal(record, stateRecord)

	// Remove supply
	record.TotalShares = types.NewVaultShare("usdx", sdk.NewDec(0))
	suite.Keeper.UpdateVaultRecord(suite.Ctx, record)

	_, found = suite.Keeper.GetVaultRecord(suite.Ctx, record.TotalShares.Denom)
	suite.Require().False(found, "vault record with 0 supply should be deleted")
}

func (suite *vaultTestSuite) TestGetVaultShareRecord() {
	vaultDenom := "usdx"
	startBalance := sdk.NewInt64Coin(vaultDenom, 1000)
	acc := suite.CreateAccount(sdk.NewCoins(startBalance), 0)

	record := types.NewVaultShareRecord(acc.GetAddress(), types.NewVaultShares())

	// Check share doesn't exist before deposit

	_, found := suite.Keeper.GetVaultShareRecord(suite.Ctx, acc.GetAddress())
	suite.Require().False(found, "vault share record should not exist before deposit")

	// Update share record
	record.Shares = types.NewVaultShares(
		types.NewVaultShare(vaultDenom, sdk.NewDec(100)),
	)
	suite.Keeper.SetVaultShareRecord(suite.Ctx, record)

	// Check share exists and matches set value
	stateRecord, found := suite.Keeper.GetVaultShareRecord(suite.Ctx, acc.GetAddress())
	suite.Require().True(found)
	suite.Require().Equal(record, stateRecord)
}

func (suite *vaultTestSuite) TestUpdateVaultShareRecord() {
	vaultDenom := "usdx"
	startBalance := sdk.NewInt64Coin(vaultDenom, 1000)
	acc := suite.CreateAccount(sdk.NewCoins(startBalance), 0)

	record := types.NewVaultShareRecord(acc.GetAddress(), types.NewVaultShares(
		types.NewVaultShare(vaultDenom, sdk.NewDec(100)),
	))

	// Update vault
	suite.Keeper.UpdateVaultShareRecord(suite.Ctx, record)

	stateRecord, found := suite.Keeper.GetVaultShareRecord(suite.Ctx, acc.GetAddress())
	suite.Require().True(found, "vault share record with supply should exist")
	suite.Require().Equal(record, stateRecord)

	// Remove supply
	record.Shares = types.NewVaultShares()
	suite.Keeper.UpdateVaultShareRecord(suite.Ctx, record)

	_, found = suite.Keeper.GetVaultShareRecord(suite.Ctx, acc.GetAddress())
	suite.Require().False(found, "vault share record with 0 supply should be deleted")
}
