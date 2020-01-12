package types

const (
	EventTypeAuctionStart = "auction_start"
	EventTypeAuctionBid   = "auction_bid"
	EventTypeAuctionClose = "auction_close"

	AttributeValueCategory  = ModuleName
	AttributeKeyAuctionID   = "auction_id"
	AttributeKeyAuctionType = "auction_type"
	AttributeKeyLotDenom    = "lot_denom"
)
