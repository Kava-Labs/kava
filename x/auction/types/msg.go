package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// ensure Msg interface compliance at compile time
var _ sdk.Msg = &MsgPlaceBid{}

// MsgPlaceBid is the message type used to place a bid on any type of auction.
type MsgPlaceBid struct {
	AuctionID uint64
	Bidder    sdk.AccAddress // This can be a buyer (who increments bid), or a seller (who decrements lot) TODO rename to be clearer?
	Bid       sdk.Coin
	Lot       sdk.Coin
}

// NewMsgPlaceBid returns a new MsgPlaceBid.
func NewMsgPlaceBid(auctionID uint64, bidder sdk.AccAddress, bid sdk.Coin, lot sdk.Coin) MsgPlaceBid {
	return MsgPlaceBid{
		AuctionID: auctionID,
		Bidder:    bidder,
		Bid:       bid,
		Lot:       lot,
	}
}

// Route return the message type used for routing the message.
func (msg MsgPlaceBid) Route() string { return RouterKey }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgPlaceBid) Type() string { return "place_bid" }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgPlaceBid) ValidateBasic() sdk.Error {
	if msg.Bidder.Empty() {
		return sdk.ErrInternal("invalid (empty) bidder address")
	}
	if msg.Bid.Amount.LT(sdk.ZeroInt()) {
		return sdk.ErrInternal("invalid (negative) bid amount")
	}
	if msg.Lot.Amount.LT(sdk.ZeroInt()) {
		return sdk.ErrInternal("invalid (negative) lot amount")
	}
	// TODO check coin denoms
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