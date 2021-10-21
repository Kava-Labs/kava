package types

import (
	"bytes"
	"fmt"
)

// DefaultNextProposalID is the starting poiint for proposal IDs.
const DefaultNextProposalID uint64 = 1

// GenesisState is state that must be provided at chain genesis.
type GenesisState struct {
	NextProposalID uint64     `json:"next_proposal_id" yaml:"next_proposal_id"`
	Committees     Committees `json:"committees" yaml:"committees"`
	Proposals      []Proposal `json:"proposals" yaml:"proposals"`
	Votes          []Vote     `json:"votes" yaml:"votes"`
}

// NewGenesisState returns a new genesis state object for the module.
func NewGenesisState(nextProposalID uint64, committees Committees, proposals []Proposal, votes []Vote) GenesisState {
	return GenesisState{
		NextProposalID: nextProposalID,
		Committees:     committees,
		Proposals:      proposals,
		Votes:          votes,
	}
}

// DefaultGenesisState returns the default genesis state for the module.
func DefaultGenesisState() GenesisState {
	return NewGenesisState(
		DefaultNextProposalID,
		Committees{},
		[]Proposal{},
		[]Vote{},
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
func (gs GenesisState) Validate() error {
	// validate committees
	committeeMap := make(map[uint64]bool, len(gs.Committees))
	for _, com := range gs.Committees {
		// check there are no duplicate IDs
		if _, ok := committeeMap[com.GetID()]; ok {
			return fmt.Errorf("duplicate committee ID found in genesis state; id: %d", com.GetID())
		}
		committeeMap[com.GetID()] = true

		// validate committee
		if err := com.Validate(); err != nil {
			return err
		}
	}

	// validate proposals
	proposalMap := make(map[uint64]bool, len(gs.Proposals))
	for _, p := range gs.Proposals {
		// check there are no duplicate IDs
		if _, ok := proposalMap[p.ID]; ok {
			return fmt.Errorf("duplicate proposal ID found in genesis state; id: %d", p.ID)
		}
		proposalMap[p.ID] = true

		// validate next proposal ID
		if p.ID >= gs.NextProposalID {
			return fmt.Errorf("NextProposalID is not greater than all proposal IDs; id: %d", p.ID)
		}

		// check committee exists
		if !committeeMap[p.CommitteeID] {
			return fmt.Errorf("proposal refers to non existent committee; proposal: %+v", p)
		}

		// validate pubProposal
		if err := p.PubProposal.ValidateBasic(); err != nil {
			return fmt.Errorf("proposal %d invalid: %w", p.ID, err)
		}
	}

	// validate votes
	for _, v := range gs.Votes {
		// validate committee
		if err := v.Validate(); err != nil {
			return err
		}

		// check proposal exists
		if !proposalMap[v.ProposalID] {
			return fmt.Errorf("vote refers to non existent proposal; vote: %+v", v)
		}
	}
	return nil
}
