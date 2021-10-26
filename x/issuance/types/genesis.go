package types

import "bytes"

var (
	// DefaultAssetSupplies is used to set default asset supplies in default genesis state
	DefaultAssetSupplies = []AssetSupply{}
)

// NewGenesisState returns a new GenesisState
func NewGenesisState(params Params, assetSupplies []AssetSupply) GenesisState {
	return GenesisState{
		Params:        params,
		AssetSupplies: assetSupplies,
	}
}

// DefaultGenesisState returns the default GenesisState for the issuance module
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:        DefaultParams(),
		AssetSupplies: DefaultAssetSupplies,
	}
}

// Validate performs basic validation of genesis data returning an
// error for any failed validation criteria.
func (gs GenesisState) Validate() error {
	for _, supply := range gs.AssetSupplies {
		err := supply.Validate()
		if err != nil {
			return err
		}
	}
	return gs.Params.Validate()
}

// Equal checks whether two GenesisState structs are equivalent
func (gs GenesisState) Equal(gs2 GenesisState) bool {
	b1 := ModuleCdc.MustMarshal(&gs)
	b2 := ModuleCdc.MustMarshal(&gs2)
	return bytes.Equal(b1, b2)
}

// IsEmpty returns true if a GenesisState is empty
func (gs GenesisState) IsEmpty() bool {
	return gs.Equal(GenesisState{})
}
