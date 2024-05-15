package types

import (
	fmt "fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
)

// FractionalBalances is a slice of FractionalBalance
type FractionalBalances []FractionalBalance

// Validate returns an error if any FractionalBalance in the slice is invalid.
func (fbs FractionalBalances) Validate() error {
	seenAddresses := make(map[string]struct{})

	for _, fb := range fbs {
		// Individual FractionalBalance validation
		if err := fb.Validate(); err != nil {
			return fmt.Errorf("invalid fractional balance for %s: %w", fb.Address, err)
		}

		// Make addresses all lowercase for unique check, as ALL UPPER is also
		// a valid address.
		lowerAddr := strings.ToLower(fb.Address)

		// If this is a duplicate address, return an error
		if _, found := seenAddresses[lowerAddr]; found {
			return fmt.Errorf("duplicate address %v", lowerAddr)
		}

		// Mark it as seen
		seenAddresses[lowerAddr] = struct{}{}
	}

	return nil
}

// SumAmount returns the sum of all the amounts in the slice.
func (fbs FractionalBalances) SumAmount() sdkmath.Int {
	sum := sdkmath.ZeroInt()

	for _, fb := range fbs {
		sum = sum.Add(fb.Amount)
	}

	return sum
}
