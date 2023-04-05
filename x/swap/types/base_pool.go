package types

import (
	"fmt"
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var zero = sdk.ZeroInt()

// calculateInitialShares calculates initial shares as sqrt(A*B), the geometric mean of A and B
func calculateInitialShares(reservesA, reservesB sdkmath.Int) sdkmath.Int {
	// Big.Int allows multiplication without overflow at 255 bits.
	// In addition, Sqrt converges to a correct solution for inputs
	// where sdkmath.Int.ApproxSqrt does not converge due to exceeding
	// 100 iterations.
	var result big.Int
	result.Mul(reservesA.BigInt(), reservesB.BigInt()).Sqrt(&result)
	return sdkmath.NewIntFromBigInt(&result)
}

// BasePool implements a unitless constant-product liquidity pool.
//
// The pool is symmetric. For all A,B,s, any operation F on a pool (A,B,s) and pool (B,A,s)
// will result in equal state values of A', B', s': F(A,B,s) => (A',B',s'), F(B,A,s) => (B',A',s')
//
// In addition, the pool is protected from overflow in intermediate calculations, and will
// only overflow when A, B, or s become larger than the max sdkmath.Int.
//
// Pool operations with non-positive values are invalid, and all functions on a pool will panic
// when given zero or negative values.
type BasePool struct {
	reservesA   sdkmath.Int
	reservesB   sdkmath.Int
	totalShares sdkmath.Int
}

// NewBasePool returns a pointer to a base pool with reserves and total shares initialized
func NewBasePool(reservesA, reservesB sdkmath.Int) (*BasePool, error) {
	if reservesA.LTE(zero) || reservesB.LTE(zero) {
		return nil, errorsmod.Wrap(ErrInvalidPool, "reserves must be greater than zero")
	}

	totalShares := calculateInitialShares(reservesA, reservesB)

	return &BasePool{
		reservesA:   reservesA,
		reservesB:   reservesB,
		totalShares: totalShares,
	}, nil
}

// NewBasePoolWithExistingShares returns a pointer to a base pool with existing shares
func NewBasePoolWithExistingShares(reservesA, reservesB, totalShares sdkmath.Int) (*BasePool, error) {
	if reservesA.LTE(zero) || reservesB.LTE(zero) {
		return nil, errorsmod.Wrap(ErrInvalidPool, "reserves must be greater than zero")
	}

	if totalShares.LTE(zero) {
		return nil, errorsmod.Wrap(ErrInvalidPool, "total shares must be greater than zero")
	}

	return &BasePool{
		reservesA:   reservesA,
		reservesB:   reservesB,
		totalShares: totalShares,
	}, nil
}

// ReservesA returns the A reserves of the pool
func (p *BasePool) ReservesA() sdkmath.Int {
	return p.reservesA
}

// ReservesB returns the B reserves of the pool
func (p *BasePool) ReservesB() sdkmath.Int {
	return p.reservesB
}

// IsEmpty returns true if all reserves are zero and
// returns false if reserveA or reserveB is not empty
func (p *BasePool) IsEmpty() bool {
	return p.reservesA.IsZero() && p.reservesB.IsZero()
}

// TotalShares returns the total number of shares in the pool
func (p *BasePool) TotalShares() sdkmath.Int {
	return p.totalShares
}

// AddLiquidity adds liquidity to the pool returns the actual reservesA, reservesB deposits in addition
// to the number of shares created.  The deposits are always less than or equal to the provided and desired
// values.
func (p *BasePool) AddLiquidity(desiredA sdkmath.Int, desiredB sdkmath.Int) (sdkmath.Int, sdkmath.Int, sdkmath.Int) {
	// Panics if provided values are zero
	p.assertDepositsArePositive(desiredA, desiredB)

	// Reinitialize the pool if reserves are empty and return the initialized state.
	if p.IsEmpty() {
		p.reservesA = desiredA
		p.reservesB = desiredB
		p.totalShares = calculateInitialShares(desiredA, desiredB)
		return p.ReservesA(), p.ReservesB(), p.TotalShares()
	}

	// Panics if reserveA or reserveB is zero.
	p.assertReservesArePositive()

	// In order to preserve the reserve ratio of the pool, we must deposit
	// A and B in the same ratio of the existing reserves.  In addition,
	// we should not deposit more funds than requested.
	//
	// To meet these requirements, we first calculate the optimalB to deposit
	// if we keep desiredA fixed.  If this is less than or equal to the desiredB,
	// then we use (desiredA, optimalB) as the deposit.
	//
	// If the optimalB is greater than the desiredB, we calculate the optimalA
	// from the desiredB and use (optimalA, desiredB) as the deposit.
	//
	// These optimal values are calculated as:
	//
	// optimalB = reservesB * desiredA / reservesA
	// optimalA = reservesA * desiredB / reservesB
	//
	// Which shows us:
	//
	// if optimalB < desiredB then optimalA > desiredA
	// if optimalB = desiredB then optimalA = desiredA
	// if optimalB > desiredB then optimalA < desiredA
	//
	// so we first check if optimalB <= desiredB, then deposit
	// (desiredA, optimalB) else deposit (optimalA, desiredA).
	//
	// In order avoid precision loss, we rearrange the inequality
	// of optimalB <= desiredB
	// from:
	//   reservesB * desiredA / reservesA <= desiredB
	// to:
	//   reservesB * desiredA <= desiredB * reservesA
	//
	// which also shares the same intermediate products
	// as the calculations for optimalB and optimalA.
	actualA := desiredA.BigInt()
	actualB := desiredB.BigInt()

	// productA = reservesB * desiredA
	var productA big.Int
	productA.Mul(p.reservesB.BigInt(), actualA)

	// productB = reservesA * desiredB
	var productB big.Int
	productB.Mul(p.reservesA.BigInt(), actualB)

	// optimalB <= desiredB
	if productA.Cmp(&productB) <= 0 {
		actualB.Quo(&productA, p.reservesA.BigInt())
	} else { // optimalA < desiredA
		actualA.Quo(&productB, p.reservesB.BigInt())
	}

	var sharesA big.Int
	sharesA.Mul(actualA, p.totalShares.BigInt()).Quo(&sharesA, p.reservesA.BigInt())

	var sharesB big.Int
	sharesB.Mul(actualB, p.totalShares.BigInt()).Quo(&sharesB, p.reservesB.BigInt())

	// a/A and b/B may not be equal due to discrete math and truncation errors,
	// so use the smallest deposit ratio to calculate the number of shares
	//
	// If we do not use the min or max ratio, then the result becomes
	// dependent on the order of reserves in the pool
	//
	// Min is used to always ensure the share ratio is never larger
	// than the deposit ratio for either A or B, ensuring there are no
	// cases where a withdraw will allow funds to be removed at a higher ratio
	// than it was deposited.
	var shares sdkmath.Int
	if sharesA.Cmp(&sharesB) <= 0 {
		shares = sdkmath.NewIntFromBigInt(&sharesA)
	} else {
		shares = sdkmath.NewIntFromBigInt(&sharesB)
	}

	depositA := sdkmath.NewIntFromBigInt(actualA)
	depositB := sdkmath.NewIntFromBigInt(actualB)

	// update internal pool state
	p.reservesA = p.reservesA.Add(depositA)
	p.reservesB = p.reservesB.Add(depositB)
	p.totalShares = p.totalShares.Add(shares)

	return depositA, depositB, shares
}

// RemoveLiquidity removes liquidity from the pool and panics if the
// shares provided are greater than the total shares of the pool
// or the shares are not positive.
// In addition, also panics if reserves go negative, which should not happen.
// If panic occurs, it is a bug.
func (p *BasePool) RemoveLiquidity(shares sdkmath.Int) (sdkmath.Int, sdkmath.Int) {
	// calculate amount to withdraw from the pool based
	// on the number of shares provided. s/S * reserves
	withdrawA, withdrawB := p.ShareValue(shares)

	// update internal pool state
	p.reservesA = p.reservesA.Sub(withdrawA)
	p.reservesB = p.reservesB.Sub(withdrawB)
	p.totalShares = p.totalShares.Sub(shares)

	// Panics if reserveA or reserveB are negative
	// A zero value (100% withdraw) is OK and should not panic.
	p.assertReservesAreNotNegative()

	return withdrawA, withdrawB
}

// SwapExactAForB trades an exact value of a for b.  Returns the positive amount b
// that is removed from the pool and the portion of a that is used for paying the fee.
func (p *BasePool) SwapExactAForB(a sdkmath.Int, fee sdk.Dec) (sdkmath.Int, sdkmath.Int) {
	b, feeValue := p.calculateOutputForExactInput(a, p.reservesA, p.reservesB, fee)

	p.assertInvariantAndUpdateReserves(
		p.reservesA.Add(a), feeValue, p.reservesB.Sub(b), sdk.ZeroInt(),
	)

	return b, feeValue
}

// SwapExactBForA trades an exact value of b for a.  Returns the positive amount a
// that is removed from the pool and the portion of b that is used for paying the fee.
func (p *BasePool) SwapExactBForA(b sdkmath.Int, fee sdk.Dec) (sdkmath.Int, sdkmath.Int) {
	a, feeValue := p.calculateOutputForExactInput(b, p.reservesB, p.reservesA, fee)

	p.assertInvariantAndUpdateReserves(
		p.reservesA.Sub(a), sdk.ZeroInt(), p.reservesB.Add(b), feeValue,
	)

	return a, feeValue
}

// calculateOutputForExactInput calculates the output amount of a swap using a fixed input, returning this amount in
// addition to the amount of input that is used to pay the fee.
//
// The fee is ceiled, ensuring a minimum fee of 1 and ensuring fees of a trade can not be reduced
// by splitting a trade into multiple trades.
//
// The swap output is truncated to ensure the pool invariant is always greater than or equal to the previous invariant.
func (p *BasePool) calculateOutputForExactInput(in, inReserves, outReserves sdkmath.Int, fee sdk.Dec) (sdkmath.Int, sdkmath.Int) {
	p.assertSwapInputIsValid(in)
	p.assertFeeIsValid(fee)

	inAfterFee := sdk.NewDecFromInt(in).Mul(sdk.OneDec().Sub(fee)).TruncateInt()

	var result big.Int
	result.Mul(outReserves.BigInt(), inAfterFee.BigInt())
	result.Quo(&result, inReserves.Add(inAfterFee).BigInt())

	out := sdkmath.NewIntFromBigInt(&result)
	feeValue := in.Sub(inAfterFee)

	return out, feeValue
}

// SwapAForExactB trades a for an exact b.  Returns the positive amount a
// that is added to the pool, and the portion of a that is used to pay the fee.
func (p *BasePool) SwapAForExactB(b sdkmath.Int, fee sdk.Dec) (sdkmath.Int, sdkmath.Int) {
	a, feeValue := p.calculateInputForExactOutput(b, p.reservesB, p.reservesA, fee)

	p.assertInvariantAndUpdateReserves(
		p.reservesA.Add(a), feeValue, p.reservesB.Sub(b), sdk.ZeroInt(),
	)

	return a, feeValue
}

// SwapBForExactA trades b for an exact a.  Returns the positive amount b
// that is added to the pool, and the portion of b that is used to pay the fee.
func (p *BasePool) SwapBForExactA(a sdkmath.Int, fee sdk.Dec) (sdkmath.Int, sdkmath.Int) {
	b, feeValue := p.calculateInputForExactOutput(a, p.reservesA, p.reservesB, fee)

	p.assertInvariantAndUpdateReserves(
		p.reservesA.Sub(a), sdk.ZeroInt(), p.reservesB.Add(b), feeValue,
	)

	return b, feeValue
}

// calculateInputForExactOutput calculates the input amount of a swap using a fixed output, returning this amount in
// addition to the amount of input that is used to pay the fee.
//
// The fee is ceiled, ensuring a minimum fee of 1 and ensuring fees of a trade can not be reduced
// by splitting a trade into multiple trades.
//
// The swap input is ceiled to ensure the pool invariant is always greater than or equal to the previous invariant.
func (p *BasePool) calculateInputForExactOutput(out, outReserves, inReserves sdkmath.Int, fee sdk.Dec) (sdkmath.Int, sdkmath.Int) {
	p.assertSwapOutputIsValid(out, outReserves)
	p.assertFeeIsValid(fee)

	var result big.Int
	result.Mul(inReserves.BigInt(), out.BigInt())

	newOutReserves := outReserves.Sub(out)
	var remainder big.Int
	result.QuoRem(&result, newOutReserves.BigInt(), &remainder)

	inWithoutFee := sdkmath.NewIntFromBigInt(&result)
	if remainder.Sign() != 0 {
		inWithoutFee = inWithoutFee.Add(sdk.OneInt())
	}

	in := sdk.NewDecFromInt(inWithoutFee).Quo(sdk.OneDec().Sub(fee)).Ceil().TruncateInt()
	feeValue := in.Sub(inWithoutFee)

	return in, feeValue
}

// ShareValue returns the value of the provided shares and panics
// if the shares are greater than the total shares of the pool or
// if the shares are not positive.
func (p *BasePool) ShareValue(shares sdkmath.Int) (sdkmath.Int, sdkmath.Int) {
	p.assertSharesArePositive(shares)
	p.assertSharesAreLessThanTotal(shares)

	var resultA big.Int
	resultA.Mul(p.reservesA.BigInt(), shares.BigInt())
	resultA.Quo(&resultA, p.totalShares.BigInt())

	var resultB big.Int
	resultB.Mul(p.reservesB.BigInt(), shares.BigInt())
	resultB.Quo(&resultB, p.totalShares.BigInt())

	return sdkmath.NewIntFromBigInt(&resultA), sdkmath.NewIntFromBigInt(&resultB)
}

// assertInvariantAndUpdateRerserves asserts the constant product invariant is not violated, subtracting
// any fees first, then updates the pool reserves.  Panics if invariant is violated.
func (p *BasePool) assertInvariantAndUpdateReserves(newReservesA, feeA, newReservesB, feeB sdkmath.Int) {
	var invariant big.Int
	invariant.Mul(p.reservesA.BigInt(), p.reservesB.BigInt())

	var newInvariant big.Int
	newInvariant.Mul(newReservesA.Sub(feeA).BigInt(), newReservesB.Sub(feeB).BigInt())

	p.assertInvariant(&invariant, &newInvariant)

	p.reservesA = newReservesA
	p.reservesB = newReservesB
}

// assertSwapInputIsValid checks if the provided swap input is positive
// and panics if it is 0 or negative
func (p *BasePool) assertSwapInputIsValid(input sdkmath.Int) {
	if !input.IsPositive() {
		panic("invalid value: swap input must be positive")
	}
}

// assertSwapOutputIsValid checks if the provided swap input is positive and
// less than the provided reserves.
func (p *BasePool) assertSwapOutputIsValid(output sdkmath.Int, reserves sdkmath.Int) {
	if !output.IsPositive() {
		panic("invalid value: swap output must be positive")
	}

	if output.GTE(reserves) {
		panic("invalid value: swap output must be less than reserves")
	}
}

// assertFeeIsValid checks if the provided fee is less
func (p *BasePool) assertFeeIsValid(fee sdk.Dec) {
	if fee.IsNegative() || fee.GTE(sdk.OneDec()) {
		panic("invalid value: fee must be between 0 and 1")
	}
}

// assertSharesPositive panics if shares is zero or negative
func (p *BasePool) assertSharesArePositive(shares sdkmath.Int) {
	if !shares.IsPositive() {
		panic("invalid value: shares must be positive")
	}
}

// assertSharesLessThanTotal panics if the number of shares is greater than the total shares
func (p *BasePool) assertSharesAreLessThanTotal(shares sdkmath.Int) {
	if shares.GT(p.totalShares) {
		panic(fmt.Sprintf("out of bounds: shares %s > total shares %s", shares, p.totalShares))
	}
}

// assertDepositsPositive panics if a deposit is zero or negative
func (p *BasePool) assertDepositsArePositive(depositA, depositB sdkmath.Int) {
	if !depositA.IsPositive() {
		panic("invalid value: deposit A must be positive")
	}

	if !depositB.IsPositive() {
		panic("invalid state: deposit B must be positive")
	}
}

// assertReservesArePositive panics if any reserves are zero.  This is an invalid
// state that should never happen.  If this panic is seen, it is a bug.
func (p *BasePool) assertReservesArePositive() {
	if !p.reservesA.IsPositive() {
		panic("invalid state: reserves A must be positive")
	}

	if !p.reservesB.IsPositive() {
		panic("invalid state: reserves B must be positive")
	}
}

// assertReservesAreNotNegative panics if any reserves are negative.  This is an invalid
// state that should never happen.  If this panic is seen, it is a bug.
func (p *BasePool) assertReservesAreNotNegative() {
	if p.reservesA.IsNegative() {
		panic("invalid state: reserves A can not be negative")
	}

	if p.reservesB.IsNegative() {
		panic("invalid state: reserves B can not be negative")
	}
}

// assertInvariant panics if the new invariant is less than the previous invariant.  This
// is an invalid state that should never happen.  If this panic is seen, it is a bug.
func (p *BasePool) assertInvariant(prevInvariant, newInvariant *big.Int) {
	// invariant > newInvariant
	if prevInvariant.Cmp(newInvariant) == 1 {
		panic(fmt.Sprintf("invalid state: invariant %s decreased to %s", prevInvariant.String(), newInvariant.String()))
	}
}
