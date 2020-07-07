package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// ensure Msg interface compliance at compile time
var _ sdk.Msg = &MsgIssueTokens{}
var _ sdk.Msg = &MsgRedeemTokens{}
var _ sdk.Msg = &MsgBlockAddress{}
var _ sdk.Msg = &MsgUnblockAddress{}
var _ sdk.Msg = &MsgSetPauseStatus{}

// MsgIssueTokens message type used by the issuer to issue new tokens
type MsgIssueTokens struct {
	Sender   sdk.AccAddress `json:"sender" yaml:"sender"`
	Tokens   sdk.Coin       `json:"tokens" yaml:"tokens"`
	Receiver sdk.AccAddress `json:"receiver" yaml:"receiver"`
}

// NewMsgIssueTokens returns a new MsgIssueTokens
func NewMsgIssueTokens(sender sdk.AccAddress, tokens sdk.Coin, receiver sdk.AccAddress) MsgIssueTokens {
	return MsgIssueTokens{
		Sender:   sender,
		Tokens:   tokens,
		Receiver: receiver,
	}
}

// Route return the message type used for routing the message.
func (msg MsgIssueTokens) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgIssueTokens) Type() string { return "issue_tokens" }

// ValidateBasic does a simple validation check that doesn't require access to state.
func (msg MsgIssueTokens) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty")
	}
	if msg.Tokens.IsZero() || !msg.Tokens.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "invalid tokens %s", msg.Tokens)
	}
	if msg.Receiver.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "receiver address cannot be empty")
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg
func (msg MsgIssueTokens) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign
func (msg MsgIssueTokens) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// String implements fmt.Stringer
func (msg MsgIssueTokens) String() string {
	return fmt.Sprintf(`Issue Tokens:
	Sender %s
	Tokens %s
	Receiver %s
	`, msg.Sender, msg.Tokens, msg.Receiver,
	)
}

// MsgRedeemTokens message type used by the issuer to redeem (burn) tokens
type MsgRedeemTokens struct {
	Sender sdk.AccAddress `json:"sender" yaml:"sender"`
	Tokens sdk.Coin       `json:"tokens" yaml:"tokens"`
}

// NewMsgRedeemTokens returns a new MsgRedeemTokens
func NewMsgRedeemTokens(sender sdk.AccAddress, tokens sdk.Coin) MsgRedeemTokens {
	return MsgRedeemTokens{
		Sender: sender,
		Tokens: tokens,
	}
}

// Route return the message type used for routing the message.
func (msg MsgRedeemTokens) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgRedeemTokens) Type() string { return "redeem_tokens" }

// ValidateBasic does a simple validation check that doesn't require access to state.
func (msg MsgRedeemTokens) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty")
	}
	if msg.Tokens.IsZero() || !msg.Tokens.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "invalid tokens %s", msg.Tokens)
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg
func (msg MsgRedeemTokens) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign
func (msg MsgRedeemTokens) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// String implements fmt.Stringer
func (msg MsgRedeemTokens) String() string {
	return fmt.Sprintf(`Redeem Tokens:
	Sender %s
	Tokens %s
	`, msg.Sender, msg.Tokens,
	)
}

// MsgBlockAddress message type used by the issuer to block an address from holding or transferring tokens
type MsgBlockAddress struct {
	Sender  sdk.AccAddress `json:"sender" yaml:"sender"`
	Denom   string         `json:"denom" yaml:"denom"`
	Address sdk.AccAddress `json:"blocked_address" yaml:"blocked_address"`
}

// NewMsgBlockAddress returns a new MsgIssueTokens
func NewMsgBlockAddress(sender sdk.AccAddress, denom string, addr sdk.AccAddress) MsgBlockAddress {
	return MsgBlockAddress{
		Sender:  sender,
		Denom:   denom,
		Address: addr,
	}
}

// Route return the message type used for routing the message.
func (msg MsgBlockAddress) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgBlockAddress) Type() string { return "block_address" }

// ValidateBasic does a simple validation check that doesn't require access to state.
func (msg MsgBlockAddress) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty")
	}
	if msg.Address.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "blocked address cannot be empty")
	}
	return sdk.ValidateDenom(msg.Denom)
}

// GetSignBytes gets the canonical byte representation of the Msg
func (msg MsgBlockAddress) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign
func (msg MsgBlockAddress) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// String implements fmt.Stringer
func (msg MsgBlockAddress) String() string {
	return fmt.Sprintf(`Block Address:
	Sender %s
	Denom %s
	Address %s
	`, msg.Sender, msg.Denom, msg.Address,
	)
}

// MsgUnblockAddress message type used by the issuer to unblock an address from holding or transferring tokens
type MsgUnblockAddress struct {
	Sender  sdk.AccAddress `json:"sender" yaml:"sender"`
	Denom   string         `json:"denom" yaml:"denom"`
	Address sdk.AccAddress `json:"address" yaml:"address"`
}

// NewMsgUnblockAddress returns a new MsgIssueTokens
func NewMsgUnblockAddress(sender sdk.AccAddress, denom string, addr sdk.AccAddress) MsgUnblockAddress {
	return MsgUnblockAddress{
		Sender:  sender,
		Denom:   denom,
		Address: addr,
	}
}

// Route return the message type used for routing the message.
func (msg MsgUnblockAddress) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgUnblockAddress) Type() string { return "unblock_address" }

// ValidateBasic does a simple validation check that doesn't require access to state.
func (msg MsgUnblockAddress) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty")
	}
	if msg.Address.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "blocked address cannot be empty")
	}
	return sdk.ValidateDenom(msg.Denom)
}

// GetSignBytes gets the canonical byte representation of the Msg
func (msg MsgUnblockAddress) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign
func (msg MsgUnblockAddress) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// String implements fmt.Stringer
func (msg MsgUnblockAddress) String() string {
	return fmt.Sprintf(`Unblock Address:
	Sender %s
	Denom %s
	Address %s
	`, msg.Sender, msg.Denom, msg.Address,
	)
}

// MsgSetPauseStatus message type used by the issuer to issue new tokens
type MsgSetPauseStatus struct {
	Sender sdk.AccAddress `json:"sender" yaml:"sender"`
	Denom  string         `json:"denom" yaml:"denom"`
	Status bool           `json:"status" yaml:"status"`
}

// NewMsgSetPauseStatus returns a new MsgSetPauseStatus
func NewMsgSetPauseStatus(sender sdk.AccAddress, denom string, status bool) MsgSetPauseStatus {
	return MsgSetPauseStatus{
		Sender: sender,
		Denom:  denom,
		Status: status,
	}
}

// Route return the message type used for routing the message.
func (msg MsgSetPauseStatus) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgSetPauseStatus) Type() string { return "change_pause_status" }

// ValidateBasic does a simple validation check that doesn't require access to state.
func (msg MsgSetPauseStatus) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty")
	}
	return sdk.ValidateDenom(msg.Denom)
}

// GetSignBytes gets the canonical byte representation of the Msg
func (msg MsgSetPauseStatus) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign
func (msg MsgSetPauseStatus) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// String implements fmt.Stringer
func (msg MsgSetPauseStatus) String() string {
	return fmt.Sprintf(`Set Pause Status:
	Sender %s
	Denom %s
	Status %t
	`, msg.Sender, msg.Denom, msg.Status,
	)
}
