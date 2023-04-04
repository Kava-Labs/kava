package types

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// TypeMsgDeposit represents the type string for MsgDeposit
	TypeMsgDeposit = "swap_deposit"
	// TypeMsgWithdraw represents the type string for MsgWithdraw
	TypeMsgWithdraw = "swap_withdraw"
	// TypeSwapExactForTokens represents the type string for MsgSwapExactForTokens
	TypeSwapExactForTokens = "swap_exact_for_tokens"
	// TypeSwapForExactTokens represents the type string for MsgSwapForExactTokens
	TypeSwapForExactTokens = "swap_for_exact_tokens"
)

var (
	_ sdk.Msg         = &MsgDeposit{}
	_ MsgWithDeadline = &MsgDeposit{}
	_ sdk.Msg         = &MsgWithdraw{}
	_ MsgWithDeadline = &MsgWithdraw{}
	_ sdk.Msg         = &MsgSwapExactForTokens{}
	_ MsgWithDeadline = &MsgSwapExactForTokens{}
	_ sdk.Msg         = &MsgSwapForExactTokens{}
	_ MsgWithDeadline = &MsgSwapForExactTokens{}
)

// MsgWithDeadline allows messages to define a deadline of when they are considered invalid
type MsgWithDeadline interface {
	GetDeadline() time.Time
	DeadlineExceeded(blockTime time.Time) bool
}

// NewMsgDeposit returns a new MsgDeposit
func NewMsgDeposit(depositor string, tokenA sdk.Coin, tokenB sdk.Coin, slippage sdk.Dec, deadline int64) *MsgDeposit {
	return &MsgDeposit{
		Depositor: depositor,
		TokenA:    tokenA,
		TokenB:    tokenB,
		Slippage:  slippage,
		Deadline:  deadline,
	}
}

// Route return the message type used for routing the message.
func (msg MsgDeposit) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgDeposit) Type() string { return TypeMsgDeposit }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgDeposit) ValidateBasic() error {
	if msg.Depositor == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "depositor address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Depositor); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid depositor address: %s", err)
	}

	if !msg.TokenA.IsValid() || msg.TokenA.IsZero() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "token a deposit amount %s", msg.TokenA)
	}

	if !msg.TokenB.IsValid() || msg.TokenB.IsZero() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "token b deposit amount %s", msg.TokenB)
	}

	if msg.TokenA.Denom == msg.TokenB.Denom {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, "denominations can not be equal")
	}

	if msg.Slippage.IsNil() {
		return errorsmod.Wrapf(ErrInvalidSlippage, "slippage must be set")
	}

	if msg.Slippage.IsNegative() {
		return errorsmod.Wrapf(ErrInvalidSlippage, "slippage can not be negative")
	}

	if msg.Deadline <= 0 {
		return errorsmod.Wrapf(ErrInvalidDeadline, "deadline %d", msg.Deadline)
	}

	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgDeposit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgDeposit) GetSigners() []sdk.AccAddress {
	depositor, _ := sdk.AccAddressFromBech32(msg.Depositor)
	return []sdk.AccAddress{depositor}
}

// GetDeadline returns the time at which the msg is considered invalid
func (msg MsgDeposit) GetDeadline() time.Time {
	return time.Unix(msg.Deadline, 0)
}

// DeadlineExceeded returns if the msg has exceeded it's deadline
func (msg MsgDeposit) DeadlineExceeded(blockTime time.Time) bool {
	return blockTime.Unix() >= msg.Deadline
}

// NewMsgWithdraw returns a new MsgWithdraw
func NewMsgWithdraw(from string, shares sdkmath.Int, minTokenA, minTokenB sdk.Coin, deadline int64) *MsgWithdraw {
	return &MsgWithdraw{
		From:      from,
		Shares:    shares,
		MinTokenA: minTokenA,
		MinTokenB: minTokenB,
		Deadline:  deadline,
	}
}

// Route return the message type used for routing the message.
func (msg MsgWithdraw) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgWithdraw) Type() string { return TypeMsgWithdraw }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgWithdraw) ValidateBasic() error {
	if msg.From == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "from address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.From); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address: %s", err)
	}

	if msg.Shares.IsNil() {
		return errorsmod.Wrapf(ErrInvalidShares, "shares must be set")
	}

	if msg.Shares.IsZero() || msg.Shares.IsNegative() {
		return errorsmod.Wrapf(ErrInvalidShares, msg.Shares.String())
	}

	if !msg.MinTokenA.IsValid() || msg.MinTokenA.IsZero() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "min token a amount %s", msg.MinTokenA)
	}

	if !msg.MinTokenB.IsValid() || msg.MinTokenB.IsZero() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "min token b amount %s", msg.MinTokenB)
	}

	if msg.MinTokenA.Denom == msg.MinTokenB.Denom {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, "denominations can not be equal")
	}

	if msg.Deadline <= 0 {
		return errorsmod.Wrapf(ErrInvalidDeadline, "deadline %d", msg.Deadline)
	}

	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgWithdraw) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgWithdraw) GetSigners() []sdk.AccAddress {
	from, _ := sdk.AccAddressFromBech32(msg.From)
	return []sdk.AccAddress{from}
}

// GetDeadline returns the time at which the msg is considered invalid
func (msg MsgWithdraw) GetDeadline() time.Time {
	return time.Unix(msg.Deadline, 0)
}

// DeadlineExceeded returns if the msg has exceeded it's deadline
func (msg MsgWithdraw) DeadlineExceeded(blockTime time.Time) bool {
	return blockTime.Unix() >= msg.Deadline
}

// NewMsgSwapExactForTokens returns a new MsgSwapExactForTokens
func NewMsgSwapExactForTokens(requester string, exactTokenA sdk.Coin, tokenB sdk.Coin, slippage sdk.Dec, deadline int64) *MsgSwapExactForTokens {
	return &MsgSwapExactForTokens{
		Requester:   requester,
		ExactTokenA: exactTokenA,
		TokenB:      tokenB,
		Slippage:    slippage,
		Deadline:    deadline,
	}
}

// Route return the message type used for routing the message.
func (msg MsgSwapExactForTokens) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgSwapExactForTokens) Type() string { return TypeSwapExactForTokens }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgSwapExactForTokens) ValidateBasic() error {
	if msg.Requester == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "requester address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Requester); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid requester address: %s", err)
	}

	if !msg.ExactTokenA.IsValid() || msg.ExactTokenA.IsZero() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "exact token a deposit amount %s", msg.ExactTokenA)
	}

	if !msg.TokenB.IsValid() || msg.TokenB.IsZero() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "token b deposit amount %s", msg.TokenB)
	}

	if msg.ExactTokenA.Denom == msg.TokenB.Denom {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, "denominations can not be equal")
	}

	if msg.Slippage.IsNil() {
		return errorsmod.Wrapf(ErrInvalidSlippage, "slippage must be set")
	}

	if msg.Slippage.IsNegative() {
		return errorsmod.Wrapf(ErrInvalidSlippage, "slippage can not be negative")
	}

	if msg.Deadline <= 0 {
		return errorsmod.Wrapf(ErrInvalidDeadline, "deadline %d", msg.Deadline)
	}

	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgSwapExactForTokens) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgSwapExactForTokens) GetSigners() []sdk.AccAddress {
	requester, _ := sdk.AccAddressFromBech32(msg.Requester)
	return []sdk.AccAddress{requester}
}

// GetDeadline returns the time at which the msg is considered invalid
func (msg MsgSwapExactForTokens) GetDeadline() time.Time {
	return time.Unix(msg.Deadline, 0)
}

// DeadlineExceeded returns if the msg has exceeded it's deadline
func (msg MsgSwapExactForTokens) DeadlineExceeded(blockTime time.Time) bool {
	return blockTime.Unix() >= msg.Deadline
}

// NewMsgSwapForExactTokens returns a new MsgSwapForExactTokens
func NewMsgSwapForExactTokens(requester string, tokenA sdk.Coin, exactTokenB sdk.Coin, slippage sdk.Dec, deadline int64) *MsgSwapForExactTokens {
	return &MsgSwapForExactTokens{
		Requester:   requester,
		TokenA:      tokenA,
		ExactTokenB: exactTokenB,
		Slippage:    slippage,
		Deadline:    deadline,
	}
}

// Route return the message type used for routing the message.
func (msg MsgSwapForExactTokens) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgSwapForExactTokens) Type() string { return TypeSwapForExactTokens }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgSwapForExactTokens) ValidateBasic() error {
	if msg.Requester == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "requester address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Requester); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid requester address: %s", err)
	}

	if !msg.TokenA.IsValid() || msg.TokenA.IsZero() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "token a deposit amount %s", msg.TokenA)
	}

	if !msg.ExactTokenB.IsValid() || msg.ExactTokenB.IsZero() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "exact token b deposit amount %s", msg.ExactTokenB)
	}

	if msg.TokenA.Denom == msg.ExactTokenB.Denom {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, "denominations can not be equal")
	}

	if msg.Slippage.IsNil() {
		return errorsmod.Wrapf(ErrInvalidSlippage, "slippage must be set")
	}

	if msg.Slippage.IsNegative() {
		return errorsmod.Wrapf(ErrInvalidSlippage, "slippage can not be negative")
	}

	if msg.Deadline <= 0 {
		return errorsmod.Wrapf(ErrInvalidDeadline, "deadline %d", msg.Deadline)
	}

	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgSwapForExactTokens) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgSwapForExactTokens) GetSigners() []sdk.AccAddress {
	requester, _ := sdk.AccAddressFromBech32(msg.Requester)
	return []sdk.AccAddress{requester}
}

// GetDeadline returns the time at which the msg is considered invalid
func (msg MsgSwapForExactTokens) GetDeadline() time.Time {
	return time.Unix(msg.Deadline, 0)
}

// DeadlineExceeded returns if the msg has exceeded it's deadline
func (msg MsgSwapForExactTokens) DeadlineExceeded(blockTime time.Time) bool {
	return blockTime.Unix() >= msg.Deadline
}
