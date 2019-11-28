package types

// GenesisState is the state that must be provided at genesis.
// TODO What is globaldebt and is is separate from the global debt limit in CdpParams

type GenesisState struct {
	Params Params `json:"params" yaml:"params"`
	CDPs   CDPs   `json:"cdps" yaml:"cdps"`
	// don't need to setup CollateralStates as they are created as needed
}

// DefaultGenesisState returns a default genesis state
// TODO make this empty, load test values independent
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params: DefaultParams(),
		CDPs:   CDPs{},
	}
}

// ValidateGenesis performs basic validation of genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {

	if err := data.Params.Validate(); err != nil {
		return err
	}

	// check global debt is zero - force the chain to always start with zero stable coin, otherwise collateralStatus's will need to be set up as well. - what? This seems indefensible.
	return nil
}
