package v0_11

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

const (
	// ModuleName name that will be used throughout the module
	ModuleName = "swap"
)

// Parameter keys and default values
var (
	KeyAllowedPools     = []byte("AllowedPools")
	KeySwapFee          = []byte("SwapFee")
	DefaultAllowedPools = AllowedPools{}
	DefaultSwapFee      = sdk.ZeroDec()
	MaxSwapFee          = sdk.OneDec()
)

// AllowedPool defines a tradable pool
type AllowedPool struct {
	TokenA string `json:"token_a" yaml:"token_a"`
	TokenB string `json:"token_b" yaml:"token_b"`
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

// Name returns a unique name for a allowedPool in alphabetical order
func (p AllowedPool) Name() string {
	return fmt.Sprintf("%s/%s", p.TokenA, p.TokenB)
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

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
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

// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
	Params Params `json:"params" yaml:"params"`
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(params Params) GenesisState {
	return GenesisState{
		Params: params,
	}
}

// Validate validates the module's genesis state
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	return nil
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState(
		DefaultParams(),
	)
}
