package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetPoolName returns a pool name from two denoms
func PoolName(denomA string, denomB string) string {
	return fmt.Sprintf("%s/%s", denomA, denomB)
}

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

// Name returns the name for the allowed pool
func (p AllowedPool) Name() string {
	return PoolName(p.TokenA, p.TokenB)
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

// Pool implements a constant-product liquidty pool
type Pool struct {
	ReservesA   sdk.Coin
	ReservesB   sdk.Coin
	TotalShares sdk.Int
}

// NewPool creates a pool from an initial reserve and initializes the total shares
func NewPool(reservesA sdk.Coin, reservesB sdk.Coin) (Pool, error) {
	product := reservesA.Amount.Mul(reservesB.Amount)
	totalShares, err := product.ToDec().ApproxSqrt()

	if err != nil {
		return Pool{}, fmt.Errorf("unable to calculate total shares")
	}

	return Pool{
		ReservesA:   reservesA,
		ReservesB:   reservesB,
		TotalShares: totalShares.TruncateInt(),
	}, nil
}

// Name returns the name for the pool
func (p Pool) Name() string {
	return PoolName(p.ReservesA.Denom, p.ReservesB.Denom)
}

// ShareValue returns the reserves represented by the provided number of shares
func (p Pool) ShareValue(numShares sdk.Int) (sdk.Coins, error) {
	if p.TotalShares.Equal(sdk.ZeroInt()) {
		return sdk.Coins{}, fmt.Errorf("error calculating share value, cannot divide by 0")
	}

	valueA := p.ReservesA.Amount.Mul(numShares).Quo(p.TotalShares)
	valueB := p.ReservesB.Amount.Mul(numShares).Quo(p.TotalShares)

	return sdk.NewCoins(
		sdk.NewCoin(p.ReservesA.Denom, valueA),
		sdk.NewCoin(p.ReservesB.Denom, valueB),
	), nil
}
