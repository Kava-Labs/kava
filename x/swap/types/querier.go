package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Querier routes for the swap module
const (
	QueryGetParams   = "params"
	QueryGetDeposits = "deposits"
	QueryGetPool     = "pool"
	QueryGetPools    = "pools"
)

// QueryDepositsParams is the params for a filtered deposits query
type QueryDepositsParams struct {
	Page  int            `json:"page" yaml:"page"`
	Limit int            `json:"limit" yaml:"limit"`
	Owner sdk.AccAddress `json:"owner" yaml:"owner"`
	Pool  string         `json:"pool" yaml:"pool"`
}

// NewQueryDepositsParams creates a new QueryDepositsParams
func NewQueryDepositsParams(page, limit int, owner sdk.AccAddress, pool string) QueryDepositsParams {
	return QueryDepositsParams{
		Page:  page,
		Limit: limit,
		Owner: owner,
		Pool:  pool,
	}
}

// DepositsQueryResult contains the result of a deposits query
type DepositsQueryResult struct {
	Depositor   sdk.AccAddress `json:"depositor" yaml:"depositor"`
	PoolID      string         `json:"pool_id" yaml:"pool_id"`
	SharesOwned sdkmath.Int    `json:"shares_owned" yaml:"shares_owned"`
	SharesValue sdk.Coins      `json:"shares_value" yaml:"shares_value"`
}

// NewDepositsQueryResult creates a new DepositsQueryResult
func NewDepositsQueryResult(shareRecord ShareRecord, sharesValue sdk.Coins) DepositsQueryResult {
	return DepositsQueryResult{
		Depositor:   shareRecord.Depositor,
		PoolID:      shareRecord.PoolID,
		SharesOwned: shareRecord.SharesOwned,
		SharesValue: sharesValue,
	}
}

// DepositsQueryResults is a slice of DepositsQueryResult
type DepositsQueryResults []DepositsQueryResult

// QueryPoolParams is the params for a pool query
type QueryPoolParams struct {
	Pool string `json:"pool" yaml:"pool"`
}

// NewQueryPoolParams creates a new QueryPoolParams
func NewQueryPoolParams(pool string) QueryPoolParams {
	return QueryPoolParams{
		Pool: pool,
	}
}

// PoolStatsQueryResult contains the result of a pool query
type PoolStatsQueryResult struct {
	Name        string      `json:"name" yaml:"name"`
	Coins       sdk.Coins   `json:"coins" yaml:"coins"`
	TotalShares sdkmath.Int `json:"total_shares" yaml:"total_shares"`
}

// NewPoolStatsQueryResult creates a new PoolStatsQueryResult
func NewPoolStatsQueryResult(name string, coins sdk.Coins, totalShares sdkmath.Int) PoolStatsQueryResult {
	return PoolStatsQueryResult{
		Name:        name,
		Coins:       coins,
		TotalShares: totalShares,
	}
}

// PoolStatsQueryResults is a slice of PoolStatsQueryResult
type PoolStatsQueryResults []PoolStatsQueryResult
