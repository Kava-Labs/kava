package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Parameter keys and default values
var (
	KeyPairs       = []byte("Pairs")
	KeySwapFee     = []byte("SwapFee")
	DefaultPairs   = Pairs{}
	DefaultSwapFee = sdk.ZeroDec()
	MaxSwapFee     = sdk.OneDec()
)

// Params are governance parameters for the swap module
type Params struct {
	Pairs   Pairs   `json:"pairs" yaml:"pairs"`
	SwapFee sdk.Dec `json:"swap_fee" yaml:"swap_fee"`
}

// NewParams returns a new params object
func NewParams(pairs Pairs, swapFee sdk.Dec) Params {
	return Params{
		Pairs:   pairs,
		SwapFee: swapFee,
	}
}

// DefaultParams returns default params for swap module
func DefaultParams() Params {
	return NewParams(
		DefaultPairs,
		DefaultSwapFee,
	)
}

// String implements fmt.Stringer
func (p Params) String() string {
	return fmt.Sprintf(`Params:
	Pairs: %s
	SwapFee: %s`,
		p.Pairs, p.SwapFee)
}

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyPairs, &p.Pairs, validatePairsParams),
		params.NewParamSetPair(KeySwapFee, &p.SwapFee, validateSwapFee),
	}
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {
	if err := validatePairsParams(p.Pairs); err != nil {
		return err
	}

	return validateSwapFee(p.SwapFee)
}

func validatePairsParams(i interface{}) error {
	p, ok := i.(Pairs)
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
