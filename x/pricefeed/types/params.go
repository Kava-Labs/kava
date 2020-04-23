package types

import (
	"fmt"
	"strings"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Parameter keys
var (
	KeyMarkets     = []byte("Markets")
	DefaultMarkets = Markets{}
)

// Params params for pricefeed. Can be altered via governance
type Params struct {
	Markets Markets `json:"markets" yaml:"markets"` //  Array containing the markets supported by the pricefeed
}

// NewParams creates a new AssetParams object
func NewParams(markets Markets) Params {
	return Params{
		Markets: markets,
	}
}

// DefaultParams default params for pricefeed
func DefaultParams() Params {
	return NewParams(DefaultMarkets)
}

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of pricefeed module's parameters.
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyMarkets, &p.Markets, validateMarketParams),
	}
}

// String implements fmt.stringer
func (p Params) String() string {
	out := "Params:\n"
	for _, a := range p.Markets {
		out += fmt.Sprintf("%s\n", a.String())
	}
	return strings.TrimSpace(out)
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

	// iterate over assets and verify them
	for _, asset := range markets {
		if strings.TrimSpace(asset.MarketID) == "" {
			return sdkerrors.Wrapf(ErrInvalidMarket, "market id for asset %s cannot be blank", asset)
		}
	}

	return nil
}
