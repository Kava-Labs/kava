package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgDeposit{}
)

// MsgDeposit deposits liquidity into a pool
type MsgDeposit struct {
	Depositor sdk.AccAddress `json:"depositor" yaml:"depositor"`
	TokenA    sdk.Coin       `json:"token_a" yaml:"token_a"`
	TokenB    sdk.Coin       `json:"token_b" yaml:"token_b"`
}

// NewMsgDeposit returns a new MsgDeposit
func NewMsgDeposit(depositor sdk.AccAddress, tokenA sdk.Coin, tokenB sdk.Coin) MsgDeposit {
	return MsgDeposit{
		Depositor: depositor,
		TokenA:    tokenA,
		TokenB:    tokenB,
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
