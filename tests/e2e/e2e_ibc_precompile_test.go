package e2e_test

import (
	"context"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/kava-labs/kava/contracts/contracts/example_ibc"
	"github.com/kava-labs/kava/tests/e2e/testutil"
)

// TestIbcTransfer transfers KAVA from the primary kava chain (suite.Kava) to the ibc chain (suite.Ibc).
// Note that because the IBC chain also runs kava's binary, this tests both the sending & receiving.
func (suite *IntegrationTestSuite) TestIbcPrecompileTransfer() {
	suite.SkipIfIbcDisabled()

	// ARRANGE
	// setup helper account
	helperAcc := suite.Kava.NewFundedAccount("ibc-precompile-transfer-helper-account", sdk.NewCoins(ukava(1e6))) // 1 KAVA
	// setup kava account
	funds := ukava(1e5) // .1 KAVA
	kavaAcc := suite.Kava.NewFundedAccount("ibc-precompile-transfer-kava-side", sdk.NewCoins(funds))
	// setup ibc account
	ibcAcc := suite.Ibc.NewFundedAccount("ibc-precompile-transfer-ibc-side", sdk.NewCoins())

	ethClient, err := ethclient.Dial("http://127.0.0.1:8545")
	suite.Require().NoError(err)

	address, _, exampleIbc, err := example_ibc.DeployExampleIbc(helperAcc.EvmAuth, ethClient)
	suite.Require().NoError(err)

	// wait until contract is deployed
	suite.Eventually(func() bool {
		code, err := ethClient.CodeAt(context.Background(), address, nil)
		suite.Require().NoError(err)

		return len(code) != 0
	}, 20*time.Second, 1*time.Second)

	fundsToSend := ukava(5e4) // .005 KAVA
	_, err = exampleIbc.IbcTransferCall(
		kavaAcc.EvmAuth,
		testutil.IbcPort,
		testutil.IbcChannel,
		fundsToSend.Denom,
		fundsToSend.Amount.BigInt(),
		kavaAcc.SdkAddress.String(),
		ibcAcc.SdkAddress.String(),
		0,
		0,
		uint64(time.Now().Add(30*time.Second).UnixNano()),
	)
	suite.Require().NoError(err)

	// the balance should be deducted from kava account
	suite.Eventually(func() bool {
		// fee should be in range [0; 200]
		feeLimit := ukava(200)
		// source balance should be in range [funds - fundsToSend - feeLimit; funds - fundsToSend]
		lowerBound := funds.Sub(fundsToSend).Sub(feeLimit)
		upperBound := funds.Sub(fundsToSend)

		balance := suite.Kava.QuerySdkForBalances(kavaAcc.SdkAddress).AmountOf("ukava")
		return balance.GTE(lowerBound.Amount) && balance.LTE(upperBound.Amount)
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
