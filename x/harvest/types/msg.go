package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MultiplierName name for valid multiplier
type MultiplierName string

// ClaimType type for valid claim type strings
type ClaimType string

// Valid reward multipliers and reward types
const (
	Small  MultiplierName = "small"
	Medium MultiplierName = "medium"
	Large  MultiplierName = "large"

	LP    ClaimType = "lp"
	Stake ClaimType = "stake"
)

// Queryable claim types
var (
	ClaimTypesClaimQuery = []ClaimType{LP, Stake}
)

// IsValid checks if the input is one of the expected strings
func (mn MultiplierName) IsValid() error {
	switch mn {
	case Small, Medium, Large:
		return nil
	}
	return fmt.Errorf("invalid multiplier name: %s", mn)
}

// IsValid checks if the input is one of the expected strings
func (dt ClaimType) IsValid() error {
	switch dt {
	case LP, Stake:
		return nil
	}
	return fmt.Errorf("invalid claim type: %s", dt)
}

// ensure Msg interface compliance at compile time
var (
	_ sdk.Msg = &MsgClaimReward{}
	_ sdk.Msg = &MsgDeposit{}
	_ sdk.Msg = &MsgWithdraw{}
	_ sdk.Msg = &MsgBorrow{}
)

// MsgDeposit deposit collateral to the harvest module.
type MsgDeposit struct {
	Depositor sdk.AccAddress `json:"depositor" yaml:"depositor"`
	Amount    sdk.Coin       `json:"amount" yaml:"amount"`
}

// NewMsgDeposit returns a new MsgDeposit
func NewMsgDeposit(depositor sdk.AccAddress, amount sdk.Coin) MsgDeposit {
	return MsgDeposit{
		Depositor: depositor,
		Amount:    amount,
	}
}

// Route return the message type used for routing the message.
func (msg MsgDeposit) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgDeposit) Type() string { return "harvest_deposit" }

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

// MsgWithdraw withdraw from the harvest module.
type MsgWithdraw struct {
	Depositor sdk.AccAddress `json:"depositor" yaml:"depositor"`
	Amount    sdk.Coin       `json:"amount" yaml:"amount"`
}

// NewMsgWithdraw returns a new MsgWithdraw
func NewMsgWithdraw(depositor sdk.AccAddress, amount sdk.Coin) MsgWithdraw {
	return MsgWithdraw{
		Depositor: depositor,
		Amount:    amount,
	}
}

// Route return the message type used for routing the message.
func (msg MsgWithdraw) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgWithdraw) Type() string { return "harvest_withdraw" }

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

// MsgClaimReward message type used to claim rewards
type MsgClaimReward struct {
	Sender         sdk.AccAddress `json:"sender" yaml:"sender"`
	Receiver       sdk.AccAddress `json:"receiver" yaml:"receiver"`
	DepositDenom   string         `json:"deposit_denom" yaml:"deposit_denom"`
	MultiplierName string         `json:"multiplier_name" yaml:"multiplier_name"`
	ClaimType      string         `json:"claim_type" yaml:"claim_type"`
}

// NewMsgClaimReward returns a new MsgClaimReward.
func NewMsgClaimReward(sender, receiver sdk.AccAddress, depositDenom, claimType, multiplier string) MsgClaimReward {
	return MsgClaimReward{
		Sender:         sender,
		Receiver:       receiver,
		DepositDenom:   depositDenom,
		MultiplierName: multiplier,
		ClaimType:      claimType,
	}
}

// Route return the message type used for routing the message.
func (msg MsgClaimReward) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgClaimReward) Type() string { return "claim_harvest_reward" }

// ValidateBasic does a simple validation check that doesn't require access to state.
func (msg MsgClaimReward) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty")
	}
	if msg.Receiver.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "receiver address cannot be empty")
	}
	if err := sdk.ValidateDenom(msg.DepositDenom); err != nil {
		return fmt.Errorf("collateral type cannot be blank")
	}
	if err := ClaimType(strings.ToLower(msg.ClaimType)).IsValid(); err != nil {
		return err
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

// ---------------------------------------

// MsgBorrow borrows funds from the harvest module.
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
func (msg MsgBorrow) Type() string { return "harvest_borrow" } // TODO: or just 'borrow'

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
