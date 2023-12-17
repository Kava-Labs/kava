package e2e_test

import (
	"context"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	emtypes "github.com/evmos/ethermint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/tests/e2e/testutil"
	"github.com/kava-labs/kava/tests/util"
)

var (
	minEvmGasPrice = big.NewInt(1e10) // akava
)

func ukava(amt int64) sdk.Coin {
	return sdk.NewCoin("ukava", sdkmath.NewInt(amt))
}

type IntegrationTestSuite struct {
	testutil.E2eTestSuite
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

// example test that queries kava via SDK and EVM
func (suite *IntegrationTestSuite) TestChainID() {
	expectedEvmNetworkId, err := emtypes.ParseChainID(suite.Kava.ChainID)
	suite.NoError(err)

	// EVM query
	evmNetworkId, err := suite.Kava.EvmClient.NetworkID(context.Background())
	suite.NoError(err)
	suite.Equal(expectedEvmNetworkId, evmNetworkId)

	// SDK query
	nodeInfo, err := suite.Kava.Grpc.Query.Tm.GetNodeInfo(context.Background(), &tmservice.GetNodeInfoRequest{})
	suite.NoError(err)
	suite.Equal(suite.Kava.ChainID, nodeInfo.DefaultNodeInfo.Network)
}

// example test that funds a new account & queries its balance
func (suite *IntegrationTestSuite) TestFundedAccount() {
	funds := ukava(1e3)
	acc := suite.Kava.NewFundedAccount("example-acc", sdk.NewCoins(funds))

	// check that the sdk & evm signers are for the same account
	suite.Equal(acc.SdkAddress.String(), util.EvmToSdkAddress(acc.EvmAddress).String())
	suite.Equal(acc.EvmAddress.Hex(), util.SdkToEvmAddress(acc.SdkAddress).Hex())

	// check balance via SDK query
	res, err := suite.Kava.Grpc.Query.Bank.Balance(context.Background(), banktypes.NewQueryBalanceRequest(
		acc.SdkAddress, "ukava",
	))
	suite.NoError(err)
	suite.Equal(funds, *res.Balance)

	// check balance via EVM query
	akavaBal, err := suite.Kava.EvmClient.BalanceAt(context.Background(), acc.EvmAddress, nil)
	suite.NoError(err)
	suite.Equal(funds.Amount.MulRaw(1e12).BigInt(), akavaBal)
}

// example test that signs & broadcasts an EVM tx
func (suite *IntegrationTestSuite) TestTransferOverEVM() {
	// fund an account that can perform the transfer
	initialFunds := ukava(1e6) // 1 KAVA
	acc := suite.Kava.NewFundedAccount("evm-test-transfer", sdk.NewCoins(initialFunds))

	// get a rando account to send kava to
	randomAddr := app.RandomAddress()
	to := util.SdkToEvmAddress(randomAddr)

	// example fetching of nonce (account sequence)
	nonce, err := suite.Kava.EvmClient.PendingNonceAt(context.Background(), acc.EvmAddress)
	suite.NoError(err)
	suite.Equal(uint64(0), nonce) // sanity check. the account should have no prior txs

	// transfer kava over EVM
	kavaToTransfer := big.NewInt(1e17) // .1 KAVA; akava has 18 decimals.
	req := util.EvmTxRequest{
		Tx:   ethtypes.NewTransaction(nonce, to, kavaToTransfer, 1e5, minEvmGasPrice, nil),
		Data: "any ol' data to track this through the system",
	}
	res := acc.SignAndBroadcastEvmTx(req)
	suite.Require().NoError(res.Err)
	suite.Equal(ethtypes.ReceiptStatusSuccessful, res.Receipt.Status)

	// evm txs refund unused gas. so to know the expected balance we need to know how much gas was used.
	ukavaUsedForGas := sdkmath.NewIntFromBigInt(minEvmGasPrice).
		Mul(sdkmath.NewIntFromUint64(res.Receipt.GasUsed)).
		QuoRaw(1e12) // convert akava to ukava

	// expect (9 - gas used) KAVA remaining in account.
	balance := suite.Kava.QuerySdkForBalances(acc.SdkAddress)
	suite.Equal(sdkmath.NewInt(9e5).Sub(ukavaUsedForGas), balance.AmountOf("ukava"))
}

// TestIbcTransfer transfers KAVA from the primary kava chain (suite.Kava) to the ibc chain (suite.Ibc).
// Note that because the IBC chain also runs kava's binary, this tests both the sending & receiving.
func (suite *IntegrationTestSuite) TestIbcTransfer() {
	suite.SkipIfIbcDisabled()

	// ARRANGE
	// setup kava account
	funds := ukava(1e5) // .1 KAVA
	kavaAcc := suite.Kava.NewFundedAccount("ibc-transfer-kava-side", sdk.NewCoins(funds))
	// setup ibc account
	ibcAcc := suite.Ibc.NewFundedAccount("ibc-transfer-ibc-side", sdk.NewCoins())

	gasLimit := int64(2e5)
	fee := ukava(200)

	fundsToSend := ukava(5e4) // .005 KAVA
	transferMsg := ibctypes.NewMsgTransfer(
		testutil.IbcPort,
		testutil.IbcChannel,
		fundsToSend,
		kavaAcc.SdkAddress.String(),
		ibcAcc.SdkAddress.String(),
		ibcclienttypes.NewHeight(0, 0), // timeout height disabled when 0
		uint64(time.Now().Add(30*time.Second).UnixNano()),
		"",
	)
	// initial - sent - fee
	expectedSrcBalance := funds.Sub(fundsToSend).Sub(fee)

	// ACT
	// IBC transfer from kava -> ibc
	transferTo := util.KavaMsgRequest{
		Msgs:      []sdk.Msg{transferMsg},
		GasLimit:  uint64(gasLimit),
		FeeAmount: sdk.NewCoins(fee),
		Memo:      "sent from Kava!",
	}
	res := kavaAcc.SignAndBroadcastKavaTx(transferTo)

	// ASSERT
	suite.NoError(res.Err)

	// the balance should be deducted from kava account
	suite.Eventually(func() bool {
		balance := suite.Kava.QuerySdkForBalances(kavaAcc.SdkAddress)
		return balance.AmountOf("ukava").Equal(expectedSrcBalance.Amount)
	}, 10*time.Second, 1*time.Second)

	// expect the balance to be transferred to the ibc chain!
	suite.Eventually(func() bool {
		balance := suite.Ibc.QuerySdkForBalances(ibcAcc.SdkAddress)
		found := false
		for _, c := range balance {
			// find the ibc denom coin
			if strings.HasPrefix(c.Denom, "ibc/") {
				suite.Equal(fundsToSend.Amount, c.Amount)
				found = true
			}
		}
		return found
	}, 15*time.Second, 1*time.Second)
}
