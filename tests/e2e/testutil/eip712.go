package testutil

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"strings"

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

func dataMismatchError(encType string, encValue interface{}) error {
	return fmt.Errorf("provided data '%v' doesn't match type '%s'", encValue, encType)
}

func EncodeData(typedData apitypes.TypedData, primaryType string, data map[string]interface{}, depth int) (hexutil.Bytes, error) {
	//if err := typedData.validate(); err != nil {
	//	return nil, err
	//}

	buffer := bytes.Buffer{}

	// Verify extra data
	if exp, got := len(typedData.Types[primaryType]), len(data); exp < got {
		return nil, fmt.Errorf("there is extra data provided in the message (%d < %d)", exp, got)
	}

	fmt.Println("data", data)

	// Add typehash
	buffer.Write(typedData.TypeHash(primaryType))

	// Add field contents. Structs and arrays have special handlers.
	for _, field := range typedData.Types[primaryType] {
		fmt.Println("field", field)
		encType := field.Type
		encValue := data[field.Name]
		if encType[len(encType)-1:] == "]" {
			arrayValue, ok := encValue.([]interface{})
			fmt.Println("arrayValue 1", arrayValue, ok)
			if !ok {
				return nil, dataMismatchError(encType, encValue)
			}

			arrayBuffer := bytes.Buffer{}
			parsedType := strings.Split(encType, "[")[0]
			for _, item := range arrayValue {
				if typedData.Types[parsedType] != nil {
					mapValue, ok := item.(map[string]interface{})
					fmt.Println("mapValue 1", mapValue, ok)
					if !ok {
						return nil, dataMismatchError(parsedType, item)
					}
					encodedData, err := typedData.EncodeData(parsedType, mapValue, depth+1)
					fmt.Println("encodedData 1", encodedData, err)
					if err != nil {
						return nil, err
					}
					arrayBuffer.Write(crypto.Keccak256(encodedData))
				} else {
					bytesValue, err := typedData.EncodePrimitiveValue(parsedType, item, depth)
					fmt.Println("bytesValue 1", bytesValue, err)
					if err != nil {
						return nil, err
					}
					arrayBuffer.Write(bytesValue)
				}
			}

			buffer.Write(crypto.Keccak256(arrayBuffer.Bytes()))
		} else if typedData.Types[field.Type] != nil {
			fmt.Println("typedData.Types[field.Type]", typedData.Types[field.Type])
			fmt.Println("encValue", encValue)
			mapValue, ok := encValue.(map[string]interface{})
			fmt.Println("mapValue 2", mapValue, ok)
			if !ok {
				return nil, dataMismatchError(encType, encValue)
			}
			encodedData, err := typedData.EncodeData(field.Type, mapValue, depth+1)
			fmt.Println("encodedData 2", encodedData, err)
			if err != nil {
				return nil, err
			}
			buffer.Write(crypto.Keccak256(encodedData))
		} else {
			byteValue, err := typedData.EncodePrimitiveValue(encType, encValue, depth)
			fmt.Println("byteValue 2", byteValue, err)
			if err != nil {
				return nil, err
			}
			buffer.Write(byteValue)
		}
	}
	return buffer.Bytes(), nil
}

// NewEip712TxBuilder is a helper method for creating an EIP712 signed tx
// A tx like this is what a user signing cosmos messages with Metamask would broadcast.
func (suite *E2eTestSuite) NewEip712TxBuilder(
	acc *SigningAccount, chain *Chain, gas uint64, gasAmount sdk.Coins, msgs []sdk.Msg, memo string,
) client.TxBuilder {
	// get account details
	var accDetails sdk.AccountI
	fmt.Println("going to fetch", acc.SdkAddress.String())
	a, err := chain.Grpc.Query.Auth.Account(context.Background(), &authtypes.QueryAccountRequest{
		Address: acc.SdkAddress.String(),
	})
	suite.NoError(err)
	fmt.Println("a.Account after check: ", a)
	err = chain.EncodingConfig.InterfaceRegistry.UnpackAny(a.Account, &accDetails)
	suite.NoError(err)

	// get nonce & acc number
	nonce := accDetails.GetSequence()
	accNumber := accDetails.GetAccountNumber()

	// get chain id
	pc, err := emtypes.ParseChainID(chain.ChainID)
	suite.NoError(err)
	ethChainId := pc.Uint64()

	evmParams, err := chain.Grpc.Query.Evm.Params(context.Background(), &evmtypes.QueryParamsRequest{})
	suite.NoError(err)

	fee := legacytx.NewStdFee(gas, gasAmount)

	// build EIP712 tx
	// -- untyped data
	untypedData := eip712.ConstructUntypedEIP712Data(
		chain.ChainID,
		accNumber,
		nonce,
		0, // no timeout
		fee,
		msgs,
		memo,
	)

	fmt.Println("ConstructUntypedEIP712Data untypedData: ", untypedData)

	// -- typed data
	typedData, err := eip712.WrapTxToTypedData(ethChainId, msgs, untypedData, &eip712.FeeDelegationOptions{
		FeePayer: acc.SdkAddress,
	}, evmParams.Params)
	suite.NoError(err)

	fmt.Println("got typedData", typedData)

	fmt.Println("ConstructUntypedEIP712Data typedData: ", typedData.Message)
	encodedData, err := EncodeData(typedData, typedData.PrimaryType, typedData.Message, 1)
	fmt.Println("ConstructUntypedEIP712Data encodedData: ", encodedData, err)

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
