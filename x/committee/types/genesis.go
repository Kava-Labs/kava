package types

import (
	"bytes"
)

// DefaultNextProposalID is the starting poiint for proposal IDs.
const DefaultNextProposalID uint64 = 1

// GenesisState is state that must be provided at chain genesis.
type GenesisState struct {
	NextProposalID uint64
	Votes          []Vote
	Proposals      []Proposal
	Committees     []Committee
}

// NewGenesisState returns a new genesis state object for the module.
func NewGenesisState(nextProposalID uint64, votes []Vote, proposals []Proposal, committees []Committee) GenesisState {
	return GenesisState{
		NextProposalID: nextProposalID,
		Votes:          votes,
		Proposals:      proposals,
		Committees:     committees,
	}
}

// DefaultGenesisState returns the default genesis state for the module.
func DefaultGenesisState() GenesisState {
	return NewGenesisState(
		DefaultNextProposalID,
		[]Vote{},
		[]Proposal{},
		[]Committee{},
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

// Validate performs basic validation of genesis data.
func (gs GenesisState) Validate() error { return nil }
