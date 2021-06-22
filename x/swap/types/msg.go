package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg         = &MsgDeposit{}
	_ MsgWithDeadline = &MsgDeposit{}
	_ sdk.Msg         = &MsgWithdraw{}
	_ MsgWithDeadline = &MsgWithdraw{}
)

// MsgWithDeadline allows messages to define a deadline of when they are considered invalid
type MsgWithDeadline interface {
	GetDeadline() time.Time
	DeadlineExceeded(blockTime time.Time) bool
}

// MsgDeposit deposits liquidity into a pool
type MsgDeposit struct {
	Depositor sdk.AccAddress `json:"depositor" yaml:"depositor"`
	TokenA    sdk.Coin       `json:"token_a" yaml:"token_a"`
	TokenB    sdk.Coin       `json:"token_b" yaml:"token_b"`
	Deadline  int64          `json:"deadline" yaml:"deadline"`
}

// NewMsgDeposit returns a new MsgDeposit
func NewMsgDeposit(depositor sdk.AccAddress, tokenA sdk.Coin, tokenB sdk.Coin, deadline int64) MsgDeposit {
	return MsgDeposit{
		Depositor: depositor,
		TokenA:    tokenA,
		TokenB:    tokenB,
		Deadline:  deadline,
	}
}

// Route return the message type used for routing the message.
func (msg MsgDeposit) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgDeposit) Type() string { return "swap_deposit" }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgDeposit) ValidateBasic() error {
	if msg.Depositor.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "depositor address cannot be empty")
	}

	if !msg.TokenA.IsValid() || msg.TokenA.IsZero() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "token a deposit amount %s", msg.TokenA)
	}

	if !msg.TokenB.IsValid() || msg.TokenB.IsZero() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "token b deposit amount %s", msg.TokenB)
	}

	if msg.TokenA.Denom == msg.TokenB.Denom {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "denominations can not be equal")
	}

	if msg.Deadline <= 0 {
		return sdkerrors.Wrapf(ErrInvalidDeadline, "deadline %d", msg.Deadline)
	}

	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgDeposit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgDeposit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Depositor}
}

// GetDeadline returns the time at which the msg is considered invalid
func (msg MsgDeposit) GetDeadline() time.Time {
	return time.Unix(msg.Deadline, 0)
}

// DeadlineExceeded returns if the msg has exceeded it's deadline
func (msg MsgDeposit) DeadlineExceeded(blockTime time.Time) bool {
	return blockTime.Unix() >= msg.Deadline
}

// MsgWithdraw deposits liquidity into a pool
type MsgWithdraw struct {
	From     sdk.AccAddress `json:"from" yaml:"from"`
	Pool     string         `json:"pool" yaml:"pool"`
	Shares   sdk.Int        `json:"shares" yaml:"shares"`
	Slippage sdk.Dec        `json:"slippage" yaml:"slippage"`
	Deadline int64          `json:"deadline" yaml:"deadline"`
}

// NewMsgWithdraw returns a new MsgWithdraw
func NewMsgWithdraw(from sdk.AccAddress, pool string, shares sdk.Int, slippage sdk.Dec, deadline int64) MsgWithdraw {
	return MsgWithdraw{
		From:     from,
		Pool:     pool,
		Shares:   shares,
		Slippage: slippage,
		Deadline: deadline,
	}
}

// Route return the message type used for routing the message.
func (msg MsgWithdraw) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgWithdraw) Type() string { return "swap_withdraw" }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgWithdraw) ValidateBasic() error {
	if msg.From.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "from address cannot be empty")
	}

	if len(msg.Pool) == 0 {
		return sdkerrors.Wrap(ErrInvalidPool, "pool ID cannot be empty")
	}

	if msg.Shares.IsNil() || msg.Shares.IsZero() || msg.Shares.IsNegative() {
		return sdkerrors.Wrapf(ErrInvalidShares, fmt.Sprintf("%s", msg.Shares))
	}

	if msg.Slippage.IsNil() || msg.Slippage.IsNegative() || msg.Slippage.GT(sdk.OneDec()) {
		return sdkerrors.Wrapf(ErrInvalidSlippage, fmt.Sprintf("%s", msg.Slippage))
	}

	if msg.Deadline <= 0 {
		return sdkerrors.Wrapf(ErrInvalidDeadline, "deadline %d", msg.Deadline)
	}

	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgWithdraw) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgWithdraw) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.From}
}

// GetDeadline returns the time at which the msg is considered invalid
func (msg MsgWithdraw) GetDeadline() time.Time {
	return time.Unix(msg.Deadline, 0)
}

// DeadlineExceeded returns if the msg has exceeded it's deadline
func (msg MsgWithdraw) DeadlineExceeded(blockTime time.Time) bool {
	return blockTime.Unix() >= msg.Deadline
}
