package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter keys
var (
	KeySupportedDenoms     = []byte("SupportedDenoms")
	DefaultSupportedDenoms = []string{}
)

// NewParams creates a new Params object
func NewParams(supportedDenoms []string) Params {
	return Params{
		SupportedDenoms: supportedDenoms,
	}
}

// DefaultParams default params for savings
func DefaultParams() Params {
	return NewParams(DefaultSupportedDenoms)
}

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of savings module's parameters.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeySupportedDenoms, &p.SupportedDenoms, validateSupportedDenoms),
	}
}

// Validate ensure that params have valid values
func (p Params) Validate() error {
	return validateSupportedDenoms(p.SupportedDenoms)
}

func validateSupportedDenoms(i interface{}) error {
	supportedDenoms, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	seenDenoms := make(map[string]bool)
	for _, denom := range supportedDenoms {
		if seenDenoms[denom] {
			return fmt.Errorf("duplicated denom %s", denom)
		}
		seenDenoms[denom] = true
	}
	return nil
}
