package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Querier routes for the harvest module
const (
	QueryGetParams         = "params"
	QueryGetModuleAccounts = "accounts"
	QueryGetDeposits       = "deposits"
	QueryGetClaims         = "claims"
)

// QueryDepositParams is the params for a filtered deposit query
type QueryDepositParams struct {
	Page         int            `json:"page" yaml:"page"`
	Limit        int            `json:"limit" yaml:"limit"`
	DepositDenom string         `json:"deposit_denom" yaml:"deposit_denom"`
	Owner        sdk.AccAddress `json:"owner" yaml:"owner"`
	DepositType  DepositType    `json:"deposit_type" yaml:"deposit_type"`
}

// NewQueryDepositParams creates a new QueryDepositParams
func NewQueryDepositParams(page, limit int, depositDenom string, owner sdk.AccAddress, depositType DepositType) QueryDepositParams {
	return QueryDepositParams{
		Page:         page,
		Limit:        limit,
		DepositDenom: depositDenom,
		Owner:        owner,
		DepositType:  depositType,
	}
}

// QueryClaimParams is the params for a filtered claim query
type QueryClaimParams struct {
	Page         int            `json:"page" yaml:"page"`
	Limit        int            `json:"limit" yaml:"limit"`
	DepositDenom string         `json:"deposit_denom" yaml:"deposit_denom"`
	Owner        sdk.AccAddress `json:"owner" yaml:"owner"`
	DepositType  DepositType    `json:"deposit_type" yaml:"deposit_type"`
}

// NewQueryClaimParams creates a new QueryClaimParams
func NewQueryClaimParams(page, limit int, depositDenom string, owner sdk.AccAddress, depositType DepositType) QueryClaimParams {
	return QueryClaimParams{
		Page:         page,
		Limit:        limit,
		DepositDenom: depositDenom,
		Owner:        owner,
		DepositType:  depositType,
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
