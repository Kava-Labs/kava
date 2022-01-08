package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers all the necessary types and interfaces for the
// governance module.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgClaimUSDXMintingReward{}, "incentive/MsgClaimUSDXMintingReward", nil)
	cdc.RegisterConcrete(&MsgClaimHardReward{}, "incentive/MsgClaimHardReward", nil)
	cdc.RegisterConcrete(&MsgClaimDelegatorReward{}, "incentive/MsgClaimDelegatorReward", nil)
	cdc.RegisterConcrete(&MsgClaimSwapReward{}, "incentive/MsgClaimSwapReward", nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgClaimUSDXMintingReward{},
		&MsgClaimHardReward{},
		&MsgClaimDelegatorReward{},
		&MsgClaimSwapReward{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino = codec.NewLegacyAmino()

	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
}
