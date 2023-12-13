package e2e_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *IntegrationTestSuite) TestGrpcClientUtil_Account() {
	// ARRANGE
	// setup kava account
	kavaAcc := suite.Kava.NewFundedAccount("account-test", sdk.NewCoins(ukava(1e5)))

	// ACT
	rsp, err := suite.Kava.Grpc.BaseAccount(kavaAcc.SdkAddress.String())

	// ASSERT
	suite.Require().NoError(err)
	suite.Equal(kavaAcc.SdkAddress.String(), rsp.Address)
	suite.Greater(rsp.AccountNumber, uint64(1))
	suite.Equal(uint64(0), rsp.Sequence)
}
