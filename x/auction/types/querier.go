package types

const (
	// QueryGetAuction is the query path for querying one auction
	QueryGetAuction = "auction"
	// QueryGetAuctions is the query path for querying all auctions
	QueryGetAuctions = "auctions"
	// QueryGetParams is the query path for querying the global auction params
	QueryGetParams = "params"
)

// QueryAuctionParams params for query /auction/auction
type QueryAuctionParams struct {
	AuctionID uint64
}

// QueryAllAuctionParams is the params for an auctions query
type QueryAllAuctionParams struct {
	Page  int `json:"page" yaml:"page"`
	Limit int `json:"limit" yaml:"limit"`
}

// NewQueryAllAuctionParams creates a new QueryAllAuctionParams
func NewQueryAllAuctionParams(page int, limit int) QueryAllAuctionParams {
	return QueryAllAuctionParams{
		Page:  page,
		Limit: limit,
	}
}

// AuctionWithPhase augmented type for collateral auctions which includes auction phase for querying
type AuctionWithPhase struct {
	Auction Auction `json:"auction" yaml:"auction"`

	Type  string `json:"type" yaml:"type"`
	Phase string `json:"phase" yaml:"phase"`
}

// NewAuctionWithPhase returns new AuctionWithPhase
func NewAuctionWithPhase(a Auction) AuctionWithPhase {
	return AuctionWithPhase{
		Auction: a,
		Type:    a.GetType(),
		Phase:   a.GetPhase(),
	}
}
