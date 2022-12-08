package types

import (
	time "time"
)

var (
	// DefaultPreviousBlockTime represents a time that is unset -- no previous block has occured
	DefaultPreviousBlockTime = time.Time{}
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

// Validate validates the provided genesis state to ensure the
// expected invariants holds.
func (gs GenesisState) Validate() error {
	return gs.Params.Validate()
}
