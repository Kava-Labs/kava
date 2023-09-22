package types

// NewGenesisState returns a new genesis state object
func NewGenesisState(params Params) GenesisState {
	return GenesisState{
		Params: params,
	}
}

// DefaultGenesisState returns default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState(
		DefaultParams(),
	)
}

// Validate checks the params are valid
func (gs GenesisState) Validate() error {
	return gs.Params.Validate()
}
