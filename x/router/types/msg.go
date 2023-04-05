package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

const (
	// TypeMsgMintDeposit defines the type for MsgMintDeposit
	TypeMsgMintDeposit = "mint_deposit"
	// TypeMsgDelegateMintDeposit defines the type for MsgDelegateMintDeposit
	TypeMsgDelegateMintDeposit = "delegate_mint_deposit"
	// TypeMsgWithdrawBurn defines the type for MsgWithdrawBurn
	TypeMsgWithdrawBurn = "withdraw_burn"
	// TypeMsgWithdrawBurnUndelegate defines the type for MsgWithdrawBurnUndelegate
	TypeMsgWithdrawBurnUndelegate = "withdraw_burn_undelegate"
)

var (
	_ sdk.Msg            = &MsgMintDeposit{}
	_ legacytx.LegacyMsg = &MsgMintDeposit{}
	_ sdk.Msg            = &MsgDelegateMintDeposit{}
	_ legacytx.LegacyMsg = &MsgDelegateMintDeposit{}
	_ sdk.Msg            = &MsgWithdrawBurn{}
	_ legacytx.LegacyMsg = &MsgWithdrawBurn{}
	_ sdk.Msg            = &MsgWithdrawBurnUndelegate{}
	_ legacytx.LegacyMsg = &MsgWithdrawBurnUndelegate{}
)

// NewMsgMintDeposit returns a new MsgMintDeposit.
func NewMsgMintDeposit(depositor sdk.AccAddress, validator sdk.ValAddress, amount sdk.Coin) *MsgMintDeposit {
	return &MsgMintDeposit{
		Depositor: depositor.String(),
		Validator: validator.String(),
		Amount:    amount,
	}
}

// Route return the message type used for routing the message.
func (msg MsgMintDeposit) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgMintDeposit) Type() string { return TypeMsgMintDeposit }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgMintDeposit) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Depositor); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid depositor address: %s", err)
	}

	if _, err := sdk.ValAddressFromBech32(msg.Validator); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid validator address: %s", err)
	}

	if msg.Amount.IsNil() || !msg.Amount.IsValid() || msg.Amount.IsZero() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "'%s'", msg.Amount)
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgMintDeposit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgMintDeposit) GetSigners() []sdk.AccAddress {
	depositor, _ := sdk.AccAddressFromBech32(msg.Depositor)
	return []sdk.AccAddress{depositor}
}

// NewMsgDelegateMintDeposit returns a new MsgDelegateMintDeposit.
func NewMsgDelegateMintDeposit(depositor sdk.AccAddress, validator sdk.ValAddress, amount sdk.Coin) *MsgDelegateMintDeposit {
	return &MsgDelegateMintDeposit{
		Depositor: depositor.String(),
		Validator: validator.String(),
		Amount:    amount,
	}
}

// Route return the message type used for routing the message.
func (msg MsgDelegateMintDeposit) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgDelegateMintDeposit) Type() string { return TypeMsgDelegateMintDeposit }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgDelegateMintDeposit) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Depositor); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid depositor address: %s", err)
	}

	if _, err := sdk.ValAddressFromBech32(msg.Validator); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid validator address: %s", err)
	}

	if msg.Amount.IsNil() || !msg.Amount.IsValid() || msg.Amount.IsZero() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "'%s'", msg.Amount)
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgDelegateMintDeposit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgDelegateMintDeposit) GetSigners() []sdk.AccAddress {
	depositor, _ := sdk.AccAddressFromBech32(msg.Depositor)
	return []sdk.AccAddress{depositor}
}

// NewMsgWithdrawBurn returns a new MsgWithdrawBurn.
func NewMsgWithdrawBurn(from sdk.AccAddress, validator sdk.ValAddress, amount sdk.Coin) *MsgWithdrawBurn {
	return &MsgWithdrawBurn{
		From:      from.String(),
		Validator: validator.String(),
		Amount:    amount,
	}
}

// Route return the message type used for routing the message.
func (msg MsgWithdrawBurn) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgWithdrawBurn) Type() string { return TypeMsgWithdrawBurn }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgWithdrawBurn) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.From); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address: %s", err)
	}

	if _, err := sdk.ValAddressFromBech32(msg.Validator); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid validator address: %s", err)
	}

	if msg.Amount.IsNil() || !msg.Amount.IsValid() || msg.Amount.IsZero() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "'%s'", msg.Amount)
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgWithdrawBurn) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgWithdrawBurn) GetSigners() []sdk.AccAddress {
	from, _ := sdk.AccAddressFromBech32(msg.From)
	return []sdk.AccAddress{from}
}

// NewMsgWithdrawBurnUndelegate returns a new MsgWithdrawBurnUndelegate.
func NewMsgWithdrawBurnUndelegate(from sdk.AccAddress, validator sdk.ValAddress, amount sdk.Coin) *MsgWithdrawBurnUndelegate {
	return &MsgWithdrawBurnUndelegate{
		From:      from.String(),
		Validator: validator.String(),
		Amount:    amount,
	}
}

// Route return the message type used for routing the message.
func (msg MsgWithdrawBurnUndelegate) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgWithdrawBurnUndelegate) Type() string { return TypeMsgWithdrawBurnUndelegate }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgWithdrawBurnUndelegate) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.From); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address: %s", err)
	}

	if _, err := sdk.ValAddressFromBech32(msg.Validator); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid validator address: %s", err)
	}

	if msg.Amount.IsNil() || !msg.Amount.IsValid() || msg.Amount.IsZero() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "'%s'", msg.Amount)
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgWithdrawBurnUndelegate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgWithdrawBurnUndelegate) GetSigners() []sdk.AccAddress {
	from, _ := sdk.AccAddressFromBech32(msg.From)
	return []sdk.AccAddress{from}
}
