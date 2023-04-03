package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

// ensure Msg interface compliance at compile time
var (
	_ sdk.Msg            = &MsgFundCommunityPool{}
	_ legacytx.LegacyMsg = &MsgFundCommunityPool{}
)

// NewMsgFundCommunityPool returns a new MsgFundCommunityPool
func NewMsgFundCommunityPool(depositor sdk.AccAddress, amount sdk.Coins) MsgFundCommunityPool {
	return MsgFundCommunityPool{
		Depositor: depositor.String(),
		Amount:    amount,
	}
}

// Route return the message type used for routing the message.
func (msg MsgFundCommunityPool) Route() string { return ModuleName }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgFundCommunityPool) Type() string { return sdk.MsgTypeURL(&msg) }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgFundCommunityPool) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Depositor)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
	}

	if msg.Amount.IsAnyNil() || !msg.Amount.IsValid() || msg.Amount.IsZero() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "'%s'", msg.Amount)
	}

	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgFundCommunityPool) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgFundCommunityPool) GetSigners() []sdk.AccAddress {
	depositor, err := sdk.AccAddressFromBech32(msg.Depositor)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{depositor}
}
