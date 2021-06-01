package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/params"
)

// Parameter keys and default values
var (
	KeyPairs     = []byte("Pairs")
	DefaultPairs = Pairs{}
)

// Params governance parameters for hard module
type Params struct {
	Pairs Pairs `json:"pairs" yaml:"pairs"`
}

// NewParams returns a new params object
func NewParams(pairs Pairs) Params {
	return Params{
		Pairs: pairs,
	}
}

// DefaultParams returns default params for hard module
func DefaultParams() Params {
	return NewParams(DefaultPairs)
}

// String implements fmt.Stringer
func (p Params) String() string {
	return fmt.Sprintf(`Params:
	Pairs: %s`,
		p.Pairs)
}

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyPairs, &p.Pairs, validatePairsParams),
	}
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {
	return validatePairsParams(p.Pairs)
}

func validatePairsParams(i interface{}) error {
	p, ok := i.(Pairs)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return p.Validate()
}
