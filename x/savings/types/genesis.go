package types

// NewGenesisState creates a new genesis state for the savings module
func NewGenesisState(p Params) GenesisState {
	return GenesisState{
		Params: p,
	}
}

// DefaultGenesisState defines default GenesisState for savings
func DefaultGenesisState() GenesisState {
	return NewGenesisState(
		DefaultParams(),
	)
}

// Validate performs basic validation of genesis data returning an
// error for any failed validation criteria.
func (gs GenesisState) Validate() error {
	return gs.Params.Validate()
}
