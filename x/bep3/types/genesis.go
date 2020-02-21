package types

import (
	"bytes"
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type GenesisAtomicSwap interface {
	Swap
	GetModuleAccountCoins() sdk.Coins
	Validate() error
}

// GenesisAtomicSwaps is a slice of genesis atomic swaps.
type GenesisAtomicSwaps []GenesisAtomicSwap

// GenesisState - all bep3 state that must be provided at genesis
type GenesisState struct {
	Params      Params             `json:"params" yaml:"params"`
	AtomicSwaps GenesisAtomicSwaps `json:"atomic_swaps" yaml:"atomic_swaps"`
	Assets      []sdk.Coin         `json:"assets" yaml:"assets"`
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(params Params, genAtomicSwaps GenesisAtomicSwaps, assets []sdk.Coin) GenesisState {
	return GenesisState{
		Params:      params,
		AtomicSwaps: genAtomicSwaps,
		Assets:      assets,
	}
}

// DefaultGenesisState - default GenesisState used by Cosmos Hub
func DefaultGenesisState() GenesisState {
	return NewGenesisState(
		DefaultParams(),
		GenesisAtomicSwaps{},
		[]sdk.Coin{},
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
	for _, a := range gs.AtomicSwaps {

		if err := a.Validate(); err != nil {
			return fmt.Errorf("found invalid atomic swap: %w", err)
		}
		if ids[hex.EncodeToString(a.GetSwapID())] {
			return fmt.Errorf("found duplicate atomic swap ID (%s)", hex.EncodeToString(a.GetSwapID()))
		}
		ids[hex.EncodeToString(a.GetSwapID())] = true
	}
	return nil
}
