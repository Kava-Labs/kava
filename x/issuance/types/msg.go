package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	TypeMsgIssueTokens    = "issue_tokens"
	TypeMsgRedeemTokens   = "redeem_tokens"
	TypeMsgBlockAddress   = "block_address"
	TypeMsgUnBlockAddress = "unblock_address"
	TypeMsgSetPauseStatus = "change_pause_status"
)

// ensure Msg interface compliance at compile time
var (
	_ sdk.Msg = &MsgIssueTokens{}
	_ sdk.Msg = &MsgRedeemTokens{}
	_ sdk.Msg = &MsgBlockAddress{}
	_ sdk.Msg = &MsgUnblockAddress{}
	_ sdk.Msg = &MsgSetPauseStatus{}
)

// NewMsgIssueTokens returns a new MsgIssueTokens
func NewMsgIssueTokens(sender string, tokens sdk.Coin, receiver string) *MsgIssueTokens {
	return &MsgIssueTokens{
		Sender:   sender,
		Tokens:   tokens,
		Receiver: receiver,
	}
}

// Route return the message type used for routing the message.
func (msg MsgIssueTokens) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgIssueTokens) Type() string { return TypeMsgIssueTokens }

// ValidateBasic does a simple validation check that doesn't require access to state.
func (msg MsgIssueTokens) ValidateBasic() error {
	if len(msg.Sender) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty")
	}
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender bech32 address")
	}
	if msg.Tokens.IsZero() || !msg.Tokens.IsValid() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "invalid tokens %s", msg.Tokens)
	}
	if len(msg.Receiver) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "receiver address cannot be empty")
	}
	_, err = sdk.AccAddressFromBech32(msg.Receiver)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid receiver bech32 address")
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg
func (msg MsgIssueTokens) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign
func (msg MsgIssueTokens) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// NewMsgRedeemTokens returns a new MsgRedeemTokens
func NewMsgRedeemTokens(sender string, tokens sdk.Coin) *MsgRedeemTokens {
	return &MsgRedeemTokens{
		Sender: sender,
		Tokens: tokens,
	}
}

// Route return the message type used for routing the message.
func (msg MsgRedeemTokens) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgRedeemTokens) Type() string { return TypeMsgRedeemTokens }

// ValidateBasic does a simple validation check that doesn't require access to state.
func (msg MsgRedeemTokens) ValidateBasic() error {
	if len(msg.Sender) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty")
	}
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender bech32 address")
	}
	if msg.Tokens.IsZero() || !msg.Tokens.IsValid() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "invalid tokens %s", msg.Tokens)
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg
func (msg MsgRedeemTokens) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign
func (msg MsgRedeemTokens) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// NewMsgBlockAddress returns a new MsgBlockAddress
func NewMsgBlockAddress(sender string, denom string, addr string) *MsgBlockAddress {
	return &MsgBlockAddress{
		Sender:         sender,
		Denom:          denom,
		BlockedAddress: addr,
	}
}

// Route return the message type used for routing the message.
func (msg MsgBlockAddress) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgBlockAddress) Type() string { return TypeMsgBlockAddress }

// ValidateBasic does a simple validation check that doesn't require access to state.
func (msg MsgBlockAddress) ValidateBasic() error {
	if len(msg.Sender) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty")
	}
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender bech32 address")
	}
	if len(msg.BlockedAddress) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "blocked address cannot be empty")
	}
	return sdk.ValidateDenom(msg.Denom)
}

// GetSignBytes gets the canonical byte representation of the Msg
func (msg MsgBlockAddress) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign
func (msg MsgBlockAddress) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// NewMsgUnblockAddress returns a new MsgUnblockAddress
func NewMsgUnblockAddress(sender string, denom string, addr string) *MsgUnblockAddress {
	return &MsgUnblockAddress{
		Sender:         sender,
		Denom:          denom,
		BlockedAddress: addr,
	}
}

// Route return the message type used for routing the message.
func (msg MsgUnblockAddress) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgUnblockAddress) Type() string { return TypeMsgUnBlockAddress }

// ValidateBasic does a simple validation check that doesn't require access to state.
func (msg MsgUnblockAddress) ValidateBasic() error {
	if len(msg.Sender) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty")
	}
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender bech32 address")
	}
	if len(msg.BlockedAddress) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "blocked address cannot be empty")
	}
	return sdk.ValidateDenom(msg.Denom)
}

// GetSignBytes gets the canonical byte representation of the Msg
func (msg MsgUnblockAddress) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign
func (msg MsgUnblockAddress) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// NewMsgSetPauseStatus returns a new MsgSetPauseStatus
func NewMsgSetPauseStatus(sender string, denom string, status bool) *MsgSetPauseStatus {
	return &MsgSetPauseStatus{
		Sender: sender,
		Denom:  denom,
		Status: status,
	}
}

// Route return the message type used for routing the message.
func (msg MsgSetPauseStatus) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgSetPauseStatus) Type() string { return TypeMsgSetPauseStatus }

// ValidateBasic does a simple validation check that doesn't require access to state.
func (msg MsgSetPauseStatus) ValidateBasic() error {
	if len(msg.Sender) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty")
	}
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender bech32 address")
	}
	return sdk.ValidateDenom(msg.Denom)
}

// GetSignBytes gets the canonical byte representation of the Msg
func (msg MsgSetPauseStatus) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign
func (msg MsgSetPauseStatus) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}
