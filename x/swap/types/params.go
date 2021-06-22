package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Parameter keys and default values
var (
	KeyAllowedPools     = []byte("AllowedPools")
	KeySwapFee          = []byte("SwapFee")
	DefaultAllowedPools = AllowedPools{}
	DefaultSwapFee      = sdk.ZeroDec()
	MaxSwapFee          = sdk.OneDec()
)

// Params are governance parameters for the swap module
type Params struct {
	AllowedPools AllowedPools `json:"allowed_pools" yaml:"allowed_pools"`
	SwapFee      sdk.Dec      `json:"swap_fee" yaml:"swap_fee"`
}

// NewParams returns a new params object
func NewParams(pairs AllowedPools, swapFee sdk.Dec) Params {
	return Params{
		AllowedPools: pairs,
		SwapFee:      swapFee,
	}
}

// DefaultParams returns default params for swap module
func DefaultParams() Params {
	return NewParams(
		DefaultAllowedPools,
		DefaultSwapFee,
	)
}

// String implements fmt.Stringer
func (p Params) String() string {
	return fmt.Sprintf(`Params:
	AllowedPools: %s
	SwapFee: %s`,
		p.AllowedPools, p.SwapFee)
}

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetAllowedPools implements the ParamSet interface and returns all the key/value pairs
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyAllowedPools, &p.AllowedPools, validateAllowedPoolsParams),
		params.NewParamSetPair(KeySwapFee, &p.SwapFee, validateSwapFee),
	}
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {
	if err := validateAllowedPoolsParams(p.AllowedPools); err != nil {
		return err
	}

	return validateSwapFee(p.SwapFee)
}

func validateAllowedPoolsParams(i interface{}) error {
	p, ok := i.(AllowedPools)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return p.Validate()
}

func validateSwapFee(i interface{}) error {
	swapFee, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if swapFee.IsNil() || swapFee.IsNegative() || swapFee.GT(MaxSwapFee) {
		return fmt.Errorf(fmt.Sprintf("invalid swap fee: %s", swapFee))
	}

	return nil
}
