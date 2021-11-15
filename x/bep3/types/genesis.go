package types

import (
	"encoding/hex"
	"fmt"
	"time"
)

// NewGenesisState creates a new GenesisState object
func NewGenesisState(params Params, swaps AtomicSwaps, supplies AssetSupplies, previousBlockTime time.Time) GenesisState {
	return GenesisState{
		Params:            params,
		AtomicSwaps:       swaps,
		Supplies:          supplies,
		PreviousBlockTime: previousBlockTime,
	}
}

// DefaultGenesisState - default GenesisState used by Cosmos Hub
func DefaultGenesisState() GenesisState {
	return NewGenesisState(
		DefaultParams(),
		AtomicSwaps{},
		AssetSupplies{},
		DefaultPreviousBlockTime,
	)
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
