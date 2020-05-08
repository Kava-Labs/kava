package types

// Events for the module
const (
	EventTypeAuctionStart = "auction_start"
	EventTypeAuctionBid   = "auction_bid"
	EventTypeAuctionClose = "auction_close"

	AttributeValueCategory  = ModuleName
	AttributeKeyAuctionID   = "auction_id"
	AttributeKeyAuctionType = "auction_type"
	AttributeKeyBidder      = "bidder"
	AttributeKeyLot         = "lot"
	AttributeKeyMaxBid      = "max_bid"
	AttributeKeyBid         = "bid"
	AttributeKeyEndTime     = "end_time"
	AttributeKeyCloseBlock  = "close_block"
)
