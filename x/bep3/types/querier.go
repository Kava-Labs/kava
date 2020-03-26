package types

import (
	cmn "github.com/tendermint/tendermint/libs/common"
)

const (
	// QueryGetAssetSupply command for getting info about an asset's supply
	QueryGetAssetSupply = "supply"
	// QueryGetAtomicSwap command for getting info about an atomic swap
	QueryGetAtomicSwap = "swap"
	// QueryGetAtomicSwaps command for getting a list of atomic swaps
	QueryGetAtomicSwaps = "swaps"
	// QueryGetParams command for getting module params
	QueryGetParams = "params"
)

// QueryAssetSupply contains the params for query 'custom/bep3/supply'
type QueryAssetSupply struct {
	Denom cmn.HexBytes `json:"denom" yaml:"denom"`
}

// NewQueryAssetSupply creates a new QueryAssetSupply
func NewQueryAssetSupply(denom cmn.HexBytes) QueryAssetSupply {
	return QueryAssetSupply{
		Denom: denom,
	}
}

// QueryAtomicSwapByID contains the params for query 'custom/bep3/swap'
type QueryAtomicSwapByID struct {
	SwapID cmn.HexBytes `json:"swap_id" yaml:"swap_id"`
}

// NewQueryAtomicSwapByID creates a new QueryAtomicSwapByID
func NewQueryAtomicSwapByID(swapBytes cmn.HexBytes) QueryAtomicSwapByID {
	return QueryAtomicSwapByID{
		SwapID: swapBytes,
	}
}

// QueryAtomicSwaps contains the params for an AtomicSwaps query
type QueryAtomicSwaps struct {
	Page  int `json:"page" yaml:"page"`
	Limit int `json:"limit" yaml:"limit"`
}

// NewQueryAtomicSwaps creates a new QueryAtomicSwaps
func NewQueryAtomicSwaps(page int, limit int) QueryAtomicSwaps {
	return QueryAtomicSwaps{
		Page:  page,
		Limit: limit,
	}
}
