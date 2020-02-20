package types

// Events for auction module
const (
	EventTypeAuctionStart = "auction_start"
	EventTypeAuctionBid   = "auction_bid"
	EventTypeAuctionClose = "auction_close"

	AttributeValueCategory  = ModuleName
	AttributeKeyAuctionID   = "auction_id"
	AttributeKeyAuctionType = "auction_type"
	AttributeKeyBidder      = "bidder"
	AttributeKeyBidDenom    = "bid_denom"
	AttributeKeyLotDenom    = "lot_denom"
	AttributeKeyBidAmount   = "bid_amount"
	AttributeKeyLotAmount   = "lot_amount"
	AttributeKeyEndTime     = "end_time"
)
