package types

import (
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter keys
var ()

// NewParams creates a new Params object
func NewParams() Params {
	return Params{}
}

// DefaultParams default params for savings
func DefaultParams() Params {
	return NewParams()
}

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of savings module's parameters.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{}
}

// Validate ensure that params have valid values
func (p Params) Validate() error {
	return nil
}
