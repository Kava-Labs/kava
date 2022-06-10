package types

// NewGenesisState creates a new genesis state for the liquidstaking module
func NewGenesisState(p Params) GenesisState {
	return GenesisState{
		Params: p,
	}
}

// DefaultGenesisState defines default GenesisState for liquidstaking
func DefaultGenesisState() GenesisState {
	return NewGenesisState(
		DefaultParams(),
	)
}

// Validate performs basic validation of genesis data returning an
// error for any failed validation criteria.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	return nil
}
