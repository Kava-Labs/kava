package types

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// ensure Msg interface compliance at compile time
var _ sdk.Msg = &MsgPlaceBid{}

// NewMsgPlaceBid returns a new MsgPlaceBid.
func NewMsgPlaceBid(auctionID uint64, bidder string, amt sdk.Coin) MsgPlaceBid {
	return MsgPlaceBid{
		AuctionId: auctionID,
		Bidder:    bidder,
		Amount:    amt,
	}
}

// Route return the message type used for routing the message.
func (msg MsgPlaceBid) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgPlaceBid) Type() string { return "place_bid" }

// ValidateBasic does a simple validation check that doesn't require access to state.
func (msg MsgPlaceBid) ValidateBasic() error {
	if msg.AuctionId == 0 {
		return errors.New("auction id cannot be zero")
	}
	_, err := sdk.AccAddressFromBech32(msg.Bidder)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "bidder address cannot be empty or invalid")
	}
	if !msg.Amount.IsValid() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "bid amount %s", msg.Amount)
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgPlaceBid) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgPlaceBid) GetSigners() []sdk.AccAddress {
	bidder, err := sdk.AccAddressFromBech32(msg.Bidder)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{bidder}
}
