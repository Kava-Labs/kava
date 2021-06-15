package types

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	zero = sdk.ZeroInt()
)

// BasePool implements a unitless constant-product liquidty pool
// Reserves A is base asset, Reserves B is quote asset
// Math is done using big.Int wheere n
type BasePool struct {
	reservesA   sdk.Int
	reservesB   sdk.Int
	totalShares sdk.Int
}

// NewBasePool returns a new pool initialized with the provided results and total shares
// equal to sqrt(reservesA * reservesB)
func NewBasePool(reservesA, reservesB sdk.Int) (*BasePool, error) {
	if reservesA.LTE(zero) || reservesB.LTE(zero) {
		return nil, sdkerrors.Wrap(ErrInvalidPool, "reserves must be greater than zero")
	}

	// big.Int allows multiplication without overflow at 255 bits.
	// In addition, sqrt converges to a correct solution for inputs
	// that take longer than 100 iterations.  sdk.Int.ApproxSqrt() does
	// not always converge.
	var result big.Int
	result.Mul(reservesA.BigInt(), reservesB.BigInt()).Sqrt(&result)
	totalShares := sdk.NewIntFromBigInt(&result)

	return &BasePool{
		reservesA:   reservesA,
		reservesB:   reservesB,
		totalShares: totalShares,
	}, nil
}

// NewBasePoolWithExistingShares returns a new pool and sets the total number of shares.
func NewBasePoolWithExistingShares(reservesA, reservesB, totalShares sdk.Int) (*BasePool, error) {
	if reservesA.LTE(zero) || reservesB.LTE(zero) {
		return nil, sdkerrors.Wrap(ErrInvalidPool, "reserves must be greater than zero")
	}

	if totalShares.LTE(zero) {
		return nil, sdkerrors.Wrap(ErrInvalidPool, "total shares must be greater than zero")
	}

	return &BasePool{
		reservesA:   reservesA,
		reservesB:   reservesB,
		totalShares: totalShares,
	}, nil
}

// ReservesA returns the base asset reserves of the pool
func (p *BasePool) ReservesA() sdk.Int {
	return p.reservesA
}

// ReservesA returns the quote asset reserves of the pool
func (p *BasePool) ReservesB() sdk.Int {
	return p.reservesB
}

// TotalShares returns the total number of shares in the pool
func (p *BasePool) TotalShares() sdk.Int {
	return p.totalShares
}

// AddLiquidty adds liquidty to the pool returned the actual reservesA, reservesB, and shares created
// actual deposits are always <= desired deposits
func (p *BasePool) AddLiquidity(desiredA sdk.Int, desiredB sdk.Int) (sdk.Int, sdk.Int, sdk.Int) {
	actualA := desiredA.BigInt()
	actualB := desiredB.BigInt()

	var productA big.Int
	productA.Mul(p.reservesB.BigInt(), desiredA.BigInt())

	var productB big.Int
	productB.Mul(p.reservesA.BigInt(), desiredB.BigInt())

	// optimalB <= desiredB
	// reservesB * deisredA / reservesA <= desiredB multiplied by reservesA
	// in order to avoid truncation and loss of precision on division
	if productA.Cmp(&productB) <= 0 {
		actualB.Quo(&productA, p.reservesA.BigInt())
	} else { // optimalA <= desiredA
		actualA.Quo(&productB, p.reservesB.BigInt())
	}

	var sharesA big.Int
	sharesA.Mul(actualA, p.totalShares.BigInt()).Quo(&sharesA, p.reservesA.BigInt())

	var sharesB big.Int
	sharesB.Mul(actualB, p.totalShares.BigInt()).Quo(&sharesB, p.reservesB.BigInt())

	// a/A and b/B may not be equal - use the smallest ratio
	//
	// if we do not use the min or max ratio, then the result becomes
	// dependent on the order on the order of the reserves
	//
	// min is used to always ensure the the share ratio is never larger
	// than the deposit ratio for either A or B
	var shares sdk.Int
	if sharesA.Cmp(&sharesB) <= 0 {
		shares = sdk.NewIntFromBigInt(&sharesA)
	} else {
		shares = sdk.NewIntFromBigInt(&sharesB)
	}

	depositA := sdk.NewIntFromBigInt(actualA)
	depositB := sdk.NewIntFromBigInt(actualB)

	// update internal pool state
	p.reservesA = p.reservesA.Add(depositA)
	p.reservesB = p.reservesB.Add(depositB)
	p.totalShares = p.totalShares.Add(shares)

	return depositA, depositB, shares
}

// RemoveLiquidity removes liquidity from the pool
// panics if shares > totalShares
func (p *BasePool) RemoveLiquidity(shares sdk.Int) (sdk.Int, sdk.Int) {
	// calculate amount to withdraw from the pool based
	// on the number of shares provided
	withdrawA, withdrawB := p.ShareValue(shares)

	// update internal pool state
	p.reservesA = p.reservesA.Sub(withdrawA)
	p.reservesB = p.reservesB.Sub(withdrawB)
	p.totalShares = p.totalShares.Sub(shares)

	return withdrawA, withdrawB
}

// SwapAForB trades a for b with a percentage fee
func (p *BasePool) SwapAForB(a sdk.Int, fee sdk.Dec) sdk.Int {
	return sdk.ZeroInt()
}

// SwapAForB trades b for a with a percentage fee
func (p *BasePool) SwapBForA(b sdk.Int, fee sdk.Dec) sdk.Int {
	return sdk.ZeroInt()
}

// ShareValue returns the value of the provided shares
// panics if shares > totalShares
func (p *BasePool) ShareValue(shares sdk.Int) (sdk.Int, sdk.Int) {
	p.assertSharesLessThanTotal(shares)

	var resultA big.Int
	resultA.Mul(p.reservesA.BigInt(), shares.BigInt())
	resultA.Quo(&resultA, p.totalShares.BigInt())

	var resultB big.Int
	resultB.Mul(p.reservesB.BigInt(), shares.BigInt())
	resultB.Quo(&resultB, p.totalShares.BigInt())

	return sdk.NewIntFromBigInt(&resultA), sdk.NewIntFromBigInt(&resultB)
}

// assertSharesLessThanTotal panics if the number of shares is greater than the total shares
func (p *BasePool) assertSharesLessThanTotal(shares sdk.Int) {
	if shares.GT(p.totalShares) {
		panic(fmt.Sprintf("out of bounds: shares %s > total shares %s", shares, p.totalShares))
	}
}
