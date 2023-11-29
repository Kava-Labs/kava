package e2e_test

import (
	"context"
	"fmt"
	"math/big"
	"time"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"

	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/kava-labs/kava/tests/e2e/testutil"
	"github.com/kava-labs/kava/tests/util"
	evmutiltypes "github.com/kava-labs/kava/x/evmutil/types"
)

const initialCosmosCoinConversionDenomFunds = int64(1e4)

func setupConvertToCoinTest(
	suite *IntegrationTestSuite, accountName string,
) (denom string, initialFunds sdk.Coins, user *testutil.SigningAccount) {
	// we expect a denom to be registered to the allowed denoms param
	// and for the funded account to have a balance for that denom
	params, err := suite.Kava.Grpc.Query.Evmutil.Params(
		context.Background(),
		&evmutiltypes.QueryParamsRequest{},
	)
	suite.NoError(err)
	suite.GreaterOrEqual(
		len(params.Params.AllowedCosmosDenoms), 1,
		"kava chain expected to have at least one AllowedCosmosDenom for ERC20 conversion",
	)

	tokenInfo := params.Params.AllowedCosmosDenoms[0]
	denom = tokenInfo.CosmosDenom
	initialFunds = sdk.NewCoins(
		sdk.NewInt64Coin(suite.Kava.StakingDenom, 1e5),                 // gas money
		sdk.NewInt64Coin(denom, initialCosmosCoinConversionDenomFunds), // conversion-enabled cosmos coin
	)

	user = suite.Kava.NewFundedAccount(accountName, initialFunds)

	return denom, initialFunds, user
}

// amount must be less than initial funds (initialCosmosCoinConversionDenomFunds)
func (suite *IntegrationTestSuite) setupAccountWithCosmosCoinERC20Balance(
	accountName string, amount int64,
) (user *testutil.SigningAccount, contractAddress *evmutiltypes.InternalEVMAddress, denom string, sdkBalance sdk.Coins) {
	if amount > initialCosmosCoinConversionDenomFunds {
		panic(fmt.Sprintf("test erc20 amount must be less than %d", initialCosmosCoinConversionDenomFunds))
	}

	denom, sdkBalance, user = setupConvertToCoinTest(suite, accountName)
	convertAmount := sdk.NewInt64Coin(denom, amount)

	// setup user to have erc20 balance
	msg := evmutiltypes.NewMsgConvertCosmosCoinToERC20(
		user.SdkAddress.String(),
		user.EvmAddress.Hex(),
		convertAmount,
	)
	tx := util.KavaMsgRequest{
		Msgs:      []sdk.Msg{&msg},
		GasLimit:  4e5,
		FeeAmount: sdk.NewCoins(ukava(400)),
		Data:      "converting sdk coin to erc20",
	}
	res := user.SignAndBroadcastKavaTx(tx)
	suite.NoError(res.Err)

	// adjust sdk balance
	sdkBalance = sdkBalance.Sub(convertAmount)

	// query for the deployed contract
	deployedContracts, err := suite.Kava.Grpc.Query.Evmutil.DeployedCosmosCoinContracts(
		context.Background(),
		&evmutiltypes.QueryDeployedCosmosCoinContractsRequest{CosmosDenoms: []string{denom}},
	)
	suite.NoError(err)
	suite.Len(deployedContracts.DeployedCosmosCoinContracts, 1)

	contractAddress = deployedContracts.DeployedCosmosCoinContracts[0].Address

	return user, contractAddress, denom, sdkBalance
}

func (suite *IntegrationTestSuite) TestConvertCosmosCoinsToFromERC20() {
	denom, initialFunds, user := setupConvertToCoinTest(suite, "cosmo-coin-converter")

	convertAmount := int64(5e3)
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
		FeeAmount: sdk.NewCoins(ukava(2000)),
		Data:      "converting sdk coin to erc20",
	}
	res := user.SignAndBroadcastKavaTx(tx)
	suite.NoError(res.Err)

	// query for the deployed contract
	deployedContracts, err := suite.Kava.Grpc.Query.Evmutil.DeployedCosmosCoinContracts(
		context.Background(),
		&evmutiltypes.QueryDeployedCosmosCoinContractsRequest{CosmosDenoms: []string{denom}},
	)
	suite.NoError(err)
	suite.Len(deployedContracts.DeployedCosmosCoinContracts, 1)

	contractAddress := deployedContracts.DeployedCosmosCoinContracts[0].Address

	// check erc20 balance
	erc20Balance := suite.Kava.GetErc20Balance(contractAddress.Address, user.EvmAddress)
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
		FeeAmount: sdk.NewCoins(ukava(200)),
		Data:      "converting erc20 to cosmos coin",
	}
	res = user.SignAndBroadcastKavaTx(tx)
	suite.NoError(res.Err)

	// check erc20 balance
	erc20Balance = suite.Kava.GetErc20Balance(contractAddress.Address, user.EvmAddress)
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

	convertAmount := int64(5e3)
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
	res, err := suite.Kava.Grpc.Query.Tx.BroadcastTx(context.Background(), &txtypes.BroadcastTxRequest{
		TxBytes: txBytes,
		Mode:    txtypes.BroadcastMode_BROADCAST_MODE_SYNC,
	})
	suite.NoError(err)
	suite.Equal(sdkerrors.SuccessABCICode, res.TxResponse.Code)

	_, err = util.WaitForSdkTxCommit(suite.Kava.Grpc.Query.Tx, res.TxResponse.TxHash, 12*time.Second)
	suite.Require().NoError(err)

	// query for the deployed contract
	deployedContracts, err := suite.Kava.Grpc.Query.Evmutil.DeployedCosmosCoinContracts(
		context.Background(),
		&evmutiltypes.QueryDeployedCosmosCoinContractsRequest{CosmosDenoms: []string{denom}},
	)
	suite.NoError(err)
	suite.Len(deployedContracts.DeployedCosmosCoinContracts, 1)

	contractAddress := deployedContracts.DeployedCosmosCoinContracts[0].Address

	// check erc20 balance
	erc20Balance := suite.Kava.GetErc20Balance(contractAddress.Address, user.EvmAddress)
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
		sdk.NewCoins(ukava(200)),
		[]sdk.Msg{&convertFromErc20Msg},
		"",
	).GetTx()
	txBytes, err = suite.Kava.EncodingConfig.TxConfig.TxEncoder()(tx)
	suite.NoError(err)

	// submit the eip712 message to the chain
	res, err = suite.Kava.Grpc.Query.Tx.BroadcastTx(context.Background(), &txtypes.BroadcastTxRequest{
		TxBytes: txBytes,
		Mode:    txtypes.BroadcastMode_BROADCAST_MODE_SYNC,
	})
	suite.NoError(err)
	suite.Equal(sdkerrors.SuccessABCICode, res.TxResponse.Code)

	_, err = util.WaitForSdkTxCommit(suite.Kava.Grpc.Query.Tx, res.TxResponse.TxHash, 6*time.Second)
	suite.NoError(err)

	// check erc20 balance
	erc20Balance = suite.Kava.GetErc20Balance(contractAddress.Address, user.EvmAddress)
	suite.BigIntsEqual(big.NewInt(0), erc20Balance, "expected all erc20 to be converted back")

	// check cosmos coin is added back to account
	expectedFunds = initialFunds.AmountOf(denom)
	balance = suite.Kava.QuerySdkForBalances(user.SdkAddress).AmountOf(denom)
	suite.Equal(expectedFunds, balance)

	// check that module account has sdk coins deducted
	actualModuleBalance = suite.Kava.GetModuleBalances(evmutiltypes.ModuleName).AmountOf(denom)
	suite.Equal(initialModuleBalance, actualModuleBalance)
}

func (suite *IntegrationTestSuite) TestConvertCosmosCoins_ForbiddenERC20Calls() {
	// give them erc20 balance so we know that's not preventing these method calls
	// this test sacrifices the 1 coin by never returning it to the sdk.
	user, contractAddress, _, _ := suite.setupAccountWithCosmosCoinERC20Balance("cosmo-coin-converter-unhappy", 1)

	suite.Run("users can't mint()", func() {
		data := util.BuildErc20MintCallData(user.EvmAddress, big.NewInt(1))
		nonce, err := user.NextNonce()
		suite.NoError(err)

		mintTx := util.EvmTxRequest{
			Tx: ethtypes.NewTx(
				&ethtypes.LegacyTx{
					Nonce:    nonce,
					GasPrice: minEvmGasPrice,
					Gas:      1e5,
					To:       &contractAddress.Address,
					Data:     data,
				},
			),
			Data: "attempting to mint, should fail",
		}
		res := user.SignAndBroadcastEvmTx(mintTx)

		suite.ErrorAs(res.Err, &util.ErrEvmTxFailed)
		// TODO: when traceTransactions enabled, read actual error log
		suite.ErrorContains(res.Err, "transaction was committed but failed. likely an execution revert by contract code")
	})

	suite.Run("users can't burn()", func() {
		data := util.BuildErc20BurnCallData(user.EvmAddress, big.NewInt(1))
		nonce, err := user.NextNonce()
		suite.NoError(err)

		burnTx := util.EvmTxRequest{
			Tx: ethtypes.NewTx(
				&ethtypes.LegacyTx{
					Nonce:    nonce,
					GasPrice: minEvmGasPrice,
					Gas:      1e5,
					To:       &contractAddress.Address,
					Data:     data,
				},
			),
			Data: "attempting to burn, should fail",
		}
		res := user.SignAndBroadcastEvmTx(burnTx)

		suite.ErrorAs(res.Err, &util.ErrEvmTxFailed)
		// TODO: when traceTransactions enabled, read actual error log
		suite.ErrorContains(res.Err, "transaction was committed but failed. likely an execution revert by contract code")
	})
}

// - check approval flow of erc20. alice approves bob to move their funds
// - check complex conversion flow. bob converts funds they receive on evm back to sdk.Coin
func (suite *IntegrationTestSuite) TestConvertCosmosCoins_ERC20Magic() {
	initialAliceAmount := int64(2e3)
	alice, contractAddress, denom, _ := suite.setupAccountWithCosmosCoinERC20Balance(
		"cosmo-coin-converter-complex-alice", initialAliceAmount,
	)

	gasMoney := sdk.NewCoins(ukava(1e5))
	bob := suite.Kava.NewFundedAccount("cosmo-coin-converter-complex-bob", gasMoney)
	amount := big.NewInt(1e3) // test assumes this is half of alice's balance.

	// bob can't move alice's funds
	nonce, err := bob.NextNonce()
	suite.NoError(err)
	transferFromTxData := &ethtypes.LegacyTx{
		Nonce:    nonce,
		GasPrice: minEvmGasPrice,
		Gas:      1e5,
		To:       &contractAddress.Address,
		Value:    &big.Int{},
		Data:     util.BuildErc20TransferFromCallData(alice.EvmAddress, bob.EvmAddress, amount),
	}
	transferTx := util.EvmTxRequest{
		Tx:   ethtypes.NewTx(transferFromTxData),
		Data: "bob can't move alice's funds, should fail",
	}
	res := bob.SignAndBroadcastEvmTx(transferTx)
	suite.ErrorAs(res.Err, &util.ErrEvmTxFailed)
	suite.ErrorContains(res.Err, "transaction was committed but failed. likely an execution revert by contract code")

	// approve bob to move alice's funds
	nonce, err = alice.NextNonce()
	suite.NoError(err)
	approveTx := util.EvmTxRequest{
		Tx: ethtypes.NewTx(&ethtypes.LegacyTx{
			Nonce:    nonce,
			GasPrice: minEvmGasPrice,
			Gas:      1e5,
			To:       &contractAddress.Address,
			Value:    &big.Int{},
			Data:     util.BuildErc20ApproveCallData(bob.EvmAddress, amount),
		}),
		Data: "alice approves bob to spend amount",
	}
	res = alice.SignAndBroadcastEvmTx(approveTx)
	suite.NoError(res.Err)

	// bob can't move more than alice allowed
	transferFromTxData.Data = util.BuildErc20TransferFromCallData(
		alice.EvmAddress, bob.EvmAddress, new(big.Int).Add(amount, big.NewInt(1)),
	)
	transferFromTxData.Nonce, err = bob.NextNonce()
	suite.NoError(err)
	transferTooMuchTx := util.EvmTxRequest{
		Tx:   ethtypes.NewTx(transferFromTxData),
		Data: "transferring more than approved, should fail",
	}
	res = bob.SignAndBroadcastEvmTx(transferTooMuchTx)
	suite.ErrorAs(res.Err, &util.ErrEvmTxFailed)
	suite.ErrorContains(res.Err, "transaction was committed but failed. likely an execution revert by contract code")

	// bob can move allowed amount
	transferFromTxData.Data = util.BuildErc20TransferFromCallData(alice.EvmAddress, bob.EvmAddress, amount)
	transferFromTxData.Nonce, err = bob.NextNonce()
	suite.NoError(err)
	transferJustRightTx := util.EvmTxRequest{
		Tx:   ethtypes.NewTx(transferFromTxData),
		Data: "bob transfers alice's funds, allowed because he's approved",
	}
	res = bob.SignAndBroadcastEvmTx(transferJustRightTx)
	suite.Require().NoError(res.Err)

	// alice should have amount deducted
	erc20Balance := suite.Kava.GetErc20Balance(contractAddress.Address, alice.EvmAddress)
	suite.BigIntsEqual(big.NewInt(initialAliceAmount-amount.Int64()), erc20Balance, "alice has unexpected erc20 balance")
	// bob should have amount added
	erc20Balance = suite.Kava.GetErc20Balance(contractAddress.Address, bob.EvmAddress)
	suite.BigIntsEqual(amount, erc20Balance, "bob has unexpected erc20 balance")

	// convert bob's new funds back to an sdk.Coin
	convertMsg := evmutiltypes.NewMsgConvertCosmosCoinFromERC20(
		bob.EvmAddress.Hex(),
		bob.SdkAddress.String(),
		sdk.NewInt64Coin(denom, amount.Int64()),
	)
	convertTx := util.KavaMsgRequest{
		Msgs:      []sdk.Msg{&convertMsg},
		GasLimit:  2e5,
		FeeAmount: sdk.NewCoins(ukava(200)),
		Data:      "bob converts his new erc20 to an sdk.Coin",
	}
	convertRes := bob.SignAndBroadcastKavaTx(convertTx)
	suite.NoError(convertRes.Err)

	// bob should have no more erc20 balance
	erc20Balance = suite.Kava.GetErc20Balance(contractAddress.Address, bob.EvmAddress)
	suite.BigIntsEqual(big.NewInt(0), erc20Balance, "expected no erc20 balance for bob")
	// bob should have sdk balance
	balance := suite.Kava.QuerySdkForBalances(bob.SdkAddress).AmountOf(denom)
	suite.Equal(sdk.NewIntFromBigInt(amount), balance)

	// alice should have the remaining balance
	erc20Balance = suite.Kava.GetErc20Balance(contractAddress.Address, alice.EvmAddress)
	suite.BigIntsEqual(amount, erc20Balance, "expected alice to have half initial funds remaining")

	// convert alice's remaining balance back to sdk coins
	convertMsg = evmutiltypes.NewMsgConvertCosmosCoinFromERC20(
		alice.EvmAddress.Hex(),
		alice.SdkAddress.String(),
		sdk.NewInt64Coin(denom, amount.Int64()),
	)
	convertRes = alice.SignAndBroadcastKavaTx(convertTx)
	suite.NoError(convertRes.Err)
}
