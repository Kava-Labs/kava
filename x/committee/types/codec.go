package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// ModuleCdc is a generic codec to be used throughout module
var ModuleCdc *codec.Codec

func init() {
	cdc := codec.New()
	RegisterCodec(cdc)
	ModuleCdc = cdc
	// ModuleCdc is not sealed so that other modules can register their own pubproposal and/or permission types.

	// Register external module pubproposal types. Ideally these would be registered within the modules' types pkg init function.
	// However registration happens here as a work-around.
	RegisterProposalTypeCodec(distrtypes.CommunityPoolSpendProposal{}, "cosmos-sdk/CommunityPoolSpendProposal")
	RegisterProposalTypeCodec(paramstypes.ParameterChangeProposal{}, "cosmos-sdk/ParameterChangeProposal")
	RegisterProposalTypeCodec(govtypes.TextProposal{}, "cosmos-sdk/TextProposal")
}

// RegisterCodec registers the necessary types for the module
func RegisterCodec(cdc *codec.Codec) {

	// Proposals
	cdc.RegisterInterface((*PubProposal)(nil), nil)
	cdc.RegisterConcrete(CommitteeChangeProposal{}, "kava/CommitteeChangeProposal", nil)
	cdc.RegisterConcrete(CommitteeDeleteProposal{}, "kava/CommitteeDeleteProposal", nil)

	// Permissions
	cdc.RegisterInterface((*Permission)(nil), nil)
	cdc.RegisterConcrete(GodPermission{}, "kava/GodPermission", nil)
	cdc.RegisterConcrete(ParamChangePermission{}, "kava/ParamChangePermission", nil)
	cdc.RegisterConcrete(TextPermission{}, "kava/TextPermission", nil)
	cdc.RegisterConcrete(SubParamChangePermission{}, "kava/SubParamChangePermission", nil)

	// Msgs
	cdc.RegisterConcrete(MsgSubmitProposal{}, "kava/MsgSubmitProposal", nil)
	cdc.RegisterConcrete(MsgVote{}, "kava/MsgVote", nil)
}

// RegisterPermissionTypeCodec allows external modules to register their own permission types on
// the internal ModuleCdc. This allows the MsgSubmitProposal to be correctly Amino encoded and
// decoded (when the msg contains a CommitteeChangeProposal).
func RegisterPermissionTypeCodec(o interface{}, name string) {
	ModuleCdc.RegisterConcrete(o, name, nil)
}

// RegisterProposalTypeCodec allows external modules to register their own pubproposal types on the
// internal ModuleCdc. This allows the MsgSubmitProposal to be correctly Amino encoded and decoded.
func RegisterProposalTypeCodec(o interface{}, name string) {
	ModuleCdc.RegisterConcrete(o, name, nil)
}
