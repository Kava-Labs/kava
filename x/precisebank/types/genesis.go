package types

// Validate performs basic validation of supply genesis data returning an
// error for any failed validation criteria.
func (gs *GenesisState) Validate() error {
	return nil
}

// NewGenesisState creates a new genesis state.
func NewGenesisState() *GenesisState {
	return &GenesisState{}
}

// DefaultGenesisState returns a default genesis state.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState()
}
