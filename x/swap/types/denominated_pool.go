package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DenominatedPool implements a denominated constant-product liquidity pool
type DenominatedPool struct {
	// all pool operations are implemented in a unitless base pool
	pool *BasePool
	// track units of the reserveA and reserveB in base pool
	denomA string
	denomB string
}

// NewDenominatedPool creates a new denominated pool from reserve coins
func NewDenominatedPool(reserves sdk.Coins) (*DenominatedPool, error) {
	if len(reserves) != 2 {
		return nil, sdkerrors.Wrap(ErrInvalidPool, "reserves must have two denominations")
	}

	// Coins should always sorted, so this is deterministic, though it does not need to be.
	// The base pool calculation results do not depend on reserve order.
	reservesA := reserves[0]
	reservesB := reserves[1]

	pool, err := NewBasePool(reservesA.Amount, reservesB.Amount)
	if err != nil {
		return nil, err
	}

	return &DenominatedPool{
		pool:   pool,
		denomA: reservesA.Denom,
		denomB: reservesB.Denom,
	}, nil
}

// NewDenominatedPoolWithExistingShares creates a new denominated pool from reserve coins
func NewDenominatedPoolWithExistingShares(reserves sdk.Coins, totalShares sdk.Int) (*DenominatedPool, error) {
	if len(reserves) != 2 {
		return nil, sdkerrors.Wrap(ErrInvalidPool, "reserves must have two denominations")
	}

	// Coins should always sorted, so this is deterministic, though it does not need to be.
	// The base pool calculation results do not depend on reserve order.
	reservesA := reserves[0]
	reservesB := reserves[1]

	pool, err := NewBasePoolWithExistingShares(reservesA.Amount, reservesB.Amount, totalShares)
	if err != nil {
		return nil, err
	}

	return &DenominatedPool{
		pool:   pool,
		denomA: reservesA.Denom,
		denomB: reservesB.Denom,
	}, nil
}

// Reserves returns the reserves held in the pool
func (p *DenominatedPool) Reserves() sdk.Coins {
	return p.coins(p.pool.ReservesA(), p.pool.ReservesB())
}

// TotalShares returns the total shares for the pool
func (p *DenominatedPool) TotalShares() sdk.Int {
	return p.pool.TotalShares()
}

// IsEmpty returns true if the pool is empty
func (p *DenominatedPool) IsEmpty() bool {
	return p.pool.IsEmpty()
}

// AddLiquidity adds liquidity to the reserves and returns the added amount and shares created
func (p *DenominatedPool) AddLiquidity(deposit sdk.Coins) (sdk.Coins, sdk.Int) {
	desiredA := deposit.AmountOf(p.denomA)
	desiredB := deposit.AmountOf(p.denomB)

	actualA, actualB, shares := p.pool.AddLiquidity(desiredA, desiredB)

	return p.coins(actualA, actualB), shares
}

// RemoveLiquidity removes liquidity from the pool
func (p *DenominatedPool) RemoveLiquidity(shares sdk.Int) sdk.Coins {
	withdrawnA, withdrawnB := p.pool.RemoveLiquidity(shares)

	return p.coins(withdrawnA, withdrawnB)
}

// ShareValue returns the value of the provided shares
func (p *DenominatedPool) ShareValue(shares sdk.Int) sdk.Coins {
	valueA, valueB := p.pool.ShareValue(shares)

	return p.coins(valueA, valueB)
}

// Swap swaps a coin in one reserve for the other.  Panics if denom
// does not match the pool.  Returns the amount received for the input coins,
// and the amount paid by the fee
func (p *DenominatedPool) Swap(coin sdk.Coin, fee sdk.Dec) (sdk.Coin, sdk.Coin) {
	var (
		swapResult sdk.Int
		feePaid    sdk.Int
	)

	switch coin.Denom {
	case p.denomA:
		swapResult, feePaid = p.pool.SwapAForB(coin.Amount, fee)
		return p.coinB(swapResult), p.coinA(feePaid)
	case p.denomB:
		swapResult, feePaid = p.pool.SwapBForA(coin.Amount, fee)
		return p.coinA(swapResult), p.coinB(feePaid)
	default:
		panic(fmt.Sprintf("invalid denomination: denom '%s' does not match pool reserves", coin.Denom))
	}
}

// coins returns a new coins slice with correct reserve denoms from ordered sdk.Ints
func (p *DenominatedPool) coins(amountA, amountB sdk.Int) sdk.Coins {
	return sdk.NewCoins(p.coinA(amountA), p.coinB(amountB))
}

// coinA returns a new coin denominated in denomA
func (p *DenominatedPool) coinA(amount sdk.Int) sdk.Coin {
	return sdk.NewCoin(p.denomA, amount)
}

// coinA returns a new coin denominated in denomB
func (p *DenominatedPool) coinB(amount sdk.Int) sdk.Coin {
	return sdk.NewCoin(p.denomB, amount)
}
