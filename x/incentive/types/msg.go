package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// ensure Msg interface compliance at compile time
var _ sdk.Msg = &MsgClaimReward{}

// MsgClaimReward message type used to claim rewards
type MsgClaimReward struct {
	Sender         sdk.AccAddress `json:"sender" yaml:"sender"`
	CollateralType string         `json:"collateral_type" yaml:"collateral_type"`
	MultiplierName string         `json:"multiplier_name" yaml:"multiplier_name"`
}

// NewMsgClaimReward returns a new MsgClaimReward.
func NewMsgClaimReward(sender sdk.AccAddress, collateralType, multiplierName string) MsgClaimReward {
	return MsgClaimReward{
		Sender:         sender,
		CollateralType: collateralType,
		MultiplierName: multiplierName,
	}
}

// Route return the message type used for routing the message.
func (msg MsgClaimReward) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgClaimReward) Type() string { return "claim_reward" }

// ValidateBasic does a simple validation check that doesn't require access to state.
func (msg MsgClaimReward) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty")
	}
	if strings.TrimSpace(msg.CollateralType) == "" {
		return fmt.Errorf("collateral type cannot be blank")
	}
	return MultiplierName(strings.ToLower(msg.MultiplierName)).IsValid()
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
