package types

import (
	fmt "fmt"

	sdkmath "cosmossdk.io/math"
)

// FractionalAmount represents a fractional amount between the valid range of 1
// and maxFractionalAmount. This wraps an sdkmath.Int to provide additional
// validation methods so it can be re-used in multiple places.
type FractionalAmount struct {
	sdkmath.Int
}

// NewFractionalAmountFromInt creates a new FractionalAmount from an sdkmath.Int.
func NewFractionalAmountFromInt(i sdkmath.Int) FractionalAmount {
	return FractionalAmount{i}
}

// NewFractionalAmount creates a new FractionalAmount from an int64.
func NewFractionalAmount(i int64) FractionalAmount {
	return FractionalAmount{sdkmath.NewInt(i)}
}

// Validate checks if the FractionalAmount is valid.
func (f FractionalAmount) Validate() error {
	if f.IsNil() {
		return fmt.Errorf("nil amount")
	}

	if !f.IsPositive() {
		return fmt.Errorf("non-positive amount %v", f)
	}

	if f.GT(maxFractionalAmount) {
		return fmt.Errorf("amount %v exceeds max of %v", f, maxFractionalAmount)
	}

	return nil
}
