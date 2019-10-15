package types

import (
	"bytes"
	"fmt"
	"time"

	tmtime "github.com/tendermint/tendermint/types/time"
)

// GenesisState - all auth state that must be provided at genesis
type GenesisState struct {
	PreviousBlockTime time.Time `json:"previous_block_time" yaml:"previous_block_time"`
}

// NewGenesisState - Create a new genesis state
func NewGenesisState(prevBlockTime time.Time) GenesisState {
	return GenesisState{
		PreviousBlockTime: prevBlockTime,
	}
}

// DefaultGenesisState - Return a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState(tmtime.Canonical(time.Unix(0, 0)))
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

// ValidateGenesis returns nil because accounts are validated by auth
func ValidateGenesis(data GenesisState) error {
	if data.PreviousBlockTime.Unix() < 0 {
		return fmt.Errorf("Previous block time should be positive, is set to %v", data.PreviousBlockTime.Unix())
	}
	return nil
}
