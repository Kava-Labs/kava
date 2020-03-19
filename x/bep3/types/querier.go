package types

import (
	cmm "github.com/tendermint/tendermint/libs/common"
	cmn "github.com/tendermint/tendermint/libs/common"
)

const (
	// QueryGetAssetSupplyInfo command for getting info about an Asset's supply
	QueryGetAssetSupplyInfo = "supply"
	// QueryGetAtomicSwap command for getting info about an AtomicSwap
	QueryGetAtomicSwap = "swap"
	// QueryGetAtomicSwaps command for getting a list of AtomicSwaps
	QueryGetAtomicSwaps = "swaps"
	// QueryGetParams command for getting module params
	QueryGetParams = "params"
)

// QueryAssetSupplyInfo contains the params for query 'custom/bep3/supply'
type QueryAssetSupplyInfo struct {
	Denom string `json:"denom" yaml:"denom"`
}

// NewQueryAssetSupplyInfo creates a new QueryAssetSupplyInfo
func NewQueryAssetSupplyInfo(denom string) QueryAssetSupplyInfo {
	return QueryAssetSupplyInfo{
		Denom: denom,
	}
}

// QueryAtomicSwapByID contains the params for query 'custom/bep3/swap'
type QueryAtomicSwapByID struct {
	SwapID cmm.HexBytes `json:"swap_id" yaml:"swap_id"`
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
