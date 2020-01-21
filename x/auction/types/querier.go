package types

const (
	// QueryGetAuction is the query path for querying one auction
	QueryGetAuction = "auction"
	// QueryGetAuction is the query path for querying all auctions
	QueryGetAuctions = "auctions"
	// QueryGetAuction is the query path for querying the global auction params
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
