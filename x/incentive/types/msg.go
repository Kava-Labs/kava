package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ensure Msg interface compliance at compile time
var _ sdk.Msg = &MsgClaimReward{}

// MsgClaimReward message type used to claim rewards
type MsgClaimReward struct {
	Sender sdk.AccAddress `json:"sender" yaml:"sender"`
	Denom  string         `json:"denom" yaml:"denom"`
}

// NewMsgClaimReward returns a new MsgClaimReward.
func NewMsgClaimReward(sender sdk.AccAddress, denom string) MsgClaimReward {
	return MsgClaimReward{
		Sender: sender,
		Denom:  denom,
	}
}

// Route return the message type used for routing the message.
func (msg MsgClaimReward) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgClaimReward) Type() string { return "claim_reward" }

// ValidateBasic does a simple validation check that doesn't require access to state.
func (msg MsgClaimReward) ValidateBasic() sdk.Error {
	if msg.Sender.Empty() {
		return sdk.ErrInvalidAddress("invalid (empty) bidder address")
	}
	if msg.Denom == "" {
		return sdk.ErrInternal("invalid (empty) denom")
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgClaimReward) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgClaimReward) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
