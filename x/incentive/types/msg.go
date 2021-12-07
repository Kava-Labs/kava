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

// NewMsgClaimUSDXMintingReward returns a new MsgClaimUSDXMintingReward.
func NewMsgClaimUSDXMintingReward(sender string, multiplierName string) MsgClaimUSDXMintingReward {
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
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty or invalid")
	}
	if err := NewMultiplierNameFromString(msg.MultiplierName).IsValid(); err != nil {
		return err
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgClaimUSDXMintingReward) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgClaimUSDXMintingReward) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// NewMsgClaimUSDXMintingRewardVVesting returns a new MsgClaimUSDXMintingReward.
func NewMsgClaimUSDXMintingRewardVVesting(sender, receiver string, multiplierName string) MsgClaimUSDXMintingRewardVVesting {
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
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty or invalid")
	}
	_, err = sdk.AccAddressFromBech32(msg.Receiver)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "receiver address cannot be empty or invalid")
	}
	if err := NewMultiplierNameFromString(msg.MultiplierName).IsValid(); err != nil {
		return err
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgClaimUSDXMintingRewardVVesting) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgClaimUSDXMintingRewardVVesting) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// NewMsgClaimHardReward returns a new MsgClaimHardReward.
func NewMsgClaimHardReward(sender string, denomsToClaim Selections) MsgClaimHardReward {
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
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty or invalid")
	}
	if err := msg.DenomsToClaim.Validate(); err != nil {
		return err
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgClaimHardReward) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgClaimHardReward) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// NewMsgClaimHardRewardVVesting returns a new MsgClaimHardRewardVVesting.
func NewMsgClaimHardRewardVVesting(sender, receiver string, denomsToClaim Selections) MsgClaimHardRewardVVesting {
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
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty or invalid")
	}
	_, err = sdk.AccAddressFromBech32(msg.Receiver)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "receiver address cannot be empty or invalid")
	}
	if err := msg.DenomsToClaim.Validate(); err != nil {
		return err
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgClaimHardRewardVVesting) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgClaimHardRewardVVesting) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// NewMsgClaimDelegatorReward returns a new MsgClaimDelegatorReward.
func NewMsgClaimDelegatorReward(sender string, denomsToClaim Selections) MsgClaimDelegatorReward {
	return MsgClaimDelegatorReward{
		Sender:        sender,
		DenomsToClaim: denomsToClaim,
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
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty or invalid")
	}
	if err := msg.DenomsToClaim.Validate(); err != nil {
		return err
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgClaimDelegatorReward) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgClaimDelegatorReward) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// MsgClaimDelegatorRewardVVesting returns a new MsgClaimDelegatorRewardVVesting.
func NewMsgClaimDelegatorRewardVVesting(sender, receiver string, denomsToClaim Selections) MsgClaimDelegatorRewardVVesting {
	return MsgClaimDelegatorRewardVVesting{
		Sender:        sender,
		Receiver:      receiver,
		DenomsToClaim: denomsToClaim,
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
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty or invalid")
	}
	_, err = sdk.AccAddressFromBech32(msg.Receiver)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "receiver address cannot be empty or invalid")
	}
	if err := msg.DenomsToClaim.Validate(); err != nil {
		return err
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgClaimDelegatorRewardVVesting) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgClaimDelegatorRewardVVesting) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// NewMsgClaimSwapReward returns a new MsgClaimSwapReward.
func NewMsgClaimSwapReward(sender string, denomsToClaim Selections) MsgClaimSwapReward {
	return MsgClaimSwapReward{
		Sender:        sender,
		DenomsToClaim: denomsToClaim,
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
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty or invalid")
	}
	if err := msg.DenomsToClaim.Validate(); err != nil {
		return err
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgClaimSwapReward) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgClaimSwapReward) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// MsgClaimSwapRewardVVesting returns a new MsgClaimSwapRewardVVesting.
func NewMsgClaimSwapRewardVVesting(sender, receiver string, denomsToClaim Selections) MsgClaimSwapRewardVVesting {
	return MsgClaimSwapRewardVVesting{
		Sender:        sender,
		Receiver:      receiver,
		DenomsToClaim: denomsToClaim,
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
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty or invalid")
	}
	_, err = sdk.AccAddressFromBech32(msg.Receiver)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "receiver address cannot be empty or invalid")
	}
	if err := msg.DenomsToClaim.Validate(); err != nil {
		return err
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgClaimSwapRewardVVesting) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgClaimSwapRewardVVesting) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}
