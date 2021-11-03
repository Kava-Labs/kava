package types

import (
	"fmt"

	types "github.com/cosmos/cosmos-sdk/codec/types"
	proto "github.com/gogo/protobuf/proto"
)

// DefaultNextProposalID is the starting poiint for proposal IDs.
const DefaultNextProposalID uint64 = 1

// NewGenesisState returns a new genesis state object for the module.
func NewGenesisState(nextProposalID uint64, committees []Committee, proposals []Proposal, votes []Vote) *GenesisState {
	packedCommittees, err := PackCommittees(committees)
	if err != nil {
		panic(err)
	}
	return &GenesisState{
		NextProposalId: nextProposalID,
		Committees:     packedCommittees,
		Proposals:      proposals,
		Votes:          votes,
	}
}

// DefaultGenesisState returns the default genesis state for the module.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(
		DefaultNextProposalID,
		[]Committee{},
		[]Proposal{},
		[]Vote{},
	)
}

// Validate performs basic validation of genesis data.
func (gs GenesisState) Validate() error {
	// validate committees
	committeeMap := make(map[uint64]bool, len(gs.Committees))
	committees, err := UnpackCommittees(gs.Committees)
	if err != nil {
		return err
	}
	for _, com := range committees {
		// check there are no duplicate IDs
		if _, ok := committeeMap[com.GetId()]; ok {
			return fmt.Errorf("duplicate committee ID found in genesis state; id: %d", com.GetId())
		}
		committeeMap[com.GetId()] = true

		// validate committee
		if err := com.Validate(); err != nil {
			return err
		}
	}

	// validate proposals
	proposalMap := make(map[uint64]bool, len(gs.Proposals))
	for _, p := range gs.Proposals {
		// check there are no duplicate IDs
		if _, ok := proposalMap[p.Id]; ok {
			return fmt.Errorf("duplicate proposal ID found in genesis state; id: %d", p.ID)
		}
		proposalMap[p.Id] = true

		// validate next proposal ID
		if p.Id >= gs.NextProposalId {
			return fmt.Errorf("NextProposalID is not greater than all proposal IDs; id: %d", p.ID)
		}

		// check committee exists
		if !committeeMap[p.CommitteeId] {
			return fmt.Errorf("proposal refers to non existent committee; proposal: %+v", p)
		}

		// validate pubProposal
		if err := p.GetPubProposal().ValidateBasic(); err != nil {
			return fmt.Errorf("proposal %d invalid: %w", p.Id, err)
		}
	}

	// validate votes
	for _, v := range gs.Votes {
		// validate committee
		if err := v.Validate(); err != nil {
			return err
		}

		// check proposal exists
		if !proposalMap[v.ProposalId] {
			return fmt.Errorf("vote refers to non existent proposal; vote: %+v", v)
		}
	}
	return nil
}

// PackCommittees converts a committee slice to Any slice
func PackCommittees(committees []Committee) ([]*types.Any, error) {
	committeesAny := make([]*types.Any, len(committees))
	for i, committee := range committees {
		any, err := PackCommittee(committee)
		if err != nil {
			return nil, err
		}
		committeesAny[i] = any
	}

	return committeesAny, nil
}

// PackCommittee converts a committee to Any
func PackCommittee(committee Committee) (*types.Any, error) {
	msg, ok := committee.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("cannot proto marshal %T", committee)
	}
	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}
	return any, nil
}

// UnpackCommittees converts Any slice to Committee slice
func UnpackCommittees(committeesAny []*types.Any) ([]Committee, error) {
	committees := make([]Committee, len(committeesAny))
	for i, any := range committeesAny {
		committee, err := UnpackCommittee(any)
		if err != nil {
			return nil, err
		}
		committees[i] = committee
	}

	return committees, nil
}

// UnpackCommittee converts Any to Committee
func UnpackCommittee(committeeAny *types.Any) (Committee, error) {
	committee, ok := committeeAny.GetCachedValue().(Committee)
	if !ok {
		return nil, fmt.Errorf("unexpected committee when unpacking")
	}
	return committee, nil
}
