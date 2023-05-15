package types

import (
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter keys and default values
var (
	KeyEnabledConversionPairs  = []byte("EnabledConversionPairs")
	DefaultConversionPairs     = ConversionPairs{}
	KeyAllowedNativeDenoms     = []byte("AllowedNativeDenoms")
	DefaultAllowedNativeDenoms = AllowedNativeCoinERC20Tokens{}
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
		paramtypes.NewParamSetPair(KeyAllowedNativeDenoms, &p.AllowedNativeDenoms, validateAllowedNativeCoinERC20Tokens),
	}
}

// NewParams returns new evmutil module Params.
func NewParams(
	conversionPairs ConversionPairs,
	allowedNativeDenoms AllowedNativeCoinERC20Tokens,
) Params {
	return Params{
		EnabledConversionPairs: conversionPairs,
		AllowedNativeDenoms:    allowedNativeDenoms,
	}
}

// DefaultParams returns the default parameters for evmutil.
func DefaultParams() Params {
	return NewParams(
		DefaultConversionPairs,
		DefaultAllowedNativeDenoms,
	)
}

// Validate returns an error if the Params is invalid.
func (p *Params) Validate() error {
	if err := p.EnabledConversionPairs.Validate(); err != nil {
		return err
	}
	if err := p.AllowedNativeDenoms.Validate(); err != nil {
		return err
	}
	return nil
}
