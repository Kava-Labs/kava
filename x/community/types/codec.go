package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterLegacyAminoCodec registers all the necessary types and interfaces for the module.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgFundCommunityPool{}, "community/MsgFundCommunityPool", nil)
	cdc.RegisterConcrete(&CommunityPoolLendDepositProposal{}, "kava/CommunityPoolLendDepositProposal", nil)
	cdc.RegisterConcrete(&CommunityPoolLendWithdrawProposal{}, "kava/CommunityPoolLendWithdrawProposal", nil)
	cdc.RegisterConcrete(&CommunityPoolProposal{}, "kava/CommunityPoolProposal", nil)
}

// RegisterInterfaces registers proto messages under their interfaces for unmarshalling,
// in addition to registering the msg service for handling tx msgs.
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*govtypes.Content)(nil),
		&CommunityPoolProposal{},
	)

	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgFundCommunityPool{},
	)
	registry.RegisterImplementations((*govtypes.Content)(nil),
		&CommunityPoolLendDepositProposal{},
		&CommunityPoolLendWithdrawProposal{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
}
