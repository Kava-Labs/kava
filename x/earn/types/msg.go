package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

var (
	_ sdk.Msg            = &MsgDeposit{}
	_ sdk.Msg            = &MsgWithdraw{}
	_ legacytx.LegacyMsg = &MsgDeposit{}
	_ legacytx.LegacyMsg = &MsgWithdraw{}
)

// legacy message types
const (
	TypeMsgDeposit  = "earn_msg_deposit"
	TypeMsgWithdraw = "earn_msg_withdraw"
)

// NewMsgDeposit returns a new MsgDeposit.
func NewMsgDeposit(depositor string, amount sdk.Coin, strategy StrategyType) *MsgDeposit {
	return &MsgDeposit{
		Depositor: depositor,
		Amount:    amount,
		Strategy:  strategy,
	}
}

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgDeposit) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Depositor); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
	}

	if err := msg.Amount.Validate(); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, err.Error())
	}

	if err := msg.Strategy.Validate(); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
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
	depositor, err := sdk.AccAddressFromBech32(msg.Depositor)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{depositor}
}

// Route implements the LegacyMsg.Route method.
func (msg MsgDeposit) Route() string {
	return RouterKey
}

// Type implements the LegacyMsg.Type method.
func (msg MsgDeposit) Type() string {
	return TypeMsgDeposit
}

// NewMsgWithdraw returns a new MsgWithdraw.
func NewMsgWithdraw(from string, amount sdk.Coin, strategy StrategyType) *MsgWithdraw {
	return &MsgWithdraw{
		From:     from,
		Amount:   amount,
		Strategy: strategy,
	}
}

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgWithdraw) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.From); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
	}

	if err := msg.Amount.Validate(); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, err.Error())
	}

	if err := msg.Strategy.Validate(); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
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
	depositor, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{depositor}
}

// Route implements the LegacyMsg.Route method.
func (msg MsgWithdraw) Route() string {
	return RouterKey
}

// Type implements the LegacyMsg.Type method.
func (msg MsgWithdraw) Type() string {
	return TypeMsgWithdraw
}
