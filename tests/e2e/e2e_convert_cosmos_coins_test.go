package e2e_test

import (
	"context"
	"math/big"
	"time"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/ethereum/go-ethereum"
	"github.com/kava-labs/kava/tests/e2e/testutil"
	"github.com/kava-labs/kava/tests/util"
	evmutiltypes "github.com/kava-labs/kava/x/evmutil/types"
)

func setupConvertToCoinTest(
	suite *IntegrationTestSuite, accountName string,
) (denom string, initialFunds sdk.Coins, user *testutil.SigningAccount) {
	// we expect a denom to be registered to the allowed denoms param
	// and for the funded account to have a balance for that denom
	params, err := suite.Kava.Evmutil.Params(context.Background(), &evmutiltypes.QueryParamsRequest{})
	suite.NoError(err)
	suite.GreaterOrEqual(
		len(params.Params.AllowedCosmosDenoms), 1,
		"kava chain expected to have at least one AllowedCosmosDenom for ERC20 conversion",
	)

	tokenInfo := params.Params.AllowedCosmosDenoms[0]
	denom = tokenInfo.CosmosDenom
	initialFunds = sdk.NewCoins(
		sdk.NewInt64Coin(suite.Kava.StakingDenom, 1e6), // gas money
		sdk.NewInt64Coin(denom, 1e10),                  // conversion-enabled cosmos coin
	)

	user = suite.Kava.NewFundedAccount(accountName, initialFunds)

	return denom, initialFunds, user
}

func (suite *IntegrationTestSuite) TestConvertCosmosCoinsToFromERC20() {
	denom, initialFunds, user := setupConvertToCoinTest(suite, "cosmo-coin-converter")

	fee := sdk.NewCoins(ukava(7500))
	convertAmount := int64(5e9)
	initialModuleBalance := suite.Kava.GetModuleBalances(evmutiltypes.ModuleName).AmountOf(denom)

	///////////////////////////////
	// CONVERT COSMOS COIN -> ERC20
	///////////////////////////////
	convertToErc20Msg := evmutiltypes.NewMsgConvertCosmosCoinToERC20(
		user.SdkAddress.String(),
		user.EvmAddress.Hex(),
		sdk.NewInt64Coin(denom, convertAmount),
	)
	tx := util.KavaMsgRequest{
		Msgs:      []sdk.Msg{&convertToErc20Msg},
		GasLimit:  2e6,
		FeeAmount: fee,
		Data:      "converting sdk coin to erc20",
	}
	res := user.SignAndBroadcastKavaTx(tx)
	suite.NoError(res.Err)

	// query for the deployed contract
	deployedContracts, err := suite.Kava.Evmutil.DeployedCosmosCoinContracts(
		context.Background(),
		&evmutiltypes.QueryDeployedCosmosCoinContractsRequest{CosmosDenoms: []string{denom}},
	)
	suite.NoError(err)
	suite.Len(deployedContracts.DeployedCosmosCoinContracts, 1)

	contractAddress := deployedContracts.DeployedCosmosCoinContracts[0].Address

	// check erc20 balance
	bz, err := suite.Kava.EvmClient.CallContract(context.Background(), ethereum.CallMsg{
		To:   &contractAddress.Address,
		Data: util.BuildErc20BalanceOfCallData(user.EvmAddress),
	}, nil)
	suite.NoError(err)
	erc20Balance := new(big.Int).SetBytes(bz)
	suite.BigIntsEqual(big.NewInt(convertAmount), erc20Balance, "unexpected erc20 balance post-convert")

	// check cosmos coin is deducted from account
	expectedFunds := initialFunds.AmountOf(denom).SubRaw(convertAmount)
	balance := suite.Kava.QuerySdkForBalances(user.SdkAddress).AmountOf(denom)
	suite.Equal(expectedFunds, balance)

	// check that module account has sdk coins
	expectedModuleBalance := initialModuleBalance.AddRaw(convertAmount)
	actualModuleBalance := suite.Kava.GetModuleBalances(evmutiltypes.ModuleName).AmountOf(denom)
	suite.Equal(expectedModuleBalance, actualModuleBalance)

	///////////////////////////////
	// CONVERT ERC20 -> COSMOS COIN
	///////////////////////////////
	convertFromErc20Msg := evmutiltypes.NewMsgConvertCosmosCoinFromERC20(
		user.EvmAddress.Hex(),
		user.SdkAddress.String(),
		sdk.NewInt64Coin(denom, convertAmount),
	)

	tx = util.KavaMsgRequest{
		Msgs:      []sdk.Msg{&convertFromErc20Msg},
		GasLimit:  2e5,
		FeeAmount: fee,
		Data:      "converting erc20 to cosmos coin",
	}
	res = user.SignAndBroadcastKavaTx(tx)
	suite.NoError(res.Err)

	// check erc20 balance
	bz, err = suite.Kava.EvmClient.CallContract(context.Background(), ethereum.CallMsg{
		To:   &contractAddress.Address,
		Data: util.BuildErc20BalanceOfCallData(user.EvmAddress),
	}, nil)
	suite.NoError(err)
	erc20Balance = new(big.Int).SetBytes(bz)
	suite.BigIntsEqual(big.NewInt(0), erc20Balance, "expected all erc20 to be converted back")

	// check cosmos coin is added back to account
	expectedFunds = initialFunds.AmountOf(denom)
	balance = suite.Kava.QuerySdkForBalances(user.SdkAddress).AmountOf(denom)
	suite.Equal(expectedFunds, balance)

	// check that module account has sdk coins deducted
	actualModuleBalance = suite.Kava.GetModuleBalances(evmutiltypes.ModuleName).AmountOf(denom)
	suite.Equal(initialModuleBalance, actualModuleBalance)
}

// like the above, but confirming we can sign the messages with eip712
func (suite *IntegrationTestSuite) TestEIP712ConvertCosmosCoinsToFromERC20() {
	denom, initialFunds, user := setupConvertToCoinTest(suite, "cosmo-coin-converter-eip712")

	convertAmount := int64(5e9)
	initialModuleBalance := suite.Kava.GetModuleBalances(evmutiltypes.ModuleName).AmountOf(denom)

	///////////////////////////////
	// CONVERT COSMOS COIN -> ERC20
	///////////////////////////////
	convertToErc20Msg := evmutiltypes.NewMsgConvertCosmosCoinToERC20(
		user.SdkAddress.String(),
		user.EvmAddress.Hex(),
		sdk.NewInt64Coin(denom, convertAmount),
	)
	tx := suite.NewEip712TxBuilder(
		user,
		suite.Kava,
		2e6,
		sdk.NewCoins(ukava(1e4)),
		[]sdk.Msg{&convertToErc20Msg},
		"this is a memo",
	).GetTx()
	txBytes, err := suite.Kava.EncodingConfig.TxConfig.TxEncoder()(tx)
	suite.NoError(err)

	// submit the eip712 message to the chain.
	res, err := suite.Kava.Tx.BroadcastTx(context.Background(), &txtypes.BroadcastTxRequest{
		TxBytes: txBytes,
		Mode:    txtypes.BroadcastMode_BROADCAST_MODE_SYNC,
	})
	suite.NoError(err)
	suite.Equal(sdkerrors.SuccessABCICode, res.TxResponse.Code)

	_, err = util.WaitForSdkTxCommit(suite.Kava.Tx, res.TxResponse.TxHash, 6*time.Second)
	suite.NoError(err)

	// query for the deployed contract
	deployedContracts, err := suite.Kava.Evmutil.DeployedCosmosCoinContracts(
		context.Background(),
		&evmutiltypes.QueryDeployedCosmosCoinContractsRequest{CosmosDenoms: []string{denom}},
	)
	suite.NoError(err)
	suite.Len(deployedContracts.DeployedCosmosCoinContracts, 1)

	contractAddress := deployedContracts.DeployedCosmosCoinContracts[0].Address

	// check erc20 balance
	bz, err := suite.Kava.EvmClient.CallContract(context.Background(), ethereum.CallMsg{
		To:   &contractAddress.Address,
		Data: util.BuildErc20BalanceOfCallData(user.EvmAddress),
	}, nil)
	suite.NoError(err)
	erc20Balance := new(big.Int).SetBytes(bz)
	suite.BigIntsEqual(big.NewInt(convertAmount), erc20Balance, "unexpected erc20 balance post-convert")

	// check cosmos coin is deducted from account
	expectedFunds := initialFunds.AmountOf(denom).SubRaw(convertAmount)
	balance := suite.Kava.QuerySdkForBalances(user.SdkAddress).AmountOf(denom)
	suite.Equal(expectedFunds, balance)

	// check that module account has sdk coins
	expectedModuleBalance := initialModuleBalance.AddRaw(convertAmount)
	actualModuleBalance := suite.Kava.GetModuleBalances(evmutiltypes.ModuleName).AmountOf(denom)
	suite.Equal(expectedModuleBalance, actualModuleBalance)

	///////////////////////////////
	// CONVERT ERC20 -> COSMOS COIN
	///////////////////////////////
	convertFromErc20Msg := evmutiltypes.NewMsgConvertCosmosCoinFromERC20(
		user.EvmAddress.Hex(),
		user.SdkAddress.String(),
		sdk.NewInt64Coin(denom, convertAmount),
	)
	tx = suite.NewEip712TxBuilder(
		user,
		suite.Kava,
		2e5,
		sdk.NewCoins(ukava(1e4)),
		[]sdk.Msg{&convertFromErc20Msg},
		"",
	).GetTx()
	txBytes, err = suite.Kava.EncodingConfig.TxConfig.TxEncoder()(tx)
	suite.NoError(err)

	// submit the eip712 message to the chain
	res, err = suite.Kava.Tx.BroadcastTx(context.Background(), &txtypes.BroadcastTxRequest{
		TxBytes: txBytes,
		Mode:    txtypes.BroadcastMode_BROADCAST_MODE_SYNC,
	})
	suite.NoError(err)
	suite.Equal(sdkerrors.SuccessABCICode, res.TxResponse.Code)

	_, err = util.WaitForSdkTxCommit(suite.Kava.Tx, res.TxResponse.TxHash, 6*time.Second)
	suite.NoError(err)

	// check erc20 balance
	bz, err = suite.Kava.EvmClient.CallContract(context.Background(), ethereum.CallMsg{
		To:   &contractAddress.Address,
		Data: util.BuildErc20BalanceOfCallData(user.EvmAddress),
	}, nil)
	suite.NoError(err)
	erc20Balance = new(big.Int).SetBytes(bz)
	suite.BigIntsEqual(big.NewInt(0), erc20Balance, "expected all erc20 to be converted back")

	// check cosmos coin is added back to account
	expectedFunds = initialFunds.AmountOf(denom)
	balance = suite.Kava.QuerySdkForBalances(user.SdkAddress).AmountOf(denom)
	suite.Equal(expectedFunds, balance)

	// check that module account has sdk coins deducted
	actualModuleBalance = suite.Kava.GetModuleBalances(evmutiltypes.ModuleName).AmountOf(denom)
	suite.Equal(initialModuleBalance, actualModuleBalance)
}
