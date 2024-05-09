package types

import (
	fmt "fmt"
	"strings"
)

// FractionalBalances is a slice of FractionalBalance
type FractionalBalances []FractionalBalance

// Len returns the length of the slice.
func (fbs FractionalBalances) Len() int {
	return len(fbs)
}

// Validate returns an error if any FractionalBalance in the slice is invalid.
func (fbs FractionalBalances) Validate() error {
	seenAddresses := make(map[string]struct{})

	for _, fb := range fbs {
		// Ensure there is no white-space before/after address as that can
		// bypass duplicate address check.
		// Upper/lowercase is not an issue as the address is a bech32 encoded
		// string which only allows lowercase.
		addressClean := strings.TrimSpace(fb.Address)

		// If this is a duplicate address, return an error
		if _, found := seenAddresses[addressClean]; found {
			return fmt.Errorf("duplicate address: %v", fb.Address)
		}

		// Individual FractionalBalance validation
		if err := fb.Validate(); err != nil {
			return fmt.Errorf("invalid fractional balance for %s: %v", fb.Address, err)
		}

		// Mark it as seen
		seenAddresses[addressClean] = struct{}{}
	}

	return nil
}
