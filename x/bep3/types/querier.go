package types

import (
	cmm "github.com/tendermint/tendermint/libs/common"
	cmn "github.com/tendermint/tendermint/libs/common"
)

const (
	// QueryGetHTLT command for getting info about an HTLT
	QueryGetHTLT = "htlt"
	// QueryGetHTLTs command for getting a list of HTLTs
	QueryGetHTLTs = "htlts"
	// QueryGetParams command for getting module params
	QueryGetParams = "params"
)

// QueryHTLTByID contains the params for query 'custom/bep3/htlt'
type QueryHTLTByID struct {
	SwapID cmm.HexBytes
}

// NewQueryHTLTByID creates a new QueryHTLTByID
func NewQueryHTLTByID(swapBytes cmn.HexBytes) QueryHTLTByID {
	return QueryHTLTByID{
		SwapID: swapBytes,
	}
}

// QueryHTLTsParams contains the params for a HTLTs query
type QueryHTLTsParams struct {
	Page  int `json:"page" yaml:"page"`
	Limit int `json:"limit" yaml:"limit"`
}

// NewQueryHTLTsParams creates a new QueryHTLTsParams
func NewQueryHTLTsParams(page int, limit int) QueryHTLTsParams {
	return QueryHTLTsParams{
		Page:  page,
		Limit: limit,
	}
}
