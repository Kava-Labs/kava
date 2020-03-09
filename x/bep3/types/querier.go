package types

import (
	cmn "github.com/tendermint/tendermint/libs/common"
)

const (
	// QueryGetAtomicSwap command for getting info about an AtomicSwap
	QueryGetAtomicSwap = "swap"
	// QueryGetAtomicSwaps command for getting a list of AtomicSwaps
	QueryGetAtomicSwaps = "swaps"
	// QueryGetParams command for getting module params
	QueryGetParams = "params"
)

// QueryAtomicSwapByID contains the params for query 'custom/bep3/swap'
type QueryAtomicSwapByID struct {
	SwapID cmn.HexBytes
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
