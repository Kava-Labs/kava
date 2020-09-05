package types

import (
	"bytes"
	"fmt"
	"time"
)

// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
	Params              Params    `json:"params" yaml:"params"`
	PreviousBlockTime   time.Time `json:"previous_block_time" yaml:"previous_block_time"`
	InitialDistribution bool      `json:"initial_distribution" yaml:"initial_distribution"`
}

// NewGenesisState returns a new genesis state
func NewGenesisState(params Params, previousBlockTime time.Time, initialDistribution bool) GenesisState {
	return GenesisState{
		Params:              params,
		PreviousBlockTime:   previousBlockTime,
		InitialDistribution: initialDistribution,
	}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:              DefaultParams(),
		PreviousBlockTime:   DefaultPreviousBlockTime,
		InitialDistribution: true,
	}
}

// Validate performs basic validation of genesis data returning an
// error for any failed validation criteria.
func (gs GenesisState) Validate() error {

	if err := gs.Params.Validate(); err != nil {
		return err
	}
	if gs.PreviousBlockTime.Equal(time.Time{}) {
		return fmt.Errorf("previous block time not set")
	}
	return nil
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
