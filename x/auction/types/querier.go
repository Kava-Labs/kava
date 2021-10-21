package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// QueryGetAuction is the query path for querying one auction
	QueryGetAuction = "auction"
	// QueryGetAuctions is the query path for querying all auctions
	QueryGetAuctions = "auctions"
	// QueryGetParams is the query path for querying the global auction params
	QueryGetParams = "params"
	// QueryNextAuctionID is the query path for querying the id of the next auction
	QueryNextAuctionID = "next-auction-id"
)

// QueryAuctionParams params for query /auction/auction
type QueryAuctionParams struct {
	AuctionID uint64
}

// NewQueryAuctionParams returns a new QueryAuctionParams
func NewQueryAuctionParams(id uint64) QueryAuctionParams {
	return QueryAuctionParams{
		AuctionID: id,
	}
}

// QueryAllAuctionParams is the params for an auctions query
type QueryAllAuctionParams struct {
	Page  int            `json:"page" yaml:"page"`
	Limit int            `json:"limit" yaml:"limit"`
	Type  string         `json:"type" yaml:"type"`
	Owner sdk.AccAddress `json:"owner" yaml:"owner"`
	Denom string         `json:"denom" yaml:"denom"`
	Phase string         `json:"phase" yaml:"phase"`
}

// NewQueryAllAuctionParams creates a new QueryAllAuctionParams
func NewQueryAllAuctionParams(page, limit int, aucType, aucDenom, aucPhase string, aucOwner sdk.AccAddress) QueryAllAuctionParams {
	return QueryAllAuctionParams{
		Page:  page,
		Limit: limit,
		Type:  aucType,
		Owner: aucOwner,
		Denom: aucDenom,
		Phase: aucPhase,
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
