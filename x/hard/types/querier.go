package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Querier routes for the hard module
const (
	QueryGetParams         = "params"
	QueryGetModuleAccounts = "accounts"
	QueryGetDeposits       = "deposits"
	QueryGetClaims         = "claims"
	QueryGetBorrows        = "borrows"
	QueryGetBorrowed       = "borrowed"
)

// QueryDepositParams is the params for a filtered deposit query
type QueryDepositParams struct {
	Owner sdk.AccAddress `json:"owner" yaml:"owner"`
}

// NewQueryDepositParams creates a new QueryDepositParams
func NewQueryDepositParams(owner sdk.AccAddress) QueryDepositParams {
	return QueryDepositParams{
		Owner: owner,
	}
}

// QueryDepositsParams is the params for a filtered deposit query
type QueryDepositsParams struct {
	Page         int            `json:"page" yaml:"page"`
	Limit        int            `json:"limit" yaml:"limit"`
	DepositDenom string         `json:"deposit_denom" yaml:"deposit_denom"`
	Owner        sdk.AccAddress `json:"owner" yaml:"owner"`
}

// NewQueryDepositsParams creates a new QueryDepositsParams
func NewQueryDepositsParams(page, limit int, depositDenom string, owner sdk.AccAddress) QueryDepositsParams {
	return QueryDepositsParams{
		Page:         page,
		Limit:        limit,
		DepositDenom: depositDenom,
		Owner:        owner,
	}
}

// QueryClaimParams is the params for a filtered claim query
type QueryClaimParams struct {
	Page         int            `json:"page" yaml:"page"`
	Limit        int            `json:"limit" yaml:"limit"`
	DepositDenom string         `json:"deposit_denom" yaml:"deposit_denom"`
	Owner        sdk.AccAddress `json:"owner" yaml:"owner"`
	ClaimType    ClaimType      `json:"claim_type" yaml:"claim_type"`
}

// NewQueryClaimParams creates a new QueryClaimParams
func NewQueryClaimParams(page, limit int, depositDenom string, owner sdk.AccAddress, claimType ClaimType) QueryClaimParams {
	return QueryClaimParams{
		Page:         page,
		Limit:        limit,
		DepositDenom: depositDenom,
		Owner:        owner,
		ClaimType:    claimType,
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
	Page        int            `json:"page" yaml:"page"`
	Limit       int            `json:"limit" yaml:"limit"`
	Owner       sdk.AccAddress `json:"owner" yaml:"owner"`
	BorrowDenom string         `json:"borrow_denom" yaml:"borrow_denom"`
}

// NewQueryBorrowsParams creates a new QueryBorrowsParams
func NewQueryBorrowsParams(page, limit int, owner sdk.AccAddress, borrowDenom string) QueryBorrowsParams {
	return QueryBorrowsParams{
		Page:        page,
		Limit:       limit,
		Owner:       owner,
		BorrowDenom: borrowDenom,
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

// QueryBorrowParams is the params for a current borrow balance query
type QueryBorrowParams struct {
	Owner sdk.AccAddress `json:"owner" yaml:"owner"`
}

// NewQueryBorrowParams creates a new QueryBorrowParams
func NewQueryBorrowParams(owner sdk.AccAddress) QueryBorrowParams {
	return QueryBorrowParams{
		Owner: owner,
	}
}
