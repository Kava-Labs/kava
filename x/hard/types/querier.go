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
	QueryGetBorrow         = "borrow"
)

// QueryDepositParams is the params for a filtered deposit query
type QueryDepositParams struct {
	Page         int            `json:"page" yaml:"page"`
	Limit        int            `json:"limit" yaml:"limit"`
	DepositDenom string         `json:"deposit_denom" yaml:"deposit_denom"`
	Owner        sdk.AccAddress `json:"owner" yaml:"owner"`
}

// NewQueryDepositParams creates a new QueryDepositParams
func NewQueryDepositParams(page, limit int, depositDenom string, owner sdk.AccAddress) QueryDepositParams {
	return QueryDepositParams{
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

// QueryBorrowParams is the params for a filtered borrow query
type QueryBorrowParams struct {
	Page        int            `json:"page" yaml:"page"`
	Limit       int            `json:"limit" yaml:"limit"`
	Owner       sdk.AccAddress `json:"owner" yaml:"owner"`
	BorrowDenom string         `json:"borrow_denom" yaml:"borrow_denom"`
}

// NewQueryBorrowParams creates a new QueryBorrowParams
func NewQueryBorrowParams(page, limit int, owner sdk.AccAddress, depositDenom string) QueryBorrowParams {
	return QueryBorrowParams{
		Page:        page,
		Limit:       limit,
		Owner:       owner,
		BorrowDenom: depositDenom,
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

// QueryBorrow is the params for a current borrow balance query
type QueryBorrow struct {
	Owner sdk.AccAddress `json:"owner" yaml:"owner"`
}

// NewQueryBorrow creates a new QueryBorrow
func NewQueryBorrow(owner sdk.AccAddress) QueryBorrow {
	return QueryBorrow{
		Owner: owner,
	}
}
