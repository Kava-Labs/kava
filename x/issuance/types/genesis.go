package types

// DefaultSupplies is used to set default asset supplies in default genesis state
var DefaultSupplies = []AssetSupply{}

// NewGenesisState returns a new GenesisState
func NewGenesisState(params Params, supplies []AssetSupply) GenesisState {
	return GenesisState{
		Params:   params,
		Supplies: supplies,
	}
}

// DefaultGenesisState returns the default GenesisState for the issuance module
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:   DefaultParams(),
		Supplies: DefaultSupplies,
	}
}

// Validate performs basic validation of genesis data returning an
// error for any failed validation criteria.
func (gs GenesisState) Validate() error {
	for _, supply := range gs.Supplies {
		err := supply.Validate()
		if err != nil {
			return err
		}
	}
	return gs.Params.Validate()
}
