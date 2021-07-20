package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const MaxDenomsToClaim = 1000

// ensure Msg interface compliance at compile time
var _ sdk.Msg = &MsgClaimUSDXMintingReward{}
var _ sdk.Msg = &MsgClaimUSDXMintingRewardVVesting{}
var _ sdk.Msg = &MsgClaimHardReward{}
var _ sdk.Msg = &MsgClaimHardRewardVVesting{}
var _ sdk.Msg = &MsgClaimDelegatorReward{}
var _ sdk.Msg = &MsgClaimDelegatorRewardVVesting{}
var _ sdk.Msg = &MsgClaimSwapReward{}
var _ sdk.Msg = &MsgClaimSwapRewardVVesting{}

// MsgClaimUSDXMintingReward message type used to claim USDX minting rewards
type MsgClaimUSDXMintingReward struct {
	Sender         sdk.AccAddress `json:"sender" yaml:"sender"`
	MultiplierName string         `json:"multiplier_name" yaml:"multiplier_name"`
}

// NewMsgClaimUSDXMintingReward returns a new MsgClaimUSDXMintingReward.
func NewMsgClaimUSDXMintingReward(sender sdk.AccAddress, multiplierName string) MsgClaimUSDXMintingReward {
	return MsgClaimUSDXMintingReward{
		Sender:         sender,
		MultiplierName: multiplierName,
	}
}

// Route return the message type used for routing the message.
func (msg MsgClaimUSDXMintingReward) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgClaimUSDXMintingReward) Type() string { return "claim_usdx_minting_reward" }

// ValidateBasic does a simple validation check that doesn't require access to state.
func (msg MsgClaimUSDXMintingReward) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty")
	}
	if err := MultiplierName(msg.MultiplierName).IsValid(); err != nil {
		return err
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgClaimUSDXMintingReward) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgClaimUSDXMintingReward) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// MsgClaimUSDXMintingRewardVVesting message type used to claim USDX minting rewards for validator vesting accounts
type MsgClaimUSDXMintingRewardVVesting struct {
	Sender         sdk.AccAddress `json:"sender" yaml:"sender"`
	Receiver       sdk.AccAddress `json:"receiver" yaml:"receiver"`
	MultiplierName string         `json:"multiplier_name" yaml:"multiplier_name"`
}

// NewMsgClaimUSDXMintingRewardVVesting returns a new MsgClaimUSDXMintingReward.
func NewMsgClaimUSDXMintingRewardVVesting(sender, receiver sdk.AccAddress, multiplierName string) MsgClaimUSDXMintingRewardVVesting {
	return MsgClaimUSDXMintingRewardVVesting{
		Sender:         sender,
		Receiver:       receiver,
		MultiplierName: multiplierName,
	}
}

// Route return the message type used for routing the message.
func (msg MsgClaimUSDXMintingRewardVVesting) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgClaimUSDXMintingRewardVVesting) Type() string {
	return "claim_usdx_minting_reward_vvesting"
}

// ValidateBasic does a simple validation check that doesn't require access to state.
func (msg MsgClaimUSDXMintingRewardVVesting) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty")
	}
	if msg.Receiver.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "receiver address cannot be empty")
	}
	if err := MultiplierName(msg.MultiplierName).IsValid(); err != nil {
		return err
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgClaimUSDXMintingRewardVVesting) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgClaimUSDXMintingRewardVVesting) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// MsgClaimHardReward message type used to claim Hard liquidity provider rewards
type MsgClaimHardReward struct {
	Sender        sdk.AccAddress `json:"sender" yaml:"sender"`
	DenomsToClaim Selections     `json:"denoms_to_claim" yaml:"denoms_to_claim"`
}

// NewMsgClaimHardReward returns a new MsgClaimHardReward.
func NewMsgClaimHardReward(sender sdk.AccAddress, denomsToClaim ...Selection) MsgClaimHardReward {
	return MsgClaimHardReward{
		Sender:        sender,
		DenomsToClaim: denomsToClaim,
	}
}

// Route return the message type used for routing the message.
func (msg MsgClaimHardReward) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgClaimHardReward) Type() string {
	return "claim_hard_reward"
}

// ValidateBasic does a simple validation check that doesn't require access to state.
func (msg MsgClaimHardReward) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty")
	}
	if err := msg.DenomsToClaim.Validate(); err != nil {
		return err
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgClaimHardReward) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgClaimHardReward) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// MsgClaimHardRewardVVesting message type used to claim Hard liquidity provider rewards for validator vesting accounts
type MsgClaimHardRewardVVesting struct {
	Sender        sdk.AccAddress `json:"sender" yaml:"sender"`
	Receiver      sdk.AccAddress `json:"receiver" yaml:"receiver"`
	DenomsToClaim Selections     `json:"denoms_to_claim" yaml:"denoms_to_claim"`
}

// NewMsgClaimHardRewardVVesting returns a new MsgClaimHardRewardVVesting.
func NewMsgClaimHardRewardVVesting(sender, receiver sdk.AccAddress, denomsToClaim ...Selection) MsgClaimHardRewardVVesting {
	return MsgClaimHardRewardVVesting{
		Sender:        sender,
		Receiver:      receiver,
		DenomsToClaim: denomsToClaim,
	}
}

// Route return the message type used for routing the message.
func (msg MsgClaimHardRewardVVesting) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgClaimHardRewardVVesting) Type() string {
	return "claim_hard_reward_vvesting"
}

// ValidateBasic does a simple validation check that doesn't require access to state.
func (msg MsgClaimHardRewardVVesting) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty")
	}
	if msg.Receiver.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "receiver address cannot be empty")
	}
	if err := msg.DenomsToClaim.Validate(); err != nil {
		return err
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgClaimHardRewardVVesting) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgClaimHardRewardVVesting) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// MsgClaimDelegatorReward message type used to claim delegator rewards
type MsgClaimDelegatorReward struct {
	Sender         sdk.AccAddress `json:"sender" yaml:"sender"`
	MultiplierName string         `json:"multiplier_name" yaml:"multiplier_name"`
	DenomsToClaim  []string       `json:"denoms_to_claim" yaml:"denoms_to_claim"`
}

// NewMsgClaimDelegatorReward returns a new MsgClaimDelegatorReward.
func NewMsgClaimDelegatorReward(sender sdk.AccAddress, multiplierName string, denomsToClaim []string) MsgClaimDelegatorReward {
	return MsgClaimDelegatorReward{
		Sender:         sender,
		MultiplierName: multiplierName,
		DenomsToClaim:  denomsToClaim,
	}
}

// Route return the message type used for routing the message.
func (msg MsgClaimDelegatorReward) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgClaimDelegatorReward) Type() string {
	return "claim_delegator_reward"
}

// ValidateBasic does a simple validation check that doesn't require access to state.
func (msg MsgClaimDelegatorReward) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty")
	}
	if err := MultiplierName(msg.MultiplierName).IsValid(); err != nil {
		return err
	}
	for i, d := range msg.DenomsToClaim {
		if i >= MaxDenomsToClaim {
			return sdkerrors.Wrapf(ErrInvalidClaimDenoms, "cannot claim more than %d denoms", MaxDenomsToClaim)
		}
		if err := sdk.ValidateDenom(d); err != nil {
			return sdkerrors.Wrap(ErrInvalidClaimDenoms, err.Error())
		}
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgClaimDelegatorReward) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgClaimDelegatorReward) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// MsgClaimDelegatorRewardVVesting message type used to claim delegator rewards for validator vesting accounts
type MsgClaimDelegatorRewardVVesting struct {
	Sender         sdk.AccAddress `json:"sender" yaml:"sender"`
	Receiver       sdk.AccAddress `json:"receiver" yaml:"receiver"`
	MultiplierName string         `json:"multiplier_name" yaml:"multiplier_name"`
	DenomsToClaim  []string       `json:"denoms_to_claim" yaml:"denoms_to_claim"`
}

// MsgClaimDelegatorRewardVVesting returns a new MsgClaimDelegatorRewardVVesting.
func NewMsgClaimDelegatorRewardVVesting(sender, receiver sdk.AccAddress, multiplierName string, denomsToClaim []string) MsgClaimDelegatorRewardVVesting {
	return MsgClaimDelegatorRewardVVesting{
		Sender:         sender,
		Receiver:       receiver,
		MultiplierName: multiplierName,
		DenomsToClaim:  denomsToClaim,
	}
}

// Route return the message type used for routing the message.
func (msg MsgClaimDelegatorRewardVVesting) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgClaimDelegatorRewardVVesting) Type() string {
	return "claim_delegator_reward_vvesting"
}

// ValidateBasic does a simple validation check that doesn't require access to state.
func (msg MsgClaimDelegatorRewardVVesting) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty")
	}
	if msg.Receiver.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "receiver address cannot be empty")
	}
	if err := MultiplierName(msg.MultiplierName).IsValid(); err != nil {
		return err
	}
	for i, d := range msg.DenomsToClaim {
		if i >= MaxDenomsToClaim {
			return sdkerrors.Wrapf(ErrInvalidClaimDenoms, "cannot claim more than %d denoms", MaxDenomsToClaim)
		}
		if err := sdk.ValidateDenom(d); err != nil {
			return sdkerrors.Wrap(ErrInvalidClaimDenoms, err.Error())
		}
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgClaimDelegatorRewardVVesting) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgClaimDelegatorRewardVVesting) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// MsgClaimSwapReward message type used to claim delegator rewards
type MsgClaimSwapReward struct {
	Sender         sdk.AccAddress `json:"sender" yaml:"sender"`
	MultiplierName string         `json:"multiplier_name" yaml:"multiplier_name"`
	DenomsToClaim  []string       `json:"denoms_to_claim" yaml:"denoms_to_claim"`
}

// NewMsgClaimSwapReward returns a new MsgClaimSwapReward.
func NewMsgClaimSwapReward(sender sdk.AccAddress, multiplierName string, denomsToClaim []string) MsgClaimSwapReward {
	return MsgClaimSwapReward{
		Sender:         sender,
		MultiplierName: multiplierName,
		DenomsToClaim:  denomsToClaim,
	}
}

// Route return the message type used for routing the message.
func (msg MsgClaimSwapReward) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgClaimSwapReward) Type() string {
	return "claim_swap_reward"
}

// ValidateBasic does a simple validation check that doesn't require access to state.
func (msg MsgClaimSwapReward) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty")
	}
	if err := MultiplierName(msg.MultiplierName).IsValid(); err != nil {
		return err
	}
	for i, d := range msg.DenomsToClaim {
		if i >= MaxDenomsToClaim {
			return sdkerrors.Wrapf(ErrInvalidClaimDenoms, "cannot claim more than %d denoms", MaxDenomsToClaim)
		}
		if err := sdk.ValidateDenom(d); err != nil {
			return sdkerrors.Wrap(ErrInvalidClaimDenoms, err.Error())
		}
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgClaimSwapReward) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgClaimSwapReward) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// MsgClaimSwapRewardVVesting message type used to claim delegator rewards for validator vesting accounts
type MsgClaimSwapRewardVVesting struct {
	Sender         sdk.AccAddress `json:"sender" yaml:"sender"`
	Receiver       sdk.AccAddress `json:"receiver" yaml:"receiver"`
	MultiplierName string         `json:"multiplier_name" yaml:"multiplier_name"`
	DenomsToClaim  []string       `json:"denoms_to_claim" yaml:"denoms_to_claim"`
}

// MsgClaimSwapRewardVVesting returns a new MsgClaimSwapRewardVVesting.
func NewMsgClaimSwapRewardVVesting(sender, receiver sdk.AccAddress, multiplierName string, denomsToClaim []string) MsgClaimSwapRewardVVesting {
	return MsgClaimSwapRewardVVesting{
		Sender:         sender,
		Receiver:       receiver,
		MultiplierName: multiplierName,
		DenomsToClaim:  denomsToClaim,
	}
}

// Route return the message type used for routing the message.
func (msg MsgClaimSwapRewardVVesting) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgClaimSwapRewardVVesting) Type() string {
	return "claim_swap_reward_vvesting"
}

// ValidateBasic does a simple validation check that doesn't require access to state.
func (msg MsgClaimSwapRewardVVesting) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty")
	}
	if msg.Receiver.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "receiver address cannot be empty")
	}
	if err := MultiplierName(msg.MultiplierName).IsValid(); err != nil {
		return err
	}
	for i, d := range msg.DenomsToClaim {
		if i >= MaxDenomsToClaim {
			return sdkerrors.Wrapf(ErrInvalidClaimDenoms, "cannot claim more than %d denoms", MaxDenomsToClaim)
		}
		if err := sdk.ValidateDenom(d); err != nil {
			return sdkerrors.Wrap(ErrInvalidClaimDenoms, err.Error())
		}
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgClaimSwapRewardVVesting) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgClaimSwapRewardVVesting) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
