package e2e_test

import (
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	precisebanktypes "github.com/kava-labs/kava/x/precisebank/types"
)

func (suite *IntegrationTestSuite) TestUpgrade_PreciseBankReserveTransfer() {
	suite.SkipIfUpgradeDisabled()

	beforeUpgradeCtx := suite.Kava.Grpc.CtxAtHeight(suite.UpgradeHeight - 1)
	afterUpgradeCtx := suite.Kava.Grpc.CtxAtHeight(suite.UpgradeHeight)

	grpcClient := suite.Kava.Grpc

	// -----------------------------
	// Get initial reserve balances
	evmutilAddr := "kava1w9vxuke5dz6hyza2j932qgmxltnfxwl78u920k"
	precisebankAddr := "kava12yfe2jaupmtjruwxsec7hg7er60fhaa4uz7ffl"

	previousEvmutilBalRes, err := grpcClient.Query.Bank.Balance(beforeUpgradeCtx, &banktypes.QueryBalanceRequest{
		Address: evmutilAddr,
		Denom:   precisebanktypes.IntegerCoinDenom,
	})
	suite.Require().NoError(err)
	suite.Require().NotNil(previousEvmutilBalRes.Balance)
	suite.Require().True(
		previousEvmutilBalRes.Balance.Amount.IsPositive(),
		"should have reserve balance before upgrade",
	)

	previousPrecisebankBalRes, err := grpcClient.Query.Bank.Balance(beforeUpgradeCtx, &banktypes.QueryBalanceRequest{
		Address: precisebankAddr,
		Denom:   precisebanktypes.IntegerCoinDenom,
	})
	suite.Require().NoError(err)
	suite.Require().NotNil(previousPrecisebankBalRes.Balance)
	suite.Require().True(
		previousPrecisebankBalRes.Balance.Amount.IsZero(),
		"should be empty before upgrade",
	)

	suite.T().Logf("x/evmutil balances before upgrade: %s", previousEvmutilBalRes.Balance)
	suite.T().Logf("x/precisebank balances before upgrade: %s", previousPrecisebankBalRes.Balance)

	// -----------------------------
	// After upgrade
	// - Check reserve balance transfer
	// - Check reserve fully backs fractional amounts
	afterEvmutilBalRes, err := grpcClient.Query.Bank.Balance(afterUpgradeCtx, &banktypes.QueryBalanceRequest{
		Address: evmutilAddr,
		Denom:   precisebanktypes.IntegerCoinDenom,
	})
	suite.Require().NoError(err)
	suite.Require().NotNil(afterEvmutilBalRes.Balance)
	suite.Require().Truef(
		afterEvmutilBalRes.Balance.Amount.IsZero(),
		"should have transferred all reserve balance to precisebank, expected 0 but got %s",
		afterEvmutilBalRes.Balance,
	)

	afterPrecisebankBalRes, err := grpcClient.Query.Bank.Balance(afterUpgradeCtx, &banktypes.QueryBalanceRequest{
		Address: precisebankAddr,
		Denom:   precisebanktypes.IntegerCoinDenom,
	})
	suite.Require().NoError(err)
	suite.Require().NotNil(afterPrecisebankBalRes.Balance)
	suite.Require().True(afterPrecisebankBalRes.Balance.Amount.IsPositive())

	suite.T().Logf("x/evmutil balances after upgrade: %s", afterEvmutilBalRes.Balance)
	suite.T().Logf("x/precisebank balances after upgrade: %s", afterPrecisebankBalRes.Balance)

	sumFractional, err := grpcClient.Query.Precisebank.TotalFractionalBalances(
		afterUpgradeCtx,
		&precisebanktypes.QueryTotalFractionalBalancesRequest{},
	)
	suite.Require().NoError(err)

	suite.Require().Equal(
		sumFractional.Total,
		afterPrecisebankBalRes.Balance.Amount.Mul(precisebanktypes.ConversionFactor()),
		"reserve should match exactly sum fractional balances",
	)
}
