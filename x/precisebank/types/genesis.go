package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
)

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

	if gs.Remainder.GT(MaxFractionalAmount()) {
		return fmt.Errorf("remainder %v exceeds max of %v", gs.Remainder, MaxFractionalAmount())
	}

	// Determine if sum(fractionalBalances) + remainder = whole integer value
	// i.e total of all fractional balances + remainder == 0 fractional digits
	sum := gs.Balances.SumAmount()
	total := sum.Add(gs.Remainder)

	if !total.Mod(ConversionFactor()).IsZero() {
		return fmt.Errorf(
			"sum of fractional balances %v + remainder %v is not a whole integer value %v",
			total,
			sum, gs.Remainder,
		)
	}

	return nil
}

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
