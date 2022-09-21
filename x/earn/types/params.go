package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// Parameter keys and default values
var (
	KeyAllowedVaults     = []byte("AllowedVaults")
	DefaultAllowedVaults = AllowedVaults{
		// ukava - Community Pool
		NewAllowedVault(
			"ukava",
			StrategyTypes{STRATEGY_TYPE_SAVINGS},
			true,
			[]sdk.AccAddress{authtypes.NewModuleAddress(distrtypes.ModuleName)},
		),
		// usdx
		NewAllowedVault(
			"usdx",
			StrategyTypes{STRATEGY_TYPE_HARD},
			false,
			[]sdk.AccAddress{},
		),
		NewAllowedVault(
			"bkava",
			StrategyTypes{STRATEGY_TYPE_SAVINGS},
			false,
			[]sdk.AccAddress{},
		),
		NewAllowedVault(
			"erc20/multichain/usdc",
			StrategyTypes{STRATEGY_TYPE_SAVINGS},
			false,
			[]sdk.AccAddress{},
		),
	}
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
