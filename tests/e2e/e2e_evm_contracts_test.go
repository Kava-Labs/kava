package e2e_test

import (
	"context"
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/ethermint/ethereum/eip712"
	emtypes "github.com/evmos/ethermint/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	tmmempool "github.com/tendermint/tendermint/mempool"

	"github.com/kava-labs/kava/app"
)

// func (suite *IntegrationTestSuite) TestEthCallToGreeterContract() {
// 	// this test manipulates state of the Greeter contract which means other tests shouldn't use it.

// 	// setup funded account to interact with contract
// 	user := suite.Kava.NewFundedAccount("greeter-contract-user", sdk.NewCoins(ukava(10e6)))

// 	greeterAddr := suite.Kava.ContractAddrs["greeter"]
// 	contract, err := greeter.NewGreeter(greeterAddr, suite.Kava.EvmClient)
// 	suite.NoError(err)

// 	beforeGreeting, err := contract.Greet(nil)
// 	suite.NoError(err)

// 	updatedGreeting := "look at me, using the evm"
// 	tx, err := contract.SetGreeting(user.EvmAuth, updatedGreeting)
// 	suite.NoError(err)

// 	_, err = util.WaitForEvmTxReceipt(suite.Kava.EvmClient, tx.Hash(), 10*time.Second)
// 	suite.NoError(err)

// 	afterGreeting, err := contract.Greet(nil)
// 	suite.NoError(err)

// 	suite.Equal("what's up!", beforeGreeting)
// 	suite.Equal(updatedGreeting, afterGreeting)
// }

// func (suite *IntegrationTestSuite) TestEthCallToErc20() {
// 	randoReceiver := util.SdkToEvmAddress(app.RandomAddress())
// 	amount := big.NewInt(1e6)

// 	// make unauthenticated eth_call query to check balance
// 	beforeBalance := suite.GetErc20Balance(randoReceiver)

// 	// make authenticate eth_call to transfer tokens
// 	res := suite.FundKavaErc20Balance(randoReceiver, amount)
// 	suite.NoError(res.Err)

// 	// make another unauthenticated eth_call query to check new balance
// 	afterBalance := suite.GetErc20Balance(randoReceiver)

// 	suite.BigIntsEqual(big.NewInt(0), beforeBalance, "expected before balance to be zero")
// 	suite.BigIntsEqual(amount, afterBalance, "unexpected post-transfer balance")
// }

func (suite *IntegrationTestSuite) TestEip712BasicMessageAuthorization() {
	// create new funded account
	depositor := suite.Kava.NewFundedAccount("eip712-msgSend", sdk.NewCoins(ukava(100e6)))

	// get account details
	var depositorAcc authtypes.AccountI
	a, err := suite.Kava.Auth.Account(context.Background(), &authtypes.QueryAccountRequest{
		Address: depositor.SdkAddress.String(),
	})
	suite.NoError(err)
	err = suite.Kava.EncodingConfig.InterfaceRegistry.UnpackAny(a.Account, &depositorAcc)
	suite.NoError(err)

	nonce := depositorAcc.GetSequence()

	// get data necessary for building eip712 tx
	ethChainId, err := emtypes.ParseChainID(suite.Kava.ChainId)
	suite.NoError(err)
	evmParams, err := suite.Kava.Evm.Params(context.Background(), &evmtypes.QueryParamsRequest{})
	suite.NoError(err)

	// setup message for sending 1KAVA to random receiver
	receiver := app.RandomAddress()
	msgs := []sdk.Msg{
		banktypes.NewMsgSend(depositor.SdkAddress, receiver, sdk.NewCoins(ukava(1e6))),
	}

	// build EIP712 tx
	// -- untyped data
	untypedData := eip712.ConstructUntypedEIP712Data(
		suite.Kava.ChainId,
		depositorAcc.GetAccountNumber(),
		nonce,
		0, // no timeout
		legacytx.NewStdFee(1e10, sdk.NewCoins(ukava(1e5))),
		msgs,
		"memo goes here",
		nil,
	)
	// -- typed data
	typedData, err := eip712.WrapTxToTypedData(ethChainId.Uint64(), msgs, untypedData, &eip712.FeeDelegationOptions{
		FeePayer: depositor.SdkAddress,
	}, evmParams.Params)

	fmt.Println("typed data: ", typedData)
	suite.NoError(err)

	// -- raw data hash!
	data, err := eip712.ComputeTypedDataHash(typedData)
	suite.NoError(err)

	// -- sign the hash
	signature, pubKey, err := depositor.SignRawEvmData(data)
	suite.NoError(err)
	signature[crypto.RecoveryIDOffset] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper

	// add ExtensionOptionsWeb3Tx extension
	var option *codectypes.Any
	option, err = codectypes.NewAnyWithValue(&emtypes.ExtensionOptionsWeb3Tx{
		FeePayer:         depositor.SdkAddress.String(),
		TypedDataChainID: ethChainId.Uint64(),
		FeePayerSig:      signature,
	})
	suite.NoError(err)

	// TODO: do i need to call this like in the ante test?
	// suite.Kava.EncodingConfig.TxConfig.SignModeHandler()

	// create cosmos sdk tx builder
	txBuilder := suite.Kava.EncodingConfig.TxConfig.NewTxBuilder()
	builder, ok := txBuilder.(authtx.ExtensionOptionsTxBuilder)
	suite.Require().True(ok)

	builder.SetExtensionOptions(option)
	builder.SetFeeAmount(sdk.NewCoins(ukava(1e4)))
	builder.SetGasLimit(2e6)

	sigsV2 := signing.SignatureV2{
		PubKey: pubKey,
		Data: &signing.SingleSignatureData{
			SignMode: signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
		},
		Sequence: nonce,
	}

	err = builder.SetSignatures(sigsV2)
	suite.Require().NoError(err)

	err = builder.SetMsgs(msgs...)
	suite.Require().NoError(err)

	// NOTE: the messages work as expected when passed directly to kava sdk.
	// tx := builder.GetTx()
	// res := depositor.SignAndBroadcastKavaTx(util.KavaMsgRequest{
	// 	Msgs:      tx.GetMsgs(),
	// 	GasLimit:  tx.GetGas(),
	// 	FeeAmount: tx.GetFee(),
	// 	Memo:      tx.GetMemo(),
	// 	Data:      "test eip but not really. do the messages work?",
	// })
	// fmt.Println("txhash! ", res.Result.TxHash)
	// suite.NoError(res.Err)

	fmt.Printf("tx: %+v\n", txBuilder.GetTx())

	txBytes, err := suite.Kava.EncodingConfig.TxConfig.TxEncoder()(builder.GetTx())
	suite.NoError(err)

	res, err := suite.Kava.Tx.BroadcastTx(context.Background(), &txtypes.BroadcastTxRequest{
		TxBytes: txBytes,
		Mode:    txtypes.BroadcastMode_BROADCAST_MODE_SYNC,
	})

	if tmmempool.IsPreCheckError(err) {
		fmt.Println("is a pre-check error!")
	}
	suite.NoError(err)
	suite.Equal(sdkerrors.SuccessABCICode, res.TxResponse.Code)

	fmt.Println("txhash! ", res.TxResponse.TxHash)

	// check that the message was processed & the kava is transferred.
	balRes, err := suite.Kava.Bank.Balance(context.Background(), &banktypes.QueryBalanceRequest{
		Address: receiver.String(),
		Denom:   "ukava",
	})
	suite.NoError(err)
	suite.Equal(sdk.NewInt(1e6), balRes.Balance.Amount)
}
