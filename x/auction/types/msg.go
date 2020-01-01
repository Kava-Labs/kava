package types

import sdk "github.com/cosmos/cosmos-sdk/types"

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
func (msg MsgPlaceBid) Route() string { return "auction" }

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

// The CDP system doesn't need Msgs for starting auctions. But they could be added to allow people to create random auctions of their own, and to make this module more general purpose.

// type MsgStartForwardAuction struct {
// 	Seller sdk.AccAddress
// 	Amount sdk.Coins
// 	// TODO add starting bid amount?
// 	// TODO specify asset denom to be received
// }

// // NewMsgStartAuction returns a new MsgStartAuction.
// func NewMsgStartAuction(seller sdk.AccAddress, amount sdk.Coins, maxBid sdk.Coins) MsgStartAuction {
// 	return MsgStartAuction{
// 		Seller: seller,
// 		Amount: amount,
// 		MaxBid: maxBid,
// 	}
// }

// // Route return the message type used for routing the message.
// func (msg MsgStartAuction) Route() string { return "auction" }

// // Type returns a human-readable string for the message, intended for utilization within tags.
// func (msg MsgStartAuction) Type() string { return "start_auction" }

// // ValidateBasic does a simple validation check that doesn't require access to any other information.
// func (msg MsgStartAuction) ValidateBasic() sdk.Error {
// 	return nil
// }

// // GetSignBytes gets the canonical byte representation of the Msg.
// func (msg MsgStartAuction) GetSignBytes() []byte {
// 	bz := msgCdc.MustMarshalJSON(msg)
// 	return sdk.MustSortJSON(bz)
// }

// // GetSigners returns the addresses of signers that must sign.
// func (msg MsgStartAuction) GetSigners() []sdk.AccAddress {
// 	return []sdk.AccAddress{msg.Seller}
// }
