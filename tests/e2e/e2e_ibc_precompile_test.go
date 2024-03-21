package e2e_test

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/kava-labs/kava/contracts/contracts/example_ibc"
	"github.com/kava-labs/kava/tests/e2e/testutil"
)

// TestIBCPrecompileTransferKava transfers KAVA from the primary kava chain (suite.Kava) to the ibc chain (suite.Ibc).
// Note that because the IBC chain also runs kava's binary, this tests both the sending & receiving.
func (suite *IntegrationTestSuite) TestIBCPrecompileTransferKava() {
	suite.SkipIfIbcDisabled()

	// ARRANGE
	// setup helper account
	helperAcc := suite.Kava.NewFundedAccount("ibc-precompile-transfer-kava-helper-account", sdk.NewCoins(ukava(1e6))) // 1 KAVA
	// setup kava account
	funds := ukava(1e5) // .1 KAVA
	kavaAcc := suite.Kava.NewFundedAccount("ibc-precompile-transfer-kava-kava-side", sdk.NewCoins(funds))
	// setup ibc account
	ibcAcc := suite.Ibc.NewFundedAccount("ibc-precompile-transfer-kava-ibc-side", sdk.NewCoins())

	ethClient, err := ethclient.Dial("http://127.0.0.1:8545")
	suite.Require().NoError(err)

	ibcExampleAddr, _, exampleIbc, err := example_ibc.DeployExampleIbc(helperAcc.EvmAuth, ethClient)
	suite.Require().NoError(err)

	// wait until contract is deployed
	suite.Eventually(func() bool {
		code, err := ethClient.CodeAt(context.Background(), ibcExampleAddr, nil)
		suite.Require().NoError(err)

		return len(code) != 0
	}, 20*time.Second, 1*time.Second)

	// send kava to IBCExample Contract Account
	suite.Kava.FundAccount(ibcExampleAddr[:], sdk.NewCoins(funds))
	// check that IBCExample Contract Account has enough kava to send it to IBC Account
	suite.Eventually(func() bool {
		balance := suite.Kava.QuerySdkForBalances(ibcExampleAddr[:]).AmountOf("ukava")
		return balance.Equal(funds.Amount)
	}, 10*time.Second, 1*time.Second)

	// Payment: Kava Account -> IBCExample Contract Account -> IBC Account
	// 1. blockchain fee is charged from Kava Account
	// 2. payment value is charged from IBCExample Contract Account
	fundsToSend := ukava(5e4) // .005 KAVA
	evmAuth := *kavaAcc.EvmAuth
	evmAuth.Value = fundsToSend.Amount.BigInt()
	_, err = exampleIbc.TransferKavaCall(
		&evmAuth,
		testutil.IbcPort,
		testutil.IbcChannel,
		ibcAcc.SdkAddress.String(),
		0,
		0,
		uint64(time.Now().Add(30*time.Second).UnixNano()),
		"test memo",
	)
	suite.Require().NoError(err)

	// blockchain fee should be deducted from Kava Account
	suite.Eventually(func() bool {
		// fee should be in range [1; 200]
		feeLimit := ukava(200)
		// source balance should be in range [funds - feeLimit; funds - 1]
		lowerBound := funds.Sub(feeLimit)
		upperBound := funds.Sub(ukava(1))

		balance := suite.Kava.QuerySdkForBalances(kavaAcc.SdkAddress).AmountOf("ukava")
		return balance.GTE(lowerBound.Amount) && balance.LTE(upperBound.Amount)
	}, 10*time.Second, 1*time.Second)

	// payment value should be deducted from IBCExample Contract Account
	suite.Eventually(func() bool {
		balance := suite.Kava.QuerySdkForBalances(ibcExampleAddr[:]).AmountOf("ukava")
		return balance.Equal(funds.Sub(fundsToSend).Amount)
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

// TestIBCPrecompileTransferCosmosDenom transfers KAVA from the primary kava chain (suite.Kava) to the ibc chain (suite.Ibc).
// Note that because the IBC chain also runs kava's binary, this tests both the sending & receiving.
func (suite *IntegrationTestSuite) TestIBCPrecompileTransferCosmosDenom() {
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

	ibcExampleAddr, _, exampleIbc, err := example_ibc.DeployExampleIbc(helperAcc.EvmAuth, ethClient)
	suite.Require().NoError(err)

	// wait until contract is deployed
	suite.Eventually(func() bool {
		code, err := ethClient.CodeAt(context.Background(), ibcExampleAddr, nil)
		suite.Require().NoError(err)

		return len(code) != 0
	}, 20*time.Second, 1*time.Second)

	// send kava to IBCExample Contract Account
	suite.Kava.FundAccount(ibcExampleAddr[:], sdk.NewCoins(funds))
	// check that IBCExample Contract Account has enough kava to send it to IBC Account
	suite.Eventually(func() bool {
		balance := suite.Kava.QuerySdkForBalances(ibcExampleAddr[:]).AmountOf("ukava")
		return balance.Equal(funds.Amount)
	}, 10*time.Second, 1*time.Second)

	// Payment: Kava Account -> IBCExample Contract Account -> IBC Account
	// 1. blockchain fee is charged from Kava Account
	// 2. payment value is charged from IBCExample Contract Account
	fundsToSend := ukava(5e4) // .005 KAVA
	_, err = exampleIbc.TransferCosmosDenomCall(
		kavaAcc.EvmAuth,
		testutil.IbcPort,
		testutil.IbcChannel,
		fundsToSend.Denom,
		fundsToSend.Amount.BigInt(),
		ibcAcc.SdkAddress.String(),
		0,
		0,
		uint64(time.Now().Add(30*time.Second).UnixNano()),
		"test memo",
	)
	suite.Require().NoError(err)

	// blockchain fee should be deducted from Kava Account
	suite.Eventually(func() bool {
		// fee should be in range [1; 200]
		feeLimit := ukava(200)
		// source balance should be in range [funds - feeLimit; funds - 1]
		lowerBound := funds.Sub(feeLimit)
		upperBound := funds.Sub(ukava(1))

		balance := suite.Kava.QuerySdkForBalances(kavaAcc.SdkAddress).AmountOf("ukava")
		return balance.GTE(lowerBound.Amount) && balance.LTE(upperBound.Amount)
	}, 10*time.Second, 1*time.Second)

	// payment value should be deducted from IBCExample Contract Account
	suite.Eventually(func() bool {
		balance := suite.Kava.QuerySdkForBalances(ibcExampleAddr[:]).AmountOf("ukava")
		return balance.Equal(funds.Sub(fundsToSend).Amount)
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

// TestIBCPrecompileTransferERC20 transfers KAVA from the primary kava chain (suite.Kava) to the ibc chain (suite.Ibc).
// Note that because the IBC chain also runs kava's binary, this tests both the sending & receiving.
func (suite *IntegrationTestSuite) TestIBCPrecompileTransferERC20() {
	suite.SkipIfIbcDisabled()

	// ARRANGE
	// setup helper account
	helperAcc := suite.Kava.NewFundedAccount("ibc-precompile-token-transfer-helper-account", sdk.NewCoins(ukava(1e6)))
	// setup kava account
	initialKavaFunds := ukava(1e6)
	kavaAcc := suite.Kava.NewFundedAccount("ibc-precompile-token-transfer-kava-side", sdk.NewCoins(initialKavaFunds))
	// setup ibc account
	ibcAcc := suite.Ibc.NewFundedAccount("ibc-precompile-token-transfer-ibc-side", sdk.NewCoins())

	ethClient, err := ethclient.Dial("http://127.0.0.1:8545")
	suite.Require().NoError(err)
	ibcExampleAddr, _, exampleIbc, err := example_ibc.DeployExampleIbc(helperAcc.EvmAuth, ethClient)
	suite.Require().NoError(err)
	// wait until contract is deployed
	suite.Eventually(func() bool {
		code, err := ethClient.CodeAt(context.Background(), ibcExampleAddr, nil)
		suite.Require().NoError(err)

		return len(code) != 0
	}, 20*time.Second, 1*time.Second)

	// TODO: delete me
	{
		precompileAddrHex := "0x0300000000000000000000000000000000000002"
		precompileAddr := common.HexToAddress(precompileAddrHex)

		code, err := ethClient.CodeAt(context.Background(), precompileAddr, nil)
		suite.Require().NoError(err)
		fmt.Printf("code before call: %v\n", code)
	}

	initialERC20Tokens := big.NewInt(1e5)
	fundRes := suite.FundKavaErc20Balance(ibcExampleAddr, initialERC20Tokens)
	suite.NoError(fundRes.Err)

	tokensToSend := big.NewInt(1e4)
	_, err = exampleIbc.TransferERC20Call(
		kavaAcc.EvmAuth,
		testutil.IbcPort,
		testutil.IbcChannel,
		tokensToSend,
		ibcAcc.SdkAddress.String(),
		0,
		0,
		uint64(time.Now().Add(30*time.Second).UnixNano()),
		"test memo",
		suite.DeployedErc20.Address.String(),
	)
	suite.Require().NoError(err)

	// TODO: delete me
	{
		precompileAddrHex := "0x0300000000000000000000000000000000000002"
		precompileAddr := common.HexToAddress(precompileAddrHex)

		code, err := ethClient.CodeAt(context.Background(), precompileAddr, nil)
		suite.Require().NoError(err)
		fmt.Printf("code after call: %v\n", code)
	}

	_, err = exampleIbc.TransferERC20(
		kavaAcc.EvmAuth,
		testutil.IbcPort,
		testutil.IbcChannel,
		tokensToSend,
		ibcAcc.SdkAddress.String(),
		0,
		0,
		uint64(time.Now().Add(30*time.Second).UnixNano()),
		"test memo",
		suite.DeployedErc20.Address.String(),
	)
	suite.Require().NoError(err)

	// blockchain fee should be deducted from Kava Account
	suite.Eventually(func() bool {
		// fee should be in range [1; 200]
		feeLimit := ukava(200)
		// source balance should be in range [funds - feeLimit; funds - 1]
		lowerBound := initialKavaFunds.Sub(feeLimit)
		upperBound := initialKavaFunds.Sub(ukava(1))

		balance := suite.Kava.QuerySdkForBalances(kavaAcc.SdkAddress).AmountOf("ukava")
		return balance.GTE(lowerBound.Amount) && balance.LTE(upperBound.Amount)
	}, 10*time.Second, 1*time.Second)

	// payment value should be deducted from IBCExample Contract Account
	suite.Eventually(func() bool {
		var expectedBalance big.Int
		expectedBalance.Sub(initialERC20Tokens, tokensToSend)

		balance := suite.Kava.GetErc20Balance(suite.DeployedErc20.Address, ibcExampleAddr)
		return balance.Cmp(&expectedBalance) == 0
	}, 10*time.Second, 1*time.Second)

	// expect the balance to be transferred to the ibc chain!
	suite.Eventually(func() bool {
		balance := suite.Ibc.QuerySdkForBalances(ibcAcc.SdkAddress)
		found := false
		for _, c := range balance {
			// find the ibc denom coin
			if strings.HasPrefix(c.Denom, "ibc/") {
				suite.Equal(sdkmath.NewIntFromBigInt(tokensToSend), c.Amount)
				found = true
			}
		}
		return found
	}, 15*time.Second, 1*time.Second)
}

// TestIBCPrecompileTransferERC20 transfers KAVA from the primary kava chain (suite.Kava) to the ibc chain (suite.Ibc).
// Note that because the IBC chain also runs kava's binary, this tests both the sending & receiving.
func (suite *IntegrationTestSuite) TestIBCPrecompileTransferERC20Call() {
	suite.SkipIfIbcDisabled()

	// ARRANGE
	// setup helper account
	helperAcc := suite.Kava.NewFundedAccount("ibc-precompile-token-transfer-helper-account", sdk.NewCoins(ukava(1e6)))
	// setup kava account
	initialKavaFunds := ukava(1e6)
	kavaAcc := suite.Kava.NewFundedAccount("ibc-precompile-token-transfer-kava-side", sdk.NewCoins(initialKavaFunds))
	// setup ibc account
	ibcAcc := suite.Ibc.NewFundedAccount("ibc-precompile-token-transfer-ibc-side", sdk.NewCoins())

	ethClient, err := ethclient.Dial("http://127.0.0.1:8545")
	suite.Require().NoError(err)
	ibcExampleAddr, _, exampleIbc, err := example_ibc.DeployExampleIbc(helperAcc.EvmAuth, ethClient)
	suite.Require().NoError(err)
	// wait until contract is deployed
	suite.Eventually(func() bool {
		code, err := ethClient.CodeAt(context.Background(), ibcExampleAddr, nil)
		suite.Require().NoError(err)

		return len(code) != 0
	}, 20*time.Second, 1*time.Second)

	initialERC20Tokens := big.NewInt(1e5)
	fundRes := suite.FundKavaErc20Balance(ibcExampleAddr, initialERC20Tokens)
	suite.NoError(fundRes.Err)

	tokensToSend := big.NewInt(1e4)
	_, err = exampleIbc.TransferERC20Call(
		kavaAcc.EvmAuth,
		testutil.IbcPort,
		testutil.IbcChannel,
		tokensToSend,
		ibcAcc.SdkAddress.String(),
		0,
		0,
		uint64(time.Now().Add(30*time.Second).UnixNano()),
		"test memo",
		suite.DeployedErc20.Address.String(),
	)
	suite.Require().NoError(err)

	// blockchain fee should be deducted from Kava Account
	suite.Eventually(func() bool {
		// fee should be in range [1; 200]
		feeLimit := ukava(200)
		// source balance should be in range [funds - feeLimit; funds - 1]
		lowerBound := initialKavaFunds.Sub(feeLimit)
		upperBound := initialKavaFunds.Sub(ukava(1))

		balance := suite.Kava.QuerySdkForBalances(kavaAcc.SdkAddress).AmountOf("ukava")
		return balance.GTE(lowerBound.Amount) && balance.LTE(upperBound.Amount)
	}, 10*time.Second, 1*time.Second)

	// payment value should be deducted from IBCExample Contract Account
	suite.Eventually(func() bool {
		var expectedBalance big.Int
		expectedBalance.Sub(initialERC20Tokens, tokensToSend)

		balance := suite.Kava.GetErc20Balance(suite.DeployedErc20.Address, ibcExampleAddr)
		return balance.Cmp(&expectedBalance) == 0
	}, 10*time.Second, 1*time.Second)

	// expect the balance to be transferred to the ibc chain!
	suite.Eventually(func() bool {
		balance := suite.Ibc.QuerySdkForBalances(ibcAcc.SdkAddress)
		found := false
		for _, c := range balance {
			// find the ibc denom coin
			if strings.HasPrefix(c.Denom, "ibc/") {
				suite.Equal(sdkmath.NewIntFromBigInt(tokensToSend), c.Amount)
				found = true
			}
		}
		return found
	}, 15*time.Second, 1*time.Second)
}
