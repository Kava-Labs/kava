package e2e_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/tests/e2e/testutil"
)

func ukava(amt int64) sdk.Coin {
	return sdk.NewCoin("ukava", sdk.NewInt(amt))
}

type IntegrationTestSuite struct {
	testutil.E2eTestSuite
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

// example test that queries kava chain & kava's EVM
func (suite *IntegrationTestSuite) TestChainID() {
	// TODO: make chain agnostic, don't hardcode expected chain ids

	evmNetworkId, err := suite.EvmClient.NetworkID(context.Background())
	suite.NoError(err)
	suite.Equal(big.NewInt(8888), evmNetworkId)

	nodeInfo, err := suite.Tm.GetNodeInfo(context.Background(), &tmservice.GetNodeInfoRequest{})
	suite.NoError(err)
	suite.Equal(testutil.ChainId, nodeInfo.DefaultNodeInfo.Network)
}

// example test that funds a new account & queries its balance
func (suite *IntegrationTestSuite) TestFundedAccount() {
	funds := ukava(1e7)
	acc := suite.NewFundedAccount("example-acc", sdk.NewCoins(funds))
	res, err := suite.Bank.Balance(context.Background(), banktypes.NewQueryBalanceRequest(
		acc.Address, "ukava",
	))
	suite.NoError(err)
	suite.Equal(funds, *res.Balance)
}
