package types

import (
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter keys and default values
var ()

// NewParams returns a new params object
func NewParams() Params {
	return Params{}
}

// DefaultParams returns default params for earn module
func DefaultParams() Params {
	return NewParams()
}

// ParamKeyTable for earn module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		// paramtypes.NewParamSetPair(...),
	}
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {
	// TODO:
	return nil
}
