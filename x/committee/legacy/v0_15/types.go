package v0_15

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState is state that must be provided at chain genesis.
type GenesisState struct {
	NextProposalID uint64     `json:"next_proposal_id" yaml:"next_proposal_id"`
	Committees     Committees `json:"committees" yaml:"committees"`
	Proposals      []Proposal `json:"proposals" yaml:"proposals"`
	Votes          []Vote     `json:"votes" yaml:"votes"`
}

// MsgSubmitProposal is used by committee members to create a new proposal that they can vote on.
type MsgSubmitProposal struct {
	PubProposal PubProposal    `json:"pub_proposal" yaml:"pub_proposal"`
	Proposer    sdk.AccAddress `json:"proposer" yaml:"proposer"`
	CommitteeID uint64         `json:"committee_id" yaml:"committee_id"`
}

// MsgVote is submitted by committee members to vote on proposals.
type MsgVote struct {
	ProposalID uint64         `json:"proposal_id" yaml:"proposal_id"`
	Voter      sdk.AccAddress `json:"voter" yaml:"voter"`
	VoteType   VoteType       `json:"vote_type" yaml:"vote_type"`
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

	// Msgs
	cdc.RegisterConcrete(MsgSubmitProposal{}, "kava/MsgSubmitProposal", nil)
	cdc.RegisterConcrete(MsgVote{}, "kava/MsgVote", nil)
}
