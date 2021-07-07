package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// Querier routes for the swap module
const (
	QueryGetParams   = "params"
	QueryGetDeposits = "deposits"
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
	SharesOwned sdk.Int        `json:"shares_owned" yaml:"shares_owned"`
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
