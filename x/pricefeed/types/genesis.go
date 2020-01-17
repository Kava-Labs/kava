package types

import (
	"bytes"
)

// GenesisState - pricefeed state that must be provided at genesis
type GenesisState struct {
	Params       Params        `json:"params" yaml:"params"`
	PostedPrices []PostedPrice `json:"posted_prices" yaml:"posted_prices"`
}

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
	b1 := ModuleCdc.MustMarshalBinaryBare(gs)
	b2 := ModuleCdc.MustMarshalBinaryBare(gs2)
	return bytes.Equal(b1, b2)
}

// IsEmpty returns true if a GenesisState is empty
func (gs GenesisState) IsEmpty() bool {
	return gs.Equal(GenesisState{})
}

// ValidateGenesis performs basic validation of genesis data returning an
// error for any failed validation criteria.
func (gs GenesisState) Validate() error {

	if err := gs.Params.Validate(); err != nil {
		return err
	}
	return nil
}
