package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/earn/testutil"
	"github.com/kava-labs/kava/x/earn/types"
)

type vaultShareTestSuite struct {
	testutil.Suite
}

func (suite *vaultShareTestSuite) SetupTest() {
	suite.Suite.SetupTest()
	suite.Keeper.SetParams(suite.Ctx, types.DefaultParams())
}

func TestVaultShareTestSuite(t *testing.T) {
	suite.Run(t, new(vaultShareTestSuite))
}

func (suite *vaultShareTestSuite) TestConvertToShares() {
	vaultDenom := "usdx"

	tests := []struct {
		name          string
		beforeConvert func()
		giveAmount    sdk.Coin
		wantShares    types.VaultShare
	}{
		{
			name:          "initial 1:1",
			beforeConvert: func() {},
			giveAmount:    sdk.NewCoin(vaultDenom, sdk.NewInt(100)),
			wantShares:    types.NewVaultShare(vaultDenom, sdk.NewDec(100)),
		},
		{
			name: "value doubled",

			beforeConvert: func() {
				// set total shares set total value for hard
				// value is double than shares
				// shares is 2x price now
				suite.addTotalShareAndValue(vaultDenom, sdk.NewDec(100), sdk.NewInt(200))
			},
			giveAmount: sdk.NewCoin(vaultDenom, sdk.NewInt(100)),
			wantShares: types.NewVaultShare(vaultDenom, sdk.NewDec(50)),
		},
		{
			name: "truncate",

			beforeConvert: func() {
				suite.addTotalShareAndValue(vaultDenom, sdk.NewDec(1000), sdk.NewInt(1001))
			},
			giveAmount: sdk.NewCoin(vaultDenom, sdk.NewInt(100)),
			// 100 * 100 / 101 = 99.0099something
			wantShares: types.NewVaultShare(vaultDenom, sdk.NewDec(100).MulInt64(1000).QuoInt64(1001)),
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset state
			suite.Suite.SetupTest()
			suite.CreateVault(vaultDenom, types.StrategyTypes{types.STRATEGY_TYPE_HARD}, false, nil)
			err := suite.App.FundModuleAccount(
				suite.Ctx,
				types.ModuleName,
				sdk.NewCoins(sdk.NewInt64Coin(vaultDenom, 10000)),
			)
			suite.Require().NoError(err)

			// Run any deposits or any other setup
			tt.beforeConvert()

			issuedShares, err := suite.Keeper.ConvertToShares(suite.Ctx, tt.giveAmount)
			suite.Require().NoError(err)

			suite.Equal(tt.wantShares, issuedShares)
		})
	}
}

func (suite *vaultShareTestSuite) addTotalShareAndValue(
	vaultDenom string,
	vaultShares sdk.Dec,
	hardDeposit sdk.Int,
) {
	macc := suite.AccountKeeper.GetModuleAccount(suite.Ctx, types.ModuleName)

	vaultRecord, found := suite.Keeper.GetVaultRecord(suite.Ctx, vaultDenom)
	if !found {
		vaultRecord = types.NewVaultRecord(vaultDenom, sdk.ZeroDec())
	}

	// Add to vault record
	vaultRecord.TotalShares.Amount = vaultRecord.TotalShares.Amount.Add(vaultShares)

	// set total shares
	suite.Keeper.UpdateVaultRecord(
		suite.Ctx,
		vaultRecord,
	)
	// add value for hard -- this does not set
	err := suite.HardKeeper.Deposit(
		suite.Ctx,
		macc.GetAddress(),
		sdk.NewCoins(sdk.NewCoin(vaultDenom, hardDeposit)),
	)
	suite.Require().NoError(err)
}

func TestPrecisionMulQuoOrder(t *testing.T) {
	assetAmount := sdk.NewDec(100)
	totalShares := sdk.NewDec(100)
	totalValue := sdk.NewDec(105)

	// issuedShares =  assetAmount * (totalValue / totalShares)
	//              = (assetAmount * totalShares) / totalValue
	mulFirst := assetAmount.Mul(totalShares).QuoTruncate(totalValue)
	quoFirst := assetAmount.Mul(totalShares.QuoTruncate(totalValue))

	assert.Equal(t, sdk.MustNewDecFromStr("95.238095238095238095"), mulFirst)
	assert.Equal(t, sdk.MustNewDecFromStr("95.238095238095238000"), quoFirst)

	assert.NotEqual(t, mulFirst, quoFirst)
}
