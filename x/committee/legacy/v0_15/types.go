package v0_15

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// GenesisState is state that must be provided at chain genesis.
type GenesisState struct {
	NextProposalID uint64     `json:"next_proposal_id" yaml:"next_proposal_id"`
	Committees     Committees `json:"committees" yaml:"committees"`
	Proposals      []Proposal `json:"proposals" yaml:"proposals"`
	Votes          []Vote     `json:"votes" yaml:"votes"`
}

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Proposals
	cdc.RegisterInterface((*PubProposal)(nil), nil)

	// Committees
	cdc.RegisterInterface((*Committee)(nil), nil)
	cdc.RegisterConcrete(BaseCommittee{}, "kava/BaseCommittee", nil)
	cdc.RegisterConcrete(MemberCommittee{}, "kava/MemberCommittee", nil)
	cdc.RegisterConcrete(TokenCommittee{}, "kava/TokenCommittee", nil)

	// Permissions
	cdc.RegisterInterface((*Permission)(nil), nil)
	cdc.RegisterConcrete(GodPermission{}, "kava/GodPermission", nil)
	cdc.RegisterConcrete(SimpleParamChangePermission{}, "kava/SimpleParamChangePermission", nil)
	cdc.RegisterConcrete(TextPermission{}, "kava/TextPermission", nil)
	cdc.RegisterConcrete(SoftwareUpgradePermission{}, "kava/SoftwareUpgradePermission", nil)
	cdc.RegisterConcrete(SubParamChangePermission{}, "kava/SubParamChangePermission", nil)
}
