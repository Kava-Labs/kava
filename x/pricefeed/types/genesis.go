package types

import (
	"bytes"
)

// NewGenesisState creates a new genesis state for the pricefeed module
func NewGenesisState(p Params, pp []PostedPrice) GenesisState {
	return GenesisState{
		Params:       p,
		PostedPrices: pp,
	}
}

// DefaultGenesisState defines default GenesisState for pricefeed
func DefaultGenesisState() GenesisState {
	return NewGenesisState(
		DefaultParams(),
		[]PostedPrice{},
	)
}

// Equal checks whether two gov GenesisState structs are equivalent
func (gs GenesisState) Equal(gs2 GenesisState) bool {
	b1 := ModuleCdc.MustMarshal(&gs)
	b2 := ModuleCdc.MustMarshal(&gs2)
	return bytes.Equal(b1, b2)
}

// IsEmpty returns true if a GenesisState is empty
func (gs GenesisState) IsEmpty() bool {
	return gs.Equal(GenesisState{})
}

// Validate performs basic validation of genesis data returning an
// error for any failed validation criteria.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	return ValidatePostedPrices(gs.PostedPrices)
}
