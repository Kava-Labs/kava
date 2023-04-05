package v0_15

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName name that will be used throughout the module
	ModuleName = "swap"
)

// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
	Params       Params `json:"params" yaml:"params"`
	PoolRecords  `json:"pool_records" yaml:"pool_records"`
	ShareRecords `json:"share_records" yaml:"share_records"`
}

// Params are governance parameters for the swap module
type Params struct {
	AllowedPools AllowedPools `json:"allowed_pools" yaml:"allowed_pools"`
	SwapFee      sdk.Dec      `json:"swap_fee" yaml:"swap_fee"`
}

// PoolRecords is a slice of PoolRecord
type PoolRecords []PoolRecord

// PoolRecord represents the state of a liquidity pool
// and is used to store the state of a denominated pool
type PoolRecord struct {
	// primary key
	PoolID      string      `json:"pool_id" yaml:"pool_id"`
	ReservesA   sdk.Coin    `json:"reserves_a" yaml:"reserves_a"`
	ReservesB   sdk.Coin    `json:"reserves_b" yaml:"reserves_b"`
	TotalShares sdkmath.Int `json:"total_shares" yaml:"total_shares"`
}

// ShareRecords is a slice of ShareRecord
type ShareRecords []ShareRecord

// ShareRecord stores the shares owned for a depositor and pool
type ShareRecord struct {
	// primary key
	Depositor sdk.AccAddress `json:"depositor" yaml:"depositor"`
	// secondary / sort key
	PoolID      string      `json:"pool_id" yaml:"pool_id"`
	SharesOwned sdkmath.Int `json:"shares_owned" yaml:"shares_owned"`
}

// AllowedPools is a slice of AllowedPool
type AllowedPools []AllowedPool

// AllowedPool defines a tradable pool
type AllowedPool struct {
	TokenA string `json:"token_a" yaml:"token_a"`
	TokenB string `json:"token_b" yaml:"token_b"`
}
