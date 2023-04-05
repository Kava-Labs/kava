package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authzcodec "github.com/cosmos/cosmos-sdk/x/authz/codec"
)

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

// RegisterLegacyAminoCodec registers all the necessary types and interfaces for the
// issuance module.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgIssueTokens{}, "issuance/MsgIssueTokens", nil)
	cdc.RegisterConcrete(&MsgRedeemTokens{}, "issuance/MsgRedeemTokens", nil)
	cdc.RegisterConcrete(&MsgBlockAddress{}, "issuance/MsgBlockAddress", nil)
	cdc.RegisterConcrete(&MsgUnblockAddress{}, "issuance/MsgUnblockAddress", nil)
	cdc.RegisterConcrete(&MsgSetPauseStatus{}, "issuance/MsgChangePauseStatus", nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgIssueTokens{},
		&MsgRedeemTokens{},
		&MsgBlockAddress{},
		&MsgUnblockAddress{},
		&MsgSetPauseStatus{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)

	// Register all Amino interfaces and concrete types on the authz Amino codec so that this can later be
	// used to properly serialize MsgGrant and MsgExec instances
	RegisterLegacyAminoCodec(authzcodec.Amino)
}
