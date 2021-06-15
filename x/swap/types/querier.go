package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// Querier routes for the swap module
const (
	QueryGetParams   = "params"
	QueryGetDeposits = "deposits"
	QueryGetPool     = "pool"
	QueryGetPools    = "pools"
)

// QueryDepositsParams is the params for a filtered deposits query
type QueryDepositsParams struct {
	Owner sdk.AccAddress `json:"owner" yaml:"owner"`
	Pool  string         `json:"pool" yaml:"pool"`
}

// NewQueryDepositsParams creates a new QueryDepositsParams
func NewQueryDepositsParams(owner sdk.AccAddress, pool string) QueryDepositsParams {
	return QueryDepositsParams{
		Owner: owner,
		Pool:  pool,
	}
}

// QueryPoolParams is the params for a filtered pool query
type QueryPoolParams struct {
	Pool string `json:"pool" yaml:"pool"`
}

// NewQueryPoolParams creates a new QueryPoolParams
func NewQueryPoolParams(pool string) QueryPoolParams {
	return QueryPoolParams{
		Pool: pool,
	}
}

type PoolStatsQueryResult struct {
	Coins       sdk.Coins `json:"coins" yaml:"coins"`
	TotalShares sdk.Int   `json:"total_shares" yaml:"total_shares"`
}

func NewPoolStatsQueryResult(coins sdk.Coins, totalShares sdk.Int) PoolStatsQueryResult {
	return PoolStatsQueryResult{
		Coins:       coins,
		TotalShares: totalShares,
	}
}

// PoolStatsQueryResults is a slice of PoolStatsQueryResult
type PoolStatsQueryResults []PoolStatsQueryResult
