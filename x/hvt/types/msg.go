package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// RewardMultiplier type for valid reward multiplier strings
type RewardMultiplier string

// DepositType type for valid deposit type strings
type DepositType string

// Valid reward multipliers and reward types
const (
	Small  RewardMultiplier = "small"
	Medium RewardMultiplier = "medium"
	Large  RewardMultiplier = "large"

	LP    DepositType = "lp"
	Stake DepositType = "stake"
)

// IsValid checks if the input is one of the expected strings
func (rm RewardMultiplier) IsValid() error {
	switch rm {
	case Small, Medium, Large:
		return nil
	}
	return fmt.Errorf("invalid reward multiplier: %s", rm)
}

// IsValid checks if the input is one of the expected strings
func (dt DepositType) IsValid() error {
	switch dt {
	case LP, Stake:
		return nil
	}
	return fmt.Errorf("invalid deposit type: %s", dt)
}

// ensure Msg interface compliance at compile time
var (
	_ sdk.Msg = &MsgClaimReward{}
	_ sdk.Msg = &MsgDeposit{}
	_ sdk.Msg = &MsgWithdraw{}
)

// MsgDeposit deposit collateral to an existing cdp.
type MsgDeposit struct {
	Depositor   sdk.AccAddress `json:"depositor" yaml:"depositor"`
	Amount      sdk.Coin       `json:"amount" yaml:"amount"`
	DepositType string         `json:"deposit_type" yaml:"deposit_type"`
}

// NewMsgDeposit returns a new MsgDeposit
func NewMsgDeposit(depositor sdk.AccAddress, amount sdk.Coin, depositType string) MsgDeposit {
	return MsgDeposit{
		Depositor:   depositor,
		Amount:      amount,
		DepositType: depositType,
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
	return DepositType(strings.ToLower(msg.DepositType)).IsValid()
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
	Deposit Type: %s
`, msg.Depositor, msg.Amount, msg.DepositType)
}

// MsgWithdraw deposit collateral to an existing cdp.
type MsgWithdraw struct {
	Depositor   sdk.AccAddress `json:"depositor" yaml:"depositor"`
	Amount      sdk.Coin       `json:"amount" yaml:"amount"`
	DepositType string         `json:"deposit_type" yaml:"deposit_type"`
}

// NewMsgWithdraw returns a new MsgWithdraw
func NewMsgWithdraw(depositor sdk.AccAddress, amount sdk.Coin, depositType string) MsgDeposit {
	return MsgDeposit{
		Depositor:   depositor,
		Amount:      amount,
		DepositType: depositType,
	}
}

// Route return the message type used for routing the message.
func (msg MsgWithdraw) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgWithdraw) Type() string { return "harvest_withdrawal" }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgWithdraw) ValidateBasic() error {
	if msg.Depositor.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty")
	}
	if !msg.Amount.IsValid() || msg.Amount.IsZero() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "deposit amount %s", msg.Amount)
	}
	return DepositType(strings.ToLower(msg.DepositType)).IsValid()
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
	Deposit Type: %s
`, msg.Depositor, msg.Amount, msg.DepositType)
}

// MsgClaimReward message type used to claim rewards
type MsgClaimReward struct {
	Sender           sdk.AccAddress `json:"sender" yaml:"sender"`
	DepositDenom     string         `json:"deposit_denom" yaml:"deposit_denom"`
	RewardMultiplier string         `json:"reward_multiplier" yaml:"reward_multiplier"`
	DepositType      string         `json:"deposit_type" yaml:"deposit_type"`
}

// NewMsgClaimReward returns a new MsgClaimReward.
func NewMsgClaimReward(sender sdk.AccAddress, depositDenom, depositType, multiplier string) MsgClaimReward {
	return MsgClaimReward{
		Sender:           sender,
		DepositDenom:     depositDenom,
		RewardMultiplier: multiplier,
		DepositType:      depositType,
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
	if err := sdk.ValidateDenom(msg.DepositDenom); err != nil {
		return fmt.Errorf("collateral type cannot be blank")
	}
	if err := DepositType(strings.ToLower(msg.DepositType)).IsValid(); err != nil {
		return err
	}
	return RewardMultiplier(strings.ToLower(msg.RewardMultiplier)).IsValid()
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
