package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Querier routes for the hard module
const (
	QueryGetParams         = "params"
	QueryGetModuleAccounts = "accounts"
	QueryGetDeposits       = "deposits"
	QueryGetDeposited      = "deposited"
	QueryGetClaims         = "claims"
	QueryGetBorrows        = "borrows"
	QueryGetBorrowed       = "borrowed"
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

// QueryClaimParams is the params for a filtered claim query
type QueryClaimParams struct {
	Page      int            `json:"page" yaml:"page"`
	Limit     int            `json:"limit" yaml:"limit"`
	Denom     string         `json:"denom" yaml:"denom"`
	Owner     sdk.AccAddress `json:"owner" yaml:"owner"`
	ClaimType ClaimType      `json:"claim_type" yaml:"claim_type"`
}

// NewQueryClaimParams creates a new QueryClaimParams
func NewQueryClaimParams(page, limit int, denom string, owner sdk.AccAddress, claimType ClaimType) QueryClaimParams {
	return QueryClaimParams{
		Page:      page,
		Limit:     limit,
		Denom:     denom,
		Owner:     owner,
		ClaimType: claimType,
	}
}

// QueryAccountParams is the params for a filtered module account query
type QueryAccountParams struct {
	Page  int    `json:"page" yaml:"page"`
	Limit int    `json:"limit" yaml:"limit"`
	Name  string `json:"name" yaml:"name"`
}

// NewQueryAccountParams returns QueryAccountParams
func NewQueryAccountParams(page, limit int, name string) QueryAccountParams {
	return QueryAccountParams{
		Page:  page,
		Limit: limit,
		Name:  name,
	}
}

// QueryBorrowsParams is the params for a filtered borrows query
type QueryBorrowsParams struct {
	Page  int            `json:"page" yaml:"page"`
	Limit int            `json:"limit" yaml:"limit"`
	Owner sdk.AccAddress `json:"owner" yaml:"owner"`
	Denom string         `json:"denom" yaml:"denom"`
}

// NewQueryBorrowsParams creates a new QueryBorrowsParams
func NewQueryBorrowsParams(page, limit int, owner sdk.AccAddress, denom string) QueryBorrowsParams {
	return QueryBorrowsParams{
		Page:  page,
		Limit: limit,
		Owner: owner,
		Denom: denom,
	}
}

// QueryBorrowedParams is the params for a filtered borrowed coins query
type QueryBorrowedParams struct {
	Denom string `json:"denom" yaml:"denom"`
}

// NewQueryBorrowedParams creates a new QueryBorrowedParams
func NewQueryBorrowedParams(denom string) QueryBorrowedParams {
	return QueryBorrowedParams{
		Denom: denom,
	}
}

// QueryDepositedParams is the params for a filtered deposited coins query
type QueryDepositedParams struct {
	Denom string `json:"denom" yaml:"denom"`
}

// NewQueryDepositedParams creates a new QueryDepositedParams
func NewQueryDepositedParams(denom string) QueryDepositedParams {
	return QueryDepositedParams{
		Denom: denom,
	}
}
