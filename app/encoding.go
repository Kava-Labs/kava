package app

import (
	"fmt"
	enccodec "github.com/evmos/ethermint/encoding/codec"

	"github.com/kava-labs/kava/app/params"
)

// MakeEncodingConfig creates an EncodingConfig and registers the app's types on it.
func MakeEncodingConfig() params.EncodingConfig {
	fmt.Println("MakeEncodingConfig inside app")
	encodingConfig := params.MakeEncodingConfig()
	fmt.Println("encodingConfig: ", encodingConfig)
	enccodec.RegisterLegacyAminoCodec(encodingConfig.Amino)
	enccodec.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	//ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	//ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}

// cdc := amino.NewLegacyAmino()
//	signingOptions := signing.Options{
//		AddressCodec: address.Bech32Codec{
//			Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix(),
//		},
//		ValidatorAddressCodec: address.Bech32Codec{
//			Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix(),
//		},
//		CustomGetSigners: map[protoreflect.FullName]signing.GetSignersFunc{
//			evmtypes.MsgEthereumTxCustomGetSigner.MsgType:     evmtypes.MsgEthereumTxCustomGetSigner.Fn,
//			erc20types.MsgConvertERC20CustomGetSigner.MsgType: erc20types.MsgConvertERC20CustomGetSigner.Fn,
//		},
//	}
//
//	interfaceRegistry, _ := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
//		ProtoFiles:     proto.HybridResolver,
//		SigningOptions: signingOptions,
//	})
//	codec := amino.NewProtoCodec(interfaceRegistry)
//	enccodec.RegisterLegacyAminoCodec(cdc)
//	enccodec.RegisterInterfaces(interfaceRegistry)
//
//	// This is needed for the EIP712 txs because currently is using
//	// the deprecated method legacytx.StdSignBytes
//	legacytx.RegressionTestingAminoCodec = cdc
//	eip712.SetEncodingConfig(cdc, interfaceRegistry)
//
//	return sdktestutil.TestEncodingConfig{
//		InterfaceRegistry: interfaceRegistry,
//		Codec:             codec,
//		TxConfig:          tx.NewTxConfig(codec, tx.DefaultSignModes),
//		Amino:             cdc,
//	}
