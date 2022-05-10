package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
)

const (
	// QueryGetAssetSupply command for getting info about an asset's supply
	QueryGetAssetSupply = "supply"
	// QueryGetAssetSupplies command for getting a list of asset supplies
	QueryGetAssetSupplies = "supplies"
	// QueryGetAtomicSwap command for getting info about an atomic swap
	QueryGetAtomicSwap = "swap"
	// QueryGetAtomicSwaps command for getting a list of atomic swaps
	QueryGetAtomicSwaps = "swaps"
	// QueryGetParams command for getting module params
	QueryGetParams = "parameters"
)

// Legacy querier requests

// QueryAssetSupply contains the params for query 'custom/bep3/supply'
type QueryAssetSupply struct {
	Denom string `json:"denom" yaml:"denom"`
}

// NewQueryAssetSupply creates a new QueryAssetSupply
func NewQueryAssetSupply(denom string) QueryAssetSupply {
	return QueryAssetSupply{
		Denom: denom,
	}
}

// QueryAssetSupplies contains the params for an AssetSupplies query
type QueryAssetSupplies struct {
	Page  int `json:"page" yaml:"page"`
	Limit int `json:"limit" yaml:"limit"`
}

// NewQueryAssetSupplies creates a new QueryAssetSupplies
func NewQueryAssetSupplies(page int, limit int) QueryAssetSupplies {
	return QueryAssetSupplies{
		Page:  page,
		Limit: limit,
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
	Page       int            `json:"page" yaml:"page"`
	Limit      int            `json:"limit" yaml:"limit"`
	Involve    sdk.AccAddress `json:"involve" yaml:"involve"`
	Expiration uint64         `json:"expiration" yaml:"expiration"`
	Status     SwapStatus     `json:"status" yaml:"status"`
	Direction  SwapDirection  `json:"direction" yaml:"direction"`
}

// NewQueryAtomicSwaps creates a new instance of QueryAtomicSwaps
func NewQueryAtomicSwaps(page, limit int, involve sdk.AccAddress, expiration uint64,
	status SwapStatus, direction SwapDirection,
) QueryAtomicSwaps {
	return QueryAtomicSwaps{
		Page:       page,
		Limit:      limit,
		Involve:    involve,
		Expiration: expiration,
		Status:     status,
		Direction:  direction,
	}
}
