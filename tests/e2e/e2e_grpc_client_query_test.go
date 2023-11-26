package e2e_test

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	evmutiltypes "github.com/kava-labs/kava/x/evmutil/types"
)

func (suite *IntegrationTestSuite) TestGrpcClientQueryCosmosModule_Balance() {
	// ARRANGE
	// setup kava account
	funds := ukava(1e5) // .1 KAVA
	kavaAcc := suite.Kava.NewFundedAccount("balance-test", sdk.NewCoins(funds))

	// ACT
	rsp, err := suite.Kava.Grpc.Query.Bank.Balance(context.Background(), &banktypes.QueryBalanceRequest{
		Address: kavaAcc.SdkAddress.String(),
		Denom:   funds.Denom,
	})

	// ASSERT
	suite.Require().NoError(err)
	suite.Require().Equal(funds.Amount, rsp.Balance.Amount)
}

func (suite *IntegrationTestSuite) TestGrpcClientQueryKavaModule_EvmParams() {
	// ACT
	rsp, err := suite.Kava.Grpc.Query.Evmutil.Params(
		context.Background(), &evmutiltypes.QueryParamsRequest{},
	)

	// ASSERT
	suite.Require().NoError(err)
	suite.Require().GreaterOrEqual(len(rsp.Params.AllowedCosmosDenoms), 1)
	suite.Require().GreaterOrEqual(len(rsp.Params.EnabledConversionPairs), 1)
}
