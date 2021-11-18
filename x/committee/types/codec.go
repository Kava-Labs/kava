package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	proposaltypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global x/committee module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/committee and
	// defined at the application level.
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	// amino is not sealed so that other modules can register their own pubproposal and/or permission types.

	// Register external module pubproposal types. Ideally these would be registered within the modules' types pkg init function.
	// However registration happens here as a work-around.
	RegisterProposalTypeCodec(distrtypes.CommunityPoolSpendProposal{}, "cosmos-sdk/CommunityPoolSpendProposal")
	RegisterProposalTypeCodec(proposaltypes.ParameterChangeProposal{}, "cosmos-sdk/ParameterChangeProposal")
	RegisterProposalTypeCodec(govtypes.TextProposal{}, "cosmos-sdk/TextProposal")
	RegisterProposalTypeCodec(upgradetypes.SoftwareUpgradeProposal{}, "cosmos-sdk/SoftwareUpgradeProposal")
	RegisterProposalTypeCodec(upgradetypes.CancelSoftwareUpgradeProposal{}, "cosmos-sdk/CancelSoftwareUpgradeProposal")
}

// RegisterLegacyAminoCodec registers all the necessary types and interfaces for the module.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Proposals
	cdc.RegisterInterface((*PubProposal)(nil), nil)
	cdc.RegisterConcrete(CommitteeChangeProposal{}, "kava/CommitteeChangeProposal", nil)
	cdc.RegisterConcrete(CommitteeDeleteProposal{}, "kava/CommitteeDeleteProposal", nil)

	// Committees
	cdc.RegisterInterface((*Committee)(nil), nil)
	cdc.RegisterConcrete(BaseCommittee{}, "kava/BaseCommittee", nil)
	cdc.RegisterConcrete(MemberCommittee{}, "kava/MemberCommittee", nil)
	cdc.RegisterConcrete(TokenCommittee{}, "kava/TokenCommittee", nil)

	// Permissions
	cdc.RegisterInterface((*Permission)(nil), nil)
	cdc.RegisterConcrete(GodPermission{}, "kava/GodPermission", nil)
	cdc.RegisterConcrete(TextPermission{}, "kava/TextPermission", nil)
	cdc.RegisterConcrete(SoftwareUpgradePermission{}, "kava/SoftwareUpgradePermission", nil)
	// cdc.RegisterConcrete(SubParamChangePermission{}, "kava/SubParamChangePermission", nil)

	// Msgs
	cdc.RegisterConcrete(MsgSubmitProposal{}, "kava/MsgSubmitProposal", nil)
	cdc.RegisterConcrete(MsgVote{}, "kava/MsgVote", nil)
}

// RegisterProposalTypeCodec allows external modules to register their own pubproposal types on the
// internal ModuleCdc. This allows the MsgSubmitProposal to be correctly Amino encoded and decoded.
func RegisterProposalTypeCodec(o interface{}, name string) {
	ModuleCdc.RegisterConcrete(o, name, nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterInterface(
		"kava.committee.v1beta1.Committee",
		(*Committee)(nil),
		// TODO: Might not need the base one since we just use token and member, add some tests to test this out.
		&BaseCommittee{},
		&TokenCommittee{},
		&MemberCommittee{},
	)

	registry.RegisterInterface(
		"kava.committee.v1beta1.Permission",
		(*Permission)(nil),
		&GodPermission{},
	)

	// Need to register PubProposal here since we use this as alias for the x/gov Content interface for all the proposal implementations used in this module.
	// Note that all proposals supported by x/committee needed to be registered here, including the proposals from x/gov.
	registry.RegisterInterface(
		"kava.committee.v1beta1.PubProposal",
		(*PubProposal)(nil),
		&Proposal{},
		&govtypes.TextProposal{},
		&CommitteeChangeProposal{},
		&CommitteeDeleteProposal{},
	)
}
