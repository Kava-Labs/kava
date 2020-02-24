package types

// Querier routes for the validator vesting module
const (
	QueryCirculatingSupply = "circulating-supply"
)

// QueryCirculatingSupplyParams defines the parameters necessary for querying for all Evidence.
type QueryCirculatingSupplyParams struct {
	Page  int `json:"page" yaml:"page"`
	Limit int `json:"limit" yaml:"limit"`
}

func NewQueryCirculatingSupplyParams(page, limit int) QueryCirculatingSupplyParams {
	return QueryCirculatingSupplyParams{Page: page, Limit: limit}
}
