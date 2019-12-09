package types

// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
	Params        Params `json:"params" yaml:"params"`
	CDPs          CDPs   `json:"cdps" yaml:"cdps"`
	StartingCdpID uint64 `json:"starting_cdp_id" yaml:"starting_cdp_id"`
	// don't need to setup CollateralStates as they are created as needed
}

// DefaultGenesisState returns a default genesis state
// TODO make this empty, load test values independent
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:        DefaultParams(),
		CDPs:          CDPs{},
		StartingCdpID: DefaultCdpStartingID,
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
