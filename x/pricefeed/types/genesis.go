package types

import (
	"bytes"
	"fmt"
)

// GenesisState - pricefeed state that must be provided at genesis
type GenesisState struct {
	AssetParams  AssetParams   `json:"asset_params" yaml:"asset_params"`
	OracleParams OracleParams  `json:"oracle_params" yaml:"oracle_params"`
	PostedPrices []PostedPrice `json:"posted_prices" yaml:"posted_prices"`
}

// NewGenesisState creates a new genesis state for the pricefeed module
func NewGenesisState(ap AssetParams, op OracleParams, pp []PostedPrice) GenesisState {
	return GenesisState{
		AssetParams:  ap,
		OracleParams: op,
		PostedPrices: pp,
	}
}

// DefaultGenesisState defines default GenesisState for pricefeed
func DefaultGenesisState() GenesisState {
	return NewGenesisState(
		DefaultAssetParams(),
		DefaultOracleParams(),
		[]PostedPrice{},
	)
}

// Equal checks whether two gov GenesisState structs are equivalent
func (data GenesisState) Equal(data2 GenesisState) bool {
	b1 := ModuleCdc.MustMarshalBinaryBare(data)
	b2 := ModuleCdc.MustMarshalBinaryBare(data2)
	return bytes.Equal(b1, b2)
}

// IsEmpty returns true if a GenesisState is empty
func (data GenesisState) IsEmpty() bool {
	return data.Equal(GenesisState{})
}

// ValidateGenesis performs basic validation of genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	// iterate over assets and verify them
	for _, asset := range data.AssetParams.Assets {
		if asset.AssetCode == "" {
			return fmt.Errorf("invalid asset: %s. missing asset code", asset.String())
		}
	}

	// iterate over oracles and verify them
	for _, oracle := range data.OracleParams.Oracles {
		if oracle.OracleAddress == "" {
			return fmt.Errorf("invalid oracle: %s. missing oracle address", oracle.String())
		}
	}

	return nil
}
