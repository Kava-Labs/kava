package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// QueryGetParams command for params query
	QueryGetParams = "parameters"
	// QueryGetDeposits command for deposits query
	QueryGetDeposits = "deposits"
)

// QueryDepositsParams is the params for a filtered deposit query
type QueryDepositsParams struct {
	Page  int            `json:"page" yaml:"page"`
	Limit int            `json:"limit" yaml:"limit"`
	Denom string         `json:"denom" yaml:"denom"`
	Owner sdk.AccAddress `json:"owner" yaml:"owner"`
}

// NewQueryDepositsParams creates a new QueryDepositsParams
func NewQueryDepositsParams(page, limit int, denom string, owner sdk.AccAddress) QueryDepositsParams {
	return QueryDepositsParams{
		Page:  page,
		Limit: limit,
		Denom: denom,
		Owner: owner,
	}
}
