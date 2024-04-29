package types

// NewGenesisState returns a new genesis state object for the module.
func NewGenesisState() *GenesisState {
	return &GenesisState{}
}

// DefaultGenesisState returns the default genesis state for the module.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState()
}

// Validate performs basic validation of genesis data.
func (gs GenesisState) Validate() error {
	return nil
}
