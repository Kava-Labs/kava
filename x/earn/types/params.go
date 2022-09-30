package types

import (
	fmt "fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter keys and default values
var (
	KeyAllowedVaults     = []byte("AllowedVaults")
	DefaultAllowedVaults = AllowedVaults{}
)

// NewParams returns a new params object
func NewParams(allowedVaults AllowedVaults) Params {
	return Params{
		AllowedVaults: allowedVaults,
	}
}

// DefaultParams returns default params for earn module
func DefaultParams() Params {
	return NewParams(DefaultAllowedVaults)
}

// ParamKeyTable for earn module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyAllowedVaults, &p.AllowedVaults, validateAllowedVaultsParams),
	}
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {
	return p.AllowedVaults.Validate()
}

func validateAllowedVaultsParams(i interface{}) error {
	p, ok := i.(AllowedVaults)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return p.Validate()
}
