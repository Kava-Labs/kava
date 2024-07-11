package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
)

// NewGenesisState creates a new genesis state.
func NewGenesisState(
	balances FractionalBalances,
	remainder sdkmath.Int,
) *GenesisState {
	return &GenesisState{
		Balances:  balances,
		Remainder: remainder,
	}
}

// DefaultGenesisState returns a default genesis state.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(FractionalBalances{}, sdkmath.ZeroInt())
}

// Validate performs basic validation of genesis data returning an  error for
// any failed validation criteria.
func (gs *GenesisState) Validate() error {
	// Validate all FractionalBalances
	if err := gs.Balances.Validate(); err != nil {
		return fmt.Errorf("invalid balances: %w", err)
	}

	if gs.Remainder.IsNil() {
		return fmt.Errorf("nil remainder amount")
	}

	// Validate remainder, 0 <= remainder <= maxFractionalAmount
	if gs.Remainder.IsNegative() {
		return fmt.Errorf("negative remainder amount %s", gs.Remainder)
	}

	if gs.Remainder.GTE(conversionFactor) {
		return fmt.Errorf("remainder %v exceeds max of %v", gs.Remainder, conversionFactor.SubRaw(1))
	}

	// Determine if sum(fractionalBalances) + remainder = whole integer value
	// i.e total of all fractional balances + remainder == 0 fractional digits
	sum := gs.Balances.SumAmount()
	sumWithRemainder := sum.Add(gs.Remainder)

	offBy := sumWithRemainder.Mod(conversionFactor)

	if !offBy.IsZero() {
		return fmt.Errorf(
			"sum of fractional balances %v + remainder %v is not a multiple of %v",
			sum,
			gs.Remainder,
			conversionFactor,
		)
	}

	return nil
}

// TotalAmountWithRemainder returns the total amount of all balances in the
// genesis state, including both fractional balances and the remainder. A bit
// more verbose WithRemainder to ensure its clearly different from SumAmount().
func (gs *GenesisState) TotalAmountWithRemainder() sdkmath.Int {
	return gs.Balances.SumAmount().Add(gs.Remainder)
}
