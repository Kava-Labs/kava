package types

import "fmt"

// Validate performs basic validation of genesis data returning an  error for
// any failed validation criteria.
func (gs *GenesisState) Validate() error {
	if err := gs.Balances.Validate(); err != nil {
		return fmt.Errorf("invalid balances: %w", err)
	}

	// TODO:
	// - Validate remainder amount
	// - Validate sum(fractionalBalances) + remainder = whole integer value
	// - Cannot validate here: reserve account exists & balance match

	return nil
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(balances FractionalBalances) *GenesisState {
	return &GenesisState{
		Balances: balances,
	}
}

// DefaultGenesisState returns a default genesis state.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(FractionalBalances{})
}
