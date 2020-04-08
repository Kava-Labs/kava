package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ensure Msg interface compliance at compile time
var _ sdk.Msg = &MsgPlaceBid{}

// MsgPlaceBid is the message type used to place a bid on any type of auction.
type MsgPlaceBid struct {
	AuctionID uint64         `json:"auction_id" yaml:"auction_id"`
	Bidder    sdk.AccAddress `json:"bidder" yaml:"bidder"`
	Amount    sdk.Coin       `json:"amount" yaml:"amount"` // The new bid or lot to be set on the auction.
}

// NewMsgPlaceBid returns a new MsgPlaceBid.
func NewMsgPlaceBid(auctionID uint64, bidder sdk.AccAddress, amt sdk.Coin) MsgPlaceBid {
	return MsgPlaceBid{
		AuctionID: auctionID,
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
	if msg.Bidder.Empty() {
		return sdk.ErrInvalidAddress("invalid (empty) bidder address")
	}
	if !msg.Amount.IsValid() {
		return sdk.ErrInvalidCoins(fmt.Sprintf("invalid bid amount: %s", msg.Amount))
	}
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgPlaceBid) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgPlaceBid) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Bidder}
}
