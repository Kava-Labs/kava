package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// ensure Msg interface compliance at compile time
var (
	_ sdk.Msg = &MsgMintDerivative{}
)

// NewMsgMintDerivative returns a new MsgMintDerivative
func NewMsgMintDerivative(sender sdk.AccAddress, validator sdk.ValAddress, amount sdk.Coin) MsgMintDerivative {
	return MsgMintDerivative{
		Sender:    sender.String(),
		Validator: validator.String(),
		Amount:    amount,
	}
}

// Route return the message type used for routing the message.
func (msg MsgMintDerivative) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgMintDerivative) Type() string { return "mint_derivative" }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgMintDerivative) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
	}

	_, err = sdk.ValAddressFromBech32(msg.Validator)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
	}
	if msg.Amount.IsNil() || !msg.Amount.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "'%s'", msg.Amount)
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgMintDerivative) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgMintDerivative) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// NewMsgBurnDerivative returns a new MsgBurnDerivative
func NewMsgBurnDerivative(sender sdk.AccAddress, validator sdk.ValAddress, amount sdk.Coin) MsgBurnDerivative {
	return MsgBurnDerivative{
		Sender:    sender.String(),
		Validator: validator.String(),
		Amount:    amount,
	}
}

// Route return the message type used for routing the message.
func (msg MsgBurnDerivative) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgBurnDerivative) Type() string { return "burn_derivative" }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgBurnDerivative) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
	}

	_, err = sdk.ValAddressFromBech32(msg.Validator)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
	}
	if msg.Amount.IsNil() || !msg.Amount.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "'%s'", msg.Amount)
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgBurnDerivative) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgBurnDerivative) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}
