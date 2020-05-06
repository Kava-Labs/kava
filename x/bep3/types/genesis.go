package types

import (
	"bytes"
	"encoding/hex"
	"fmt"
)

// GenesisState - all bep3 state that must be provided at genesis
type GenesisState struct {
	Params        Params        `json:"params" yaml:"params"`
	AtomicSwaps   AtomicSwaps   `json:"atomic_swaps" yaml:"atomic_swaps"`
	AssetSupplies AssetSupplies `json:"assets_supplies" yaml:"assets_supplies"`
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(params Params, swaps AtomicSwaps, supplies AssetSupplies) GenesisState {
	return GenesisState{
		Params:        params,
		AtomicSwaps:   swaps,
		AssetSupplies: supplies,
	}
}

// DefaultGenesisState - default GenesisState used by Cosmos Hub
func DefaultGenesisState() GenesisState {
	return NewGenesisState(
		DefaultParams(),
		AtomicSwaps{},
		AssetSupplies{},
	)
}

// Equal checks whether two GenesisState structs are equivalent.
func (gs GenesisState) Equal(gs2 GenesisState) bool {
	b1 := ModuleCdc.MustMarshalBinaryBare(gs)
	b2 := ModuleCdc.MustMarshalBinaryBare(gs2)
	return bytes.Equal(b1, b2)
}

// IsEmpty returns true if a GenesisState is empty.
func (gs GenesisState) IsEmpty() bool {
	return gs.Equal(GenesisState{})
}

// Validate validates genesis inputs. It returns error if validation of any input fails.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	denoms := map[string]bool{}
	for _, asset := range gs.AssetSupplies {
		if denoms[asset.Denom] {
			return fmt.Errorf("found duplicate asset denom %s", asset.Denom)
		}

		if err := asset.Validate(); err != nil {
			return err
		}

		denoms[asset.Denom] = true
	}
	ids := map[string]bool{}
	for _, a := range gs.AtomicSwaps {
		if ids[hex.EncodeToString(a.GetSwapID())] {
			return fmt.Errorf("found duplicate atomic swap ID %s", hex.EncodeToString(a.GetSwapID()))
		}
		ids[hex.EncodeToString(a.GetSwapID())] = true
	}
	return nil
}
