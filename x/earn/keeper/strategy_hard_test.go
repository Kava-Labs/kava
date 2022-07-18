package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/earn/testutil"
	"github.com/kava-labs/kava/x/earn/types"
	"github.com/stretchr/testify/suite"
)

type strategyHardTestSuite struct {
	testutil.Suite
}

func (suite *strategyHardTestSuite) SetupTest() {
	suite.Suite.SetupTest()
	suite.Keeper.SetParams(suite.Ctx, types.DefaultParams())
}

func TestStrategyLendTestSuite(t *testing.T) {
	suite.Run(t, new(strategyHardTestSuite))
}

func (suite *strategyHardTestSuite) TestGetVaultTotalSupplied() {
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
