package testutil

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/ethermint/ethereum/eip712"
	emtypes "github.com/evmos/ethermint/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

func (suite *E2eTestSuite) NewEip712TxBuilder(
	acc *SigningAccount, chain *Chain, gas uint64, gasAmount sdk.Coins, msgs []sdk.Msg, memo string,
) client.TxBuilder {
	// get account details
	var accDetails authtypes.AccountI
	a, err := chain.Auth.Account(context.Background(), &authtypes.QueryAccountRequest{
		Address: acc.SdkAddress.String(),
	})
	suite.NoError(err)
	err = chain.EncodingConfig.InterfaceRegistry.UnpackAny(a.Account, &accDetails)
	suite.NoError(err)

	// get nonce & acc number
	nonce := accDetails.GetSequence()
	accNumber := accDetails.GetAccountNumber()

	// get chain id
	pc, err := emtypes.ParseChainID(chain.ChainId)
	suite.NoError(err)
	ethChainId := pc.Uint64()

	evmParams, err := chain.Evm.Params(context.Background(), &evmtypes.QueryParamsRequest{})
	suite.NoError(err)

	fee := legacytx.NewStdFee(gas, gasAmount)

	// build EIP712 tx
	// -- untyped data
	untypedData := eip712.ConstructUntypedEIP712Data(
		chain.ChainId,
		accNumber,
		nonce,
		0, // no timeout
		fee,
		msgs,
		memo,
		nil,
	)
	// -- typed data
	typedData, err := eip712.WrapTxToTypedData(ethChainId, msgs, untypedData, &eip712.FeeDelegationOptions{
		FeePayer: acc.SdkAddress,
	}, evmParams.Params)
	suite.NoError(err)

	// -- raw data hash!
	data, err := eip712.ComputeTypedDataHash(typedData)
	suite.NoError(err)

	// -- sign the hash
	signature, pubKey, err := acc.SignRawEvmData(data)
	suite.NoError(err)
	signature[crypto.RecoveryIDOffset] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper

	// add ExtensionOptionsWeb3Tx extension
	var option *codectypes.Any
	option, err = codectypes.NewAnyWithValue(&emtypes.ExtensionOptionsWeb3Tx{
		FeePayer:         acc.SdkAddress.String(),
		TypedDataChainID: ethChainId,
		FeePayerSig:      signature,
	})
	suite.NoError(err)

	// create cosmos sdk tx builder
	txBuilder := chain.EncodingConfig.TxConfig.NewTxBuilder()
	builder, ok := txBuilder.(authtx.ExtensionOptionsTxBuilder)
	suite.True(ok)

	builder.SetExtensionOptions(option)
	builder.SetFeeAmount(fee.Amount)
	builder.SetGasLimit(fee.Gas)

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

	builder.SetMemo(memo)

	return builder
}
