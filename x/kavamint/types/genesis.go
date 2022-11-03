package types

import (
	fmt "fmt"
	time "time"
)

// NewGenesisState creates a new GenesisState object
func NewGenesisState(params Params, previousBlockTime time.Time) *GenesisState {
	return &GenesisState{
		Params:            params,
		PreviousBlockTime: previousBlockTime,
	}
}

// DefaultGenesisState creates a default GenesisState object
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:            DefaultParams(),
		PreviousBlockTime: DefaultPreviousBlockTime,
	}
}

// ValidateGenesis validates the provided genesis state to ensure the
// expected invariants holds.
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}
	if data.PreviousBlockTime.IsZero() {
		return fmt.Errorf("previous block time not set")
	}
	return nil
}
