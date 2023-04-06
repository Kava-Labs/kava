package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authzcodec "github.com/cosmos/cosmos-sdk/x/authz/codec"
)

// RegisterLegacyAminoCodec registers all the necessary types and interfaces for the module.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgMintDeposit{}, "router/MsgMintDeposit", nil)
	cdc.RegisterConcrete(&MsgDelegateMintDeposit{}, "router/MsgDelegateMintDeposit", nil)
	cdc.RegisterConcrete(&MsgWithdrawBurn{}, "router/MsgWithdrawBurn", nil)
	cdc.RegisterConcrete(&MsgWithdrawBurnUndelegate{}, "router/MsgWithdrawBurnUndelegate", nil)
}

// RegisterInterfaces registers proto messages under their interfaces for unmarshalling,
// in addition to registering the msg service for handling tx msgs
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgMintDeposit{},
		&MsgDelegateMintDeposit{},
		&MsgWithdrawBurn{},
		&MsgWithdrawBurnUndelegate{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino = codec.NewLegacyAmino()
	// ModuleCdc represents the legacy amino codec for the module
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)

	// Register all Amino interfaces and concrete types on the authz Amino codec so that this can later be
	// used to properly serialize MsgGrant and MsgExec instances
	RegisterLegacyAminoCodec(authzcodec.Amino)
}
