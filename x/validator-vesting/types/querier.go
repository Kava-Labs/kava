package types

// Querier routes for the validator vesting module
const (
	QueryCirculatingSupply = "circulating-supply"
	QueryTotalSupply       = "total-supply"
)

// QueryCirculatingSupplyParams defines the parameters necessary for querying for all Evidence.
type BaseQueryParams struct {
	Page  int `json:"page" yaml:"page"`
	Limit int `json:"limit" yaml:"limit"`
}

func NewBaseQueryParams(page, limit int) BaseQueryParams {
	return BaseQueryParams{Page: page, Limit: limit}
}
