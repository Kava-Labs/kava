package types

import "bytes"

// GenesisState is the state that must be provided at genesis for the issuance module
type GenesisState struct {
	Params   Params        `json:"params" yaml:"params"`
	Supplies AssetSupplies `json:"supplies" yaml:"supplies"`
}

// NewGenesisState returns a new GenesisState
func NewGenesisState(params Params, supplies AssetSupplies) GenesisState {
	return GenesisState{
		Params:   params,
		Supplies: supplies,
	}
}

// DefaultGenesisState returns the default GenesisState for the issuance module
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:   DefaultParams(),
		Supplies: AssetSupplies{},
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

// Equal checks whether two GenesisState structs are equivalent
func (gs GenesisState) Equal(gs2 GenesisState) bool {
	b1 := ModuleCdc.MustMarshalBinaryBare(gs)
	b2 := ModuleCdc.MustMarshalBinaryBare(gs2)
	return bytes.Equal(b1, b2)
}

// IsEmpty returns true if a GenesisState is empty
func (gs GenesisState) IsEmpty() bool {
	return gs.Equal(GenesisState{})
}
