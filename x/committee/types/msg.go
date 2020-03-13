package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	TypeMsgSubmitProposal = "submit_proposal" // TODO these are the same as the gov module, will there be collisions?
	TypeMsgVote           = "vote"
)

var _, _ sdk.Msg = MsgSubmitProposal{}, MsgVote{}

// MsgSubmitProposal is used by committee members to create a new proposal that they can vote on.
type MsgSubmitProposal struct {
	PubProposal PubProposal    `json:"pub_proposal" yaml:"pub_proposal"`
	Proposer    sdk.AccAddress `json:"proposer" yaml:"proposer"`
	CommitteeID uint64
}

// NewMsgSubmitProposal creates a new MsgSubmitProposal instance
func NewMsgSubmitProposal(pubProposal PubProposal, proposer sdk.AccAddress, committeeID uint64) MsgSubmitProposal {
	return MsgSubmitProposal{
		PubProposal: pubProposal,
		Proposer:    proposer,
		CommitteeID: committeeID,
	}
}

// Route return the message type used for routing the message.
func (msg MsgSubmitProposal) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within events.
func (msg MsgSubmitProposal) Type() string { return TypeMsgSubmitProposal }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgSubmitProposal) ValidateBasic() sdk.Error {
	if msg.PubProposal == nil {
		return sdk.ErrInternal("no proposal")
	}
	if msg.Proposer.Empty() {
		return sdk.ErrInvalidAddress(msg.Proposer.String())
	}
	// TODO
	// if !IsValidProposalType(msg.Content.ProposalType()) {
	// 	return ErrInvalidProposalType(DefaultCodespace, msg.Content.ProposalType())
	// }

	return msg.PubProposal.ValidateBasic()
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgSubmitProposal) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgSubmitProposal) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Proposer}
}

// MsgVote is submitted by committee members to vote on proposals.
type MsgVote struct {
	ProposalID uint64         `json:"proposal_id" yaml:"proposal_id"`
	Voter      sdk.AccAddress `json:"voter" yaml:"voter"`
}

// NewMsgVote creates a message to cast a vote on an active proposal
func NewMsgVote(voter sdk.AccAddress, proposalID uint64) MsgVote {
	return MsgVote{proposalID, voter}
}

// Route return the message type used for routing the message.
func (msg MsgVote) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within events.
func (msg MsgVote) Type() string { return TypeMsgVote }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgVote) ValidateBasic() sdk.Error {
	if msg.Voter.Empty() {
		return sdk.ErrInvalidAddress(msg.Voter.String())
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgVote) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgVote) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Voter}
}
