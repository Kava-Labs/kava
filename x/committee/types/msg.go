package types

import (
	fmt "fmt"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto"
)

const (
	TypeMsgSubmitProposal = "commmittee_submit_proposal" // 'committee' prefix appended to avoid potential conflicts with gov msg types
	TypeMsgVote           = "committee_vote"
)

var _, _ sdk.Msg = &MsgSubmitProposal{}, &MsgVote{}

// NewMsgSubmitProposal creates a new MsgSubmitProposal instance
func NewMsgSubmitProposal(pubProposal PubProposal, proposer sdk.AccAddress, committeeId uint64) (MsgSubmitProposal, error) {
	msg, ok := pubProposal.(proto.Message)
	if !ok {
		return MsgSubmitProposal{}, fmt.Errorf("can't proto marshal %T", msg)
	}
	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return MsgSubmitProposal{}, err
	}
	return MsgSubmitProposal{
		Any:         any,
		Proposer:    proposer.String(),
		CommitteeId: committeeId,
	}, nil
}

func (msg MsgSubmitProposal) GetPubProposal() PubProposal {
	content, ok := msg.Any.GetCachedValue().(PubProposal)
	if !ok {
		return nil
	}
	return content
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m MsgSubmitProposal) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var content PubProposal
	return unpacker.UnpackAny(m.Any, &content)
}

// Route return the message type used for routing the message.
func (msg MsgSubmitProposal) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within events.
func (msg MsgSubmitProposal) Type() string { return TypeMsgSubmitProposal }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgSubmitProposal) ValidateBasic() error {
	if msg.GetPubProposal() == nil {
		return sdkerrors.Wrap(ErrInvalidPubProposal, "pub proposal cannot be nil")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Proposer); err != nil {
		return err
	}
	return msg.GetPubProposal().ValidateBasic()
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgSubmitProposal) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgSubmitProposal) GetSigners() []sdk.AccAddress {
	proposer, err := sdk.AccAddressFromBech32(msg.Proposer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{proposer}
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

// String implements the Stringer interface.
func (vt VoteType) String() string {
	// TODO: QUESTION: Can we just use the protobuf generated string implementation?
	// Is the frontend relying on these specific values?
	switch vt {
	case VOTE_TYPE_YES:
		return "Yes"
	case VOTE_TYPE_ABSTAIN:
		return "Abstain"
	case VOTE_TYPE_NO:
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

// Format implements the fmt.Formatter interface.
func (vo VoteType) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(vo.String()))
	default:
		s.Write([]byte(fmt.Sprintf("%v", byte(vo))))
	}
}

// NewMsgVote creates a message to cast a vote on an active proposal
func NewMsgVote(voter string, proposalID uint64, voteType VoteType) MsgVote {
	return MsgVote{proposalID, voter, voteType}
}

// Route return the message type used for routing the message.
func (msg MsgVote) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within events.
func (msg MsgVote) Type() string { return TypeMsgVote }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgVote) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Voter); err != nil {
		return err
	}
	return msg.VoteType.Validate()
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgVote) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgVote) GetSigners() []sdk.AccAddress {
	voter, err := sdk.AccAddressFromBech32(msg.Voter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{voter}
}
