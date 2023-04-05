package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

const MaxDenomsToClaim = 1000

// ensure Msg interface compliance at compile time
var (
	_ sdk.Msg = &MsgClaimUSDXMintingReward{}
	_ sdk.Msg = &MsgClaimHardReward{}
	_ sdk.Msg = &MsgClaimDelegatorReward{}
	_ sdk.Msg = &MsgClaimSwapReward{}
	_ sdk.Msg = &MsgClaimSavingsReward{}
	_ sdk.Msg = &MsgClaimEarnReward{}

	_ legacytx.LegacyMsg = &MsgClaimUSDXMintingReward{}
	_ legacytx.LegacyMsg = &MsgClaimHardReward{}
	_ legacytx.LegacyMsg = &MsgClaimDelegatorReward{}
	_ legacytx.LegacyMsg = &MsgClaimSwapReward{}
	_ legacytx.LegacyMsg = &MsgClaimSavingsReward{}
	_ legacytx.LegacyMsg = &MsgClaimEarnReward{}
)

const (
	TypeMsgClaimUSDXMintingReward = "claim_usdx_minting_reward"
	TypeMsgClaimHardReward        = "claim_hard_reward"
	TypeMsgClaimDelegatorReward   = "claim_delegator_reward"
	TypeMsgClaimSwapReward        = "claim_swap_reward"
	TypeMsgClaimSavingsReward     = "claim_savings_reward"
	TypeMsgClaimEarnReward        = "claim_earn_reward"
)

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
func (msg MsgClaimUSDXMintingReward) Type() string { return TypeMsgClaimUSDXMintingReward }

// ValidateBasic does a simple validation check that doesn't require access to state.
func (msg MsgClaimUSDXMintingReward) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty or invalid")
	}
	if msg.MultiplierName == "" {
		return errorsmod.Wrap(ErrInvalidMultiplier, "multiplier name cannot be empty")
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
	return TypeMsgClaimHardReward
}

// ValidateBasic does a simple validation check that doesn't require access to state.
func (msg MsgClaimHardReward) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty or invalid")
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
	return TypeMsgClaimDelegatorReward
}

// ValidateBasic does a simple validation check that doesn't require access to state.
func (msg MsgClaimDelegatorReward) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty or invalid")
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
	return TypeMsgClaimSwapReward
}

// ValidateBasic does a simple validation check that doesn't require access to state.
func (msg MsgClaimSwapReward) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty or invalid")
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

// NewMsgClaimSavingsReward returns a new MsgClaimSavingsReward.
func NewMsgClaimSavingsReward(sender string, denomsToClaim Selections) MsgClaimSavingsReward {
	return MsgClaimSavingsReward{
		Sender:        sender,
		DenomsToClaim: denomsToClaim,
	}
}

// Route return the message type used for routing the message.
func (msg MsgClaimSavingsReward) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgClaimSavingsReward) Type() string {
	return TypeMsgClaimSavingsReward
}

// ValidateBasic does a simple validation check that doesn't require access to state.
func (msg MsgClaimSavingsReward) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty or invalid")
	}
	if err := msg.DenomsToClaim.Validate(); err != nil {
		return err
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgClaimSavingsReward) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgClaimSavingsReward) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// NewMsgClaimEarnReward returns a new MsgClaimEarnReward.
func NewMsgClaimEarnReward(sender string, denomsToClaim Selections) MsgClaimEarnReward {
	return MsgClaimEarnReward{
		Sender:        sender,
		DenomsToClaim: denomsToClaim,
	}
}

// Route return the message type used for routing the message.
func (msg MsgClaimEarnReward) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgClaimEarnReward) Type() string {
	return TypeMsgClaimEarnReward
}

// ValidateBasic does a simple validation check that doesn't require access to state.
func (msg MsgClaimEarnReward) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be empty or invalid")
	}
	if err := msg.DenomsToClaim.Validate(); err != nil {
		return err
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgClaimEarnReward) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgClaimEarnReward) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}
