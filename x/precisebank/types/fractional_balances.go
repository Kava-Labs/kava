package types

import (
	fmt "fmt"
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

		// If this is a duplicate address, return an error
		if _, found := seenAddresses[fb.Address]; found {
			return fmt.Errorf("duplicate address %v", fb.Address)
		}

		// Mark it as seen
		seenAddresses[fb.Address] = struct{}{}
	}

	return nil
}
