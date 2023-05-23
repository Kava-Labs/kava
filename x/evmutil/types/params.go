package types

import (
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter keys and default values
var (
	KeyEnabledConversionPairs  = []byte("EnabledConversionPairs")
	DefaultConversionPairs     = ConversionPairs{}
	KeyAllowedCosmosDenoms     = []byte("AllowedCosmosDenoms")
	DefaultAllowedCosmosDenoms = AllowedCosmosCoinERC20Tokens{}
)

// ParamKeyTable for evmutil module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value
// pairs pairs of the evmutil module's parameters.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyEnabledConversionPairs, &p.EnabledConversionPairs, validateConversionPairs),
		paramtypes.NewParamSetPair(KeyAllowedCosmosDenoms, &p.AllowedCosmosDenoms, validateAllowedCosmosCoinERC20Tokens),
	}
}

// NewParams returns new evmutil module Params.
func NewParams(
	conversionPairs ConversionPairs,
	allowedCosmosDenoms AllowedCosmosCoinERC20Tokens,
) Params {
	return Params{
		EnabledConversionPairs: conversionPairs,
		AllowedCosmosDenoms:    allowedCosmosDenoms,
	}
}

// DefaultParams returns the default parameters for evmutil.
func DefaultParams() Params {
	return NewParams(
		DefaultConversionPairs,
		DefaultAllowedCosmosDenoms,
	)
}

// Validate returns an error if the Params is invalid.
func (p *Params) Validate() error {
	if err := p.EnabledConversionPairs.Validate(); err != nil {
		return err
	}
	if err := p.AllowedCosmosDenoms.Validate(); err != nil {
		return err
	}
	return nil
}
