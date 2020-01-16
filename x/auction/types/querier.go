package types

import (
	"strings"
)

const (
	// QueryGetAuction command for getting the information about a particular auction
	QueryGetAuction = "getauctions"
	QueryGetParams  = "params"
)

// QueryResAuctions Result Payload for an auctions query
type QueryResAuctions []string

// implement fmt.Stringer
func (n QueryResAuctions) String() string {
	return strings.Join(n[:], "\n")
}

// QueryAllAuctionParams is the params for an auctions query
type QueryAllAuctionParams struct {
	Page  int `json"page:" yaml:"page"`
	Limit int `json"limit:" yaml:"limit"`
}

// NewQueryAllAuctionParams creates a new QueryAllAuctionParams
func NewQueryAllAuctionParams(page int, limit int) QueryAllAuctionParams {
	return QueryAllAuctionParams{
		Page:  page,
		Limit: limit,
	}
}
