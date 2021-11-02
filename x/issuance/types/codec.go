package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global x/gov module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/issuance and
	// defined at the application level.
	ModuleCdc = codec.NewAminoCodec(amino)
)

// RegisterLegacyAminoCodec registers all the necessary types and interfaces for the
// governance module.
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
}
