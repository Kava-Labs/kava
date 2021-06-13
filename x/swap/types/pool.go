package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetPoolName returns a pool name from two denoms
func PoolName(denomA string, denomB string) string {
	return fmt.Sprintf("%s/%s", denomA, denomB)
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
