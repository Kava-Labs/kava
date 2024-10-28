package params

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
)

// EncodingConfig specifies the concrete encoding types to use for a given app.
// This is provided for compatibility between protobuf and amino implementations.
type EncodingConfig struct {
	InterfaceRegistry types.InterfaceRegistry
	Marshaler         codec.Codec
	TxConfig          client.TxConfig
	Amino             *codec.LegacyAmino
}

// MakeEncodingConfig creates a new EncodingConfig.
func MakeEncodingConfig() EncodingConfig {
	amino := codec.NewLegacyAmino()
	interfaceRegistry := types.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	txCfg := tx.NewTxConfig(marshaler, tx.DefaultSignModes)

	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         marshaler,
		TxConfig:          txCfg,
		Amino:             amino,
	}
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
