package types

// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
	Params LiquidatorParams `json:"liquidator_params" yaml:"liquidator_params"`
}

// DefaultGenesisState returns a default genesis state
// TODO pick better values
func DefaultGenesisState() GenesisState {
	return GenesisState{
		DefaultParams(),
	}
}

// ValidateGenesis performs basic validation of genesis data returning an error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}
	return nil
}
