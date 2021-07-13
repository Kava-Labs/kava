package types

import "bytes"

var (
	DefaultPoolRecords  = PoolRecords{}
	DefaultShareRecords = ShareRecords{}
)

// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
	Params       Params `json:"params" yaml:"params"`
	PoolRecords  `json:"pool_records" yaml:"pool_records"`
	ShareRecords `json:"share_records" yaml:"share_records"`
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(params Params, poolRecords PoolRecords, shareRecords ShareRecords) GenesisState {
	return GenesisState{
		Params:       params,
		PoolRecords:  poolRecords,
		ShareRecords: shareRecords,
	}
}

// Validate validates the module's genesis state
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	if err := gs.PoolRecords.Validate(); err != nil {
		return err
	}
	return gs.ShareRecords.Validate()
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState(
		DefaultParams(),
		DefaultPoolRecords,
		DefaultShareRecords,
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
