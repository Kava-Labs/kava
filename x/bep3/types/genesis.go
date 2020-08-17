package types

import (
	"bytes"
	"encoding/hex"
	"fmt"
)

// GenesisState - all bep3 state that must be provided at genesis
type GenesisState struct {
	Params      Params        `json:"params" yaml:"params"`
	AtomicSwaps AtomicSwaps   `json:"atomic_swaps" yaml:"atomic_swaps"`
	Supplies    AssetSupplies `json:"supplies" yaml:"supplies"`
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(params Params, swaps AtomicSwaps, supplies AssetSupplies) GenesisState {
	return GenesisState{
		Params:      params,
		AtomicSwaps: swaps,
		Supplies:    supplies,
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

	ids := map[string]bool{}
	for _, swap := range gs.AtomicSwaps {
		if ids[hex.EncodeToString(swap.GetSwapID())] {
			return fmt.Errorf("found duplicate atomic swap ID %s", hex.EncodeToString(swap.GetSwapID()))
		}

		if err := swap.Validate(); err != nil {
			return err
		}

		ids[hex.EncodeToString(swap.GetSwapID())] = true
	}

	supplyDenoms := map[string]bool{}
	for _, supply := range gs.Supplies {
		if err := supply.Validate(); err != nil {
			return err
		}
		if supplyDenoms[supply.GetDenom()] {
			return fmt.Errorf("found duplicate denom in asset supplies %s", supply.GetDenom())
		}
		supplyDenoms[supply.GetDenom()] = true
	}
	return nil
}
