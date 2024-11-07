package params

import (
	"cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/gogoproto/proto"
	enccodec "github.com/evmos/ethermint/encoding/codec"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	communitytypes "github.com/kava-labs/kava/x/community/types"
	evmutiltypes "github.com/kava-labs/kava/x/evmutil/types"
	"google.golang.org/protobuf/reflect/protoreflect"
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
	signingOptions := signing.Options{
		AddressCodec: address.Bech32Codec{
			Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix(),
		},
		ValidatorAddressCodec: address.Bech32Codec{
			Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix(),
		},
		CustomGetSigners: map[protoreflect.FullName]signing.GetSignersFunc{
			evmtypes.MsgEthereumTxGetSigner.MsgType:                      evmtypes.MsgEthereumTxGetSigner.Fn,
			evmutiltypes.MsgConvertCoinToERC20GetSigners.MsgType:         evmutiltypes.MsgConvertCoinToERC20GetSigners.Fn,
			evmutiltypes.MsgConvertERC20ToCoinGetSigners.MsgType:         evmutiltypes.MsgConvertERC20ToCoinGetSigners.Fn,
			evmutiltypes.MsgConvertCosmosCoinToERC20GetSigners.MsgType:   evmutiltypes.MsgConvertCosmosCoinToERC20GetSigners.Fn,
			evmutiltypes.MsgConvertCosmosCoinFromERC20GetSigners.MsgType: evmutiltypes.MsgConvertCosmosCoinFromERC20GetSigners.Fn,
			cdptypes.MsgCreateCDPGetSigners.MsgType:                      cdptypes.MsgCreateCDPGetSigners.Fn,
			cdptypes.MsgDepositGetSigners.MsgType:                        cdptypes.MsgDepositGetSigners.Fn,
			cdptypes.MsgWithdrawGetSigners.MsgType:                       cdptypes.MsgWithdrawGetSigners.Fn,
			cdptypes.MsgDrawDebtGetSigners.MsgType:                       cdptypes.MsgDrawDebtGetSigners.Fn,
			cdptypes.MsgRepayDebtGetSigners.MsgType:                      cdptypes.MsgRepayDebtGetSigners.Fn,
			cdptypes.MsgLiquidateGetSigners.MsgType:                      cdptypes.MsgLiquidateGetSigners.Fn,
			// MsgFundCommunityPoolGetSigners
			// MsgUpdateParamsSigners
			communitytypes.MsgFundCommunityPoolGetSigners.MsgType: communitytypes.MsgFundCommunityPoolGetSigners.Fn,
			communitytypes.MsgUpdateParamsSigners.MsgType:         communitytypes.MsgUpdateParamsSigners.Fn,
		},
	}
	interfaceRegistry, _ := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles:     proto.HybridResolver,
		SigningOptions: signingOptions,
	})
	//interfaceRegistry := types.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	txCfg := tx.NewTxConfig(marshaler, tx.DefaultSignModes)
	if txCfg.SigningContext().Validate() != nil {
		panic("invalid tx signing context")
	}

	enccodec.RegisterLegacyAminoCodec(amino)
	enccodec.RegisterInterfaces(interfaceRegistry)

	legacytx.RegressionTestingAminoCodec = amino

	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         marshaler,
		TxConfig:          txCfg,
		Amino:             amino,
	}
}
