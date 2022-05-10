package types

// NewGenesisState creates a new genesis state for the savings module
func NewGenesisState(p Params, deposits Deposits) GenesisState {
	return GenesisState{
		Params:   p,
		Deposits: deposits,
	}
}

// DefaultGenesisState defines default GenesisState for savings
func DefaultGenesisState() GenesisState {
	return NewGenesisState(
		DefaultParams(),
		Deposits{},
	)
}

// Validate performs basic validation of genesis data returning an
// error for any failed validation criteria.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	return gs.Deposits.Validate()
}
