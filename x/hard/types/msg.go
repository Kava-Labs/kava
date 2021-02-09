package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// ensure Msg interface compliance at compile time
var (
	_ sdk.Msg = &MsgDeposit{}
	_ sdk.Msg = &MsgWithdraw{}
	_ sdk.Msg = &MsgBorrow{}
	_ sdk.Msg = &MsgRepay{}
	_ sdk.Msg = &MsgLiquidate{}
)

// MsgDeposit deposit collateral to the hard module.
type MsgDeposit struct {
	Depositor sdk.AccAddress `json:"depositor" yaml:"depositor"`
	Amount    sdk.Coins      `json:"amount" yaml:"amount"`
}

// NewMsgDeposit returns a new MsgDeposit
func NewMsgDeposit(depositor sdk.AccAddress, amount sdk.Coins) MsgDeposit {
	return MsgDeposit{
		Depositor: depositor,
		Amount:    amount,
	}
}

// Route return the message type used for routing the message.
func (msg MsgDeposit) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgDeposit) Type() string { return "hard_deposit" }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgDeposit) ValidateBasic() error {
	if msg.Depositor.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty")
	}
	if !msg.Amount.IsValid() || msg.Amount.IsZero() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "deposit amount %s", msg.Amount)
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

// String implements the Stringer interface
func (msg MsgDeposit) String() string {
	return fmt.Sprintf(`Deposit Message:
	Depositor:         %s
	Amount: %s
`, msg.Depositor, msg.Amount)
}

// MsgWithdraw withdraw from the hard module.
type MsgWithdraw struct {
	Depositor sdk.AccAddress `json:"depositor" yaml:"depositor"`
	Amount    sdk.Coins      `json:"amount" yaml:"amount"`
}

// NewMsgWithdraw returns a new MsgWithdraw
func NewMsgWithdraw(depositor sdk.AccAddress, amount sdk.Coins) MsgWithdraw {
	return MsgWithdraw{
		Depositor: depositor,
		Amount:    amount,
	}
}

// Route return the message type used for routing the message.
func (msg MsgWithdraw) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgWithdraw) Type() string { return "hard_withdraw" }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgWithdraw) ValidateBasic() error {
	if msg.Depositor.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty")
	}
	if !msg.Amount.IsValid() || msg.Amount.IsZero() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "deposit amount %s", msg.Amount)
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
	return []sdk.AccAddress{msg.Depositor}
}

// String implements the Stringer interface
func (msg MsgWithdraw) String() string {
	return fmt.Sprintf(`Withdraw Message:
	Depositor:         %s
	Amount: %s
`, msg.Depositor, msg.Amount)
}

// MsgBorrow borrows funds from the hard module.
type MsgBorrow struct {
	Borrower sdk.AccAddress `json:"borrower" yaml:"borrower"`
	Amount   sdk.Coins      `json:"amount" yaml:"amount"`
}

// NewMsgBorrow returns a new MsgBorrow
func NewMsgBorrow(borrower sdk.AccAddress, amount sdk.Coins) MsgBorrow {
	return MsgBorrow{
		Borrower: borrower,
		Amount:   amount,
	}
}

// Route return the message type used for routing the message.
func (msg MsgBorrow) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgBorrow) Type() string { return "hard_borrow" }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgBorrow) ValidateBasic() error {
	if msg.Borrower.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty")
	}
	if !msg.Amount.IsValid() || msg.Amount.IsZero() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "borrow amount %s", msg.Amount)
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgBorrow) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgBorrow) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Borrower}
}

// String implements the Stringer interface
func (msg MsgBorrow) String() string {
	return fmt.Sprintf(`Borrow Message:
	Borrower:         %s
	Amount:   %s
`, msg.Borrower, msg.Amount)
}

// MsgRepay repays funds to the hard module.
type MsgRepay struct {
	Sender sdk.AccAddress `json:"sender" yaml:"sender"`
	Owner  sdk.AccAddress `json:"owner" yaml:"owner"`
	Amount sdk.Coins      `json:"amount" yaml:"amount"`
}

// NewMsgRepay returns a new MsgRepay
func NewMsgRepay(sender, owner sdk.AccAddress, amount sdk.Coins) MsgRepay {
	return MsgRepay{
		Sender: sender,
		Owner:  owner,
		Amount: amount,
	}
}

// Route return the message type used for routing the message.
func (msg MsgRepay) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgRepay) Type() string { return "hard_repay" }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgRepay) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty")
	}
	if msg.Owner.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "owner address cannot be empty")
	}
	if !msg.Amount.IsValid() || msg.Amount.IsZero() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "repay amount %s", msg.Amount)
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgRepay) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgRepay) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// String implements the Stringer interface
func (msg MsgRepay) String() string {
	return fmt.Sprintf(`Repay Message:
	Sender:         %s
	Owner:         %s
	Amount:   %s
`, msg.Sender, msg.Owner, msg.Amount)
}

// MsgLiquidate attempts to liquidate a borrower's borrow
type MsgLiquidate struct {
	Keeper   sdk.AccAddress `json:"keeper" yaml:"keeper"`
	Borrower sdk.AccAddress `json:"borrower" yaml:"borrower"`
}

// NewMsgLiquidate returns a new MsgLiquidate
func NewMsgLiquidate(keeper, borrower sdk.AccAddress) MsgLiquidate {
	return MsgLiquidate{
		Keeper:   keeper,
		Borrower: borrower,
	}
}

// Route return the message type used for routing the message.
func (msg MsgLiquidate) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgLiquidate) Type() string { return "liquidate" }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgLiquidate) ValidateBasic() error {
	if msg.Keeper.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "keeper address cannot be empty")
	}
	if msg.Borrower.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "borrower address cannot be empty")
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgLiquidate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgLiquidate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Keeper}
}

// String implements the Stringer interface
func (msg MsgLiquidate) String() string {
	return fmt.Sprintf(`Liquidate Message:
	Keeper:           %s
	Borrower:         %s
`, msg.Keeper, msg.Borrower)
}
