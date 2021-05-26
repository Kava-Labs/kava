package types

import (
	"encoding/json"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	yaml "gopkg.in/yaml.v2"
)

const (
	TypeMsgSubmitProposal = "commmittee_submit_proposal" // 'committee' prefix appended to avoid potential conflicts with gov msg types
	TypeMsgVote           = "committee_vote"
)

var _, _ sdk.Msg = MsgSubmitProposal{}, MsgVote{}

// MsgSubmitProposal is used by committee members to create a new proposal that they can vote on.
type MsgSubmitProposal struct {
	PubProposal PubProposal    `json:"pub_proposal" yaml:"pub_proposal"`
	Proposer    sdk.AccAddress `json:"proposer" yaml:"proposer"`
	CommitteeID uint64         `json:"committee_id" yaml:"committee_id"`
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
func (msg MsgSubmitProposal) ValidateBasic() error {
	if msg.PubProposal == nil {
		return sdkerrors.Wrap(ErrInvalidPubProposal, "pub proposal cannot be nil")
	}
	if msg.Proposer.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "proposer address cannot be empty")
	}

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

type VoteType uint64

const (
	NullVoteType VoteType = iota // 0
	Yes          VoteType = iota // 1
	No           VoteType = iota // 2
	Abstain      VoteType = iota // 3
)

// VoteTypeFromString returns a VoteType from a string. It returns an error
// if the string is invalid.
func VoteTypeFromString(str string) (VoteType, error) {
	switch strings.ToLower(str) {
	case "yes", "y":
		return Yes, nil

	case "abstain", "a":
		return Abstain, nil

	case "no", "n":
		return No, nil

	default:
		return VoteType(0xff), fmt.Errorf("'%s' is not a valid vote type", str)
	}
}

// Marshal needed for protobuf compatibility.
func (vt VoteType) Marshal() ([]byte, error) {
	return []byte{byte(vt)}, nil
}

// Unmarshal needed for protobuf compatibility.
func (vt *VoteType) Unmarshal(data []byte) error {
	*vt = VoteType(data[0])
	return nil
}

// Marshals to JSON using string.
func (vt VoteType) MarshalJSON() ([]byte, error) {
	return json.Marshal(vt.String())
}

// UnmarshalJSON decodes from JSON assuming Bech32 encoding.
func (vt *VoteType) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	bz2, err := VoteTypeFromString(s)
	if err != nil {
		return err
	}

	*vt = bz2
	return nil
}

// Marshals to YAML using string.
func (vt VoteType) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(vt.String())
}

// UnmarshalJSON decodes from YAML assuming Bech32 encoding.
func (vt *VoteType) UnmarshalYAML(data []byte) error {
	var s string
	err := yaml.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	bz2, err := VoteTypeFromString(s)
	if err != nil {
		return err
	}

	*vt = bz2
	return nil
}

// String implements the Stringer interface.
func (vt VoteType) String() string {
	switch vt {
	case Yes:
		return "Yes"
	case Abstain:
		return "Abstain"
	case No:
		return "No"
	default:
		return ""
	}
}

func (vt VoteType) Validate() error {
	if vt <= 0 || vt > 3 {
		return fmt.Errorf("invalid vote type: %d", vt)
	}
	return nil
}

// MsgVote is submitted by committee members to vote on proposals.
type MsgVote struct {
	ProposalID uint64         `json:"proposal_id" yaml:"proposal_id"`
	Voter      sdk.AccAddress `json:"voter" yaml:"voter"`
	VoteType   VoteType       `json:"vote_type" yaml:"vote_type"`
}

// NewMsgVote creates a message to cast a vote on an active proposal
func NewMsgVote(voter sdk.AccAddress, proposalID uint64, voteType VoteType) MsgVote {
	return MsgVote{proposalID, voter, voteType}
}

// Route return the message type used for routing the message.
func (msg MsgVote) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within events.
func (msg MsgVote) Type() string { return TypeMsgVote }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgVote) ValidateBasic() error {
	if msg.Voter.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "voter address cannot be empty")
	}

	return msg.VoteType.Validate()
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
