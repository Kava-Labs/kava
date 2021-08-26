package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// Querier routes for the validator vesting module
const (
	QueryCirculatingSupply     = "circulating-supply"
	QueryTotalSupply           = "total-supply"
	QueryCirculatingSupplyHARD = "circulating-supply-hard"
	QueryCirculatingSupplyUSDX = "circulating-supply-usdx"
	QueryTotalSupplyHARD       = "total-supply-hard"
	QueryTotalSupplyUSDX       = "total-supply-usdx"
	QuerySpendableBalance      = "spendable-balance"
)

// BaseQueryParams defines the parameters necessary for querying for all Evidence.
type BaseQueryParams struct {
	Page  int `json:"page" yaml:"page"`
	Limit int `json:"limit" yaml:"limit"`
}

// NewBaseQueryParams returns a new BaseQueryParams
func NewBaseQueryParams(page, limit int) BaseQueryParams {
	return BaseQueryParams{
		Page:  page,
		Limit: limit,
	}
}

type SpendableBalanceParams struct {
	Address sdk.AccAddress `json:"address" yaml:"address"`
}

// NewSpendableBalanceParams creates a new instance of SpendableBalanceParams.
func NewSpendableBalanceParams(addr sdk.AccAddress) SpendableBalanceParams {
	return SpendableBalanceParams{Address: addr}
}
