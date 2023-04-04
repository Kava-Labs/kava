package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter keys and default values
var (
	KeyAllowedPools     = []byte("AllowedPools")
	KeySwapFee          = []byte("SwapFee")
	DefaultAllowedPools = AllowedPools{}
	DefaultSwapFee      = sdk.ZeroDec()
	MaxSwapFee          = sdk.OneDec()
)

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

// ParamKeyTable for swap module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyAllowedPools, &p.AllowedPools, validateAllowedPoolsParams),
		paramtypes.NewParamSetPair(KeySwapFee, &p.SwapFee, validateSwapFee),
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

	if swapFee.IsNil() || swapFee.IsNegative() || swapFee.GTE(MaxSwapFee) {
		return fmt.Errorf(fmt.Sprintf("invalid swap fee: %s", swapFee))
	}

	return nil
}

// NewAllowedPool returns a new AllowedPool object
func NewAllowedPool(tokenA, tokenB string) AllowedPool {
	return AllowedPool{
		TokenA: tokenA,
		TokenB: tokenB,
	}
}

// Validate validates allowedPool attributes and returns an error if invalid
func (p AllowedPool) Validate() error {
	err := sdk.ValidateDenom(p.TokenA)
	if err != nil {
		return err
	}

	err = sdk.ValidateDenom(p.TokenB)
	if err != nil {
		return err
	}

	// Ensure there is no colon in the token denoms as they are used as separators
	// and is now valid in Cosmos denoms.
	if strings.Contains(p.TokenA, ":") {
		return fmt.Errorf("tokenA cannot have colons in the denom: %s", p.TokenA)
	}

	if strings.Contains(p.TokenB, ":") {
		return fmt.Errorf("tokenB cannot have colons in the denom: %s", p.TokenB)
	}

	if p.TokenA == p.TokenB {
		return fmt.Errorf(
			"pool cannot have two tokens of the same type, received '%s' and '%s'",
			p.TokenA, p.TokenB,
		)
	}

	if p.TokenA > p.TokenB {
		return fmt.Errorf(
			"invalid token order: '%s' must come before '%s'",
			p.TokenB, p.TokenA,
		)
	}

	return nil
}

// Name returns the name for the allowed pool
func (p AllowedPool) Name() string {
	return PoolID(p.TokenA, p.TokenB)
}

// String pretty prints the allowedPool
func (p AllowedPool) String() string {
	return fmt.Sprintf(`AllowedPool:
  Name: %s
	Token A: %s
	Token B: %s
`, p.Name(), p.TokenA, p.TokenB)
}

// AllowedPools is a slice of AllowedPool
type AllowedPools []AllowedPool

// NewAllowedPools returns AllowedPools from the provided values
func NewAllowedPools(allowedPools ...AllowedPool) AllowedPools {
	return AllowedPools(allowedPools)
}

// Validate validates each allowedPool and returns an error if there are any duplicates
func (p AllowedPools) Validate() error {
	seenAllowedPools := make(map[string]bool)
	for _, allowedPool := range p {
		err := allowedPool.Validate()
		if err != nil {
			return err
		}

		if seen := seenAllowedPools[allowedPool.Name()]; seen {
			return fmt.Errorf("duplicate pool: %s", allowedPool.Name())
		}
		seenAllowedPools[allowedPool.Name()] = true
	}

	return nil
}
