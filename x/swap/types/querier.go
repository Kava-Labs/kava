package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// Querier routes for the swap module
const (
	QueryGetParams   = "params"
	QueryGetDeposits = "deposits"
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
