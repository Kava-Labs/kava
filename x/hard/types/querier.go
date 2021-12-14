package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// Querier routes for the hard module
const (
	QueryGetParams           = "params"
	QueryGetModuleAccounts   = "accounts"
	QueryGetDeposits         = "deposits"
	QueryGetUnsyncedDeposits = "unsynced-deposits"
	QueryGetTotalDeposited   = "total-deposited"
	QueryGetBorrows          = "borrows"
	QueryGetUnsyncedBorrows  = "unsynced-borrows"
	QueryGetTotalBorrowed    = "total-borrowed"
	QueryGetInterestRate     = "interest-rate"
	QueryGetReserves         = "reserves"
	QueryGetInterestFactors  = "interest-factors"
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

// QueryUnsyncedDepositsParams is the params for a filtered unsynced deposit query.
type QueryUnsyncedDepositsParams struct {
	Page  int            `json:"page" yaml:"page"`
	Limit int            `json:"limit" yaml:"limit"`
	Denom string         `json:"denom" yaml:"denom"`
	Owner sdk.AccAddress `json:"owner" yaml:"owner"`
}

// NewQueryUnsyncedDepositsParams creates a new QueryUnsyncedDepositsParams
func NewQueryUnsyncedDepositsParams(page, limit int, denom string, owner sdk.AccAddress) QueryUnsyncedDepositsParams {
	return QueryUnsyncedDepositsParams{
		Page:  page,
		Limit: limit,
		Denom: denom,
		Owner: owner,
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

// ModAccountWithCoins includes the module account with its coins
type ModAccountWithCoins struct {
	Account authtypes.ModuleAccountI `json:"account" yaml:"account"`
	Coins   sdk.Coins                `json:"coins" yaml:"coins"`
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

// QueryUnsyncedBorrowsParams is the params for a filtered unsynced borrows query
type QueryUnsyncedBorrowsParams struct {
	Page  int            `json:"page" yaml:"page"`
	Limit int            `json:"limit" yaml:"limit"`
	Owner sdk.AccAddress `json:"owner" yaml:"owner"`
	Denom string         `json:"denom" yaml:"denom"`
}

// NewQueryUnsyncedBorrowsParams creates a new QueryUnsyncedBorrowsParams
func NewQueryUnsyncedBorrowsParams(page, limit int, owner sdk.AccAddress, denom string) QueryUnsyncedBorrowsParams {
	return QueryUnsyncedBorrowsParams{
		Page:  page,
		Limit: limit,
		Owner: owner,
		Denom: denom,
	}
}

// QueryTotalBorrowedParams is the params for a filtered total borrowed coins query
type QueryTotalBorrowedParams struct {
	Denom string `json:"denom" yaml:"denom"`
}

// NewQueryTotalBorrowedParams creates a new QueryTotalBorrowedParams
func NewQueryTotalBorrowedParams(denom string) QueryTotalBorrowedParams {
	return QueryTotalBorrowedParams{
		Denom: denom,
	}
}

// QueryTotalDepositedParams is the params for a filtered total deposited coins query
type QueryTotalDepositedParams struct {
	Denom string `json:"denom" yaml:"denom"`
}

// NewQueryTotalDepositedParams creates a new QueryTotalDepositedParams
func NewQueryTotalDepositedParams(denom string) QueryTotalDepositedParams {
	return QueryTotalDepositedParams{
		Denom: denom,
	}
}

// QueryInterestRateParams is the params for a filtered interest rate query
type QueryInterestRateParams struct {
	Denom string `json:"denom" yaml:"denom"`
}

// NewQueryInterestRateParams creates a new QueryInterestRateParams
func NewQueryInterestRateParams(denom string) QueryInterestRateParams {
	return QueryInterestRateParams{
		Denom: denom,
	}
}

// NewMoneyMarketInterestRate returns a new instance of MoneyMarketInterestRate
func NewMoneyMarketInterestRate(denom string, supplyInterestRate, borrowInterestRate sdk.Dec) MoneyMarketInterestRate {
	return MoneyMarketInterestRate{
		Denom:              denom,
		SupplyInterestRate: supplyInterestRate.String(),
		BorrowInterestRate: borrowInterestRate.String(),
	}
}

// MoneyMarketInterestRates is a slice of MoneyMarketInterestRate
type MoneyMarketInterestRates []MoneyMarketInterestRate

// QueryReservesParams is the params for a filtered reserves query
type QueryReservesParams struct {
	Denom string `json:"denom" yaml:"denom"`
}

// NewQueryReservesParams creates a new QueryReservesParams
func NewQueryReservesParams(denom string) QueryReservesParams {
	return QueryReservesParams{
		Denom: denom,
	}
}

// QueryInterestFactorsParams is the params for a filtered interest factors query
type QueryInterestFactorsParams struct {
	Denom string `json:"denom" yaml:"denom"`
}

// NewQueryInterestFactorsParams creates a new QueryInterestFactorsParams
func NewQueryInterestFactorsParams(denom string) QueryInterestFactorsParams {
	return QueryInterestFactorsParams{
		Denom: denom,
	}
}

// NewInterestFactor returns a new instance of InterestFactor
func NewInterestFactor(denom string, supplyInterestFactor, borrowInterestFactor sdk.Dec) InterestFactor {
	return InterestFactor{
		Denom:                denom,
		SupplyInterestFactor: supplyInterestFactor.String(),
		BorrowInterestFactor: borrowInterestFactor.String(),
	}
}

// InterestFactors is a slice of InterestFactor
type InterestFactors []InterestFactor
