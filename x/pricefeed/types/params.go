package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter keys
var (
	KeyMarkets     = []byte("Markets")
	DefaultMarkets = []Market{}
)

// NewParams creates a new AssetParams object
func NewParams(markets []Market) Params {
	return Params{
		Markets: markets,
	}
}

// DefaultParams default params for pricefeed
func DefaultParams() Params {
	return NewParams(DefaultMarkets)
}

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of pricefeed module's parameters.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMarkets, &p.Markets, validateMarketParams),
	}
}

// Validate ensure that params have valid values
func (p Params) Validate() error {
	return validateMarketParams(p.Markets)
}

func validateMarketParams(i interface{}) error {
	markets, ok := i.(Markets)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return markets.Validate()
}
