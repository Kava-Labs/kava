package e2e_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	emtypes "github.com/tharsis/ethermint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/tests/e2e/testutil"
	"github.com/kava-labs/kava/tests/util"
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
	// TODO: make chain agnostic, don't hardcode expected chain ids (in testutil)

	expectedEvmNetworkId, err := emtypes.ParseChainID(testutil.ChainId)
	suite.NoError(err)
	evmNetworkId, err := suite.EvmClient.NetworkID(context.Background())
	suite.NoError(err)
	suite.Equal(expectedEvmNetworkId, evmNetworkId)

	nodeInfo, err := suite.Tm.GetNodeInfo(context.Background(), &tmservice.GetNodeInfoRequest{})
	suite.NoError(err)
	suite.Equal(testutil.ChainId, nodeInfo.DefaultNodeInfo.Network)
}

// example test that funds a new account & queries its balance
func (suite *IntegrationTestSuite) TestFundedAccount() {
	funds := ukava(1e7)
	acc := suite.NewFundedAccount("example-acc", sdk.NewCoins(funds))

	// check that the sdk & evm signers are for the same account
	suite.Equal(acc.SdkAddress.String(), util.EvmToSdkAddress(acc.EvmAddress).String())
	suite.Equal(acc.EvmAddress.Hex(), util.SdkToEvmAddress(acc.SdkAddress).Hex())

	// check balance via SDK query
	res, err := suite.Bank.Balance(context.Background(), banktypes.NewQueryBalanceRequest(
		acc.SdkAddress, "ukava",
	))
	suite.NoError(err)
	suite.Equal(funds, *res.Balance)

	// TODO: check balance via EVM query
}

func (suite *IntegrationTestSuite) TestEvmTx() {
	initialFunds := ukava(1e7) // 10 KAVA
	acc := suite.NewFundedAccount("evm-test-transfer", sdk.NewCoins(initialFunds))

	randomAddr := app.RandomAddress()
	to := util.SdkToEvmAddress(randomAddr)

	nonce, err := suite.EvmClient.PendingNonceAt(context.Background(), acc.EvmAddress)
	suite.NoError(err)
	suite.Equal(uint64(0), nonce) // sanity check. the account should have no prior txs

	kavaToTransfer := big.NewInt(1e18) // 1 KAVA; akava has 18 decimals.
	req := util.EvmTxRequest{
		Tx:   ethtypes.NewTransaction(nonce, to, kavaToTransfer, 1e5, big.NewInt(1e10), nil),
		Data: "any ol' data to track this through the system",
	}
	res := acc.SignAndBroadcastEvmTx(req)
	suite.NoError(res.Err)
	suite.Equal(ethtypes.ReceiptStatusSuccessful, res.Receipt.Status)
}
