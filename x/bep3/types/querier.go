package types

import tmbytes "github.com/tendermint/tendermint/libs/bytes"

const (
	// QueryGetAssetSupply command for getting info about an asset's supply
	QueryGetAssetSupply = "supply"
	// QueryGetAtomicSwap command for getting info about an atomic swap
	QueryGetAtomicSwap = "swap"
	// QueryGetAtomicSwaps command for getting a list of atomic swaps
	QueryGetAtomicSwaps = "swaps"
	// QueryGetParams command for getting module params
	QueryGetParams = "parameters"
)

// QueryAssetSupply contains the params for query 'custom/bep3/supply'
type QueryAssetSupply struct {
	Denom tmbytes.HexBytes `json:"denom" yaml:"denom"`
}

// NewQueryAssetSupply creates a new QueryAssetSupply
func NewQueryAssetSupply(denom tmbytes.HexBytes) QueryAssetSupply {
	return QueryAssetSupply{
		Denom: denom,
	}
}

// QueryAtomicSwapByID contains the params for query 'custom/bep3/swap'
type QueryAtomicSwapByID struct {
	SwapID tmbytes.HexBytes `json:"swap_id" yaml:"swap_id"`
}

// NewQueryAtomicSwapByID creates a new QueryAtomicSwapByID
func NewQueryAtomicSwapByID(swapBytes tmbytes.HexBytes) QueryAtomicSwapByID {
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
