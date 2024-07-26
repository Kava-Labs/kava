package types

import (
	"fmt"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	proto "github.com/cosmos/gogoproto/proto"
)

// DefaultNextProposalID is the starting point for proposal IDs.
const DefaultNextProposalID uint64 = 1

// NewGenesisState returns a new genesis state object for the module.
func NewGenesisState(nextProposalID uint64, committees []Committee, proposals Proposals, votes []Vote) *GenesisState {
	packedCommittees, err := PackCommittees(committees)
	if err != nil {
		panic(err)
	}
	return &GenesisState{
		NextProposalID: nextProposalID,
		Committees:     packedCommittees,
		Proposals:      proposals,
		Votes:          votes,
	}
}

// DefaultGenesisState returns the default genesis state for the module.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(
		DefaultNextProposalID,
		Committees{},
		Proposals{},
		[]Vote{},
	)
}

func (gs GenesisState) GetCommittees() Committees {
	committees, err := UnpackCommittees(gs.Committees)
	if err != nil {
		panic(err)
	}
	return committees
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (data GenesisState) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	for _, any := range data.Committees {
		var committee Committee
		if err := unpacker.UnpackAny(any, &committee); err != nil {
			return err
		}
	}
	for _, p := range data.Proposals {
		err := p.UnpackInterfaces(unpacker)
		if err != nil {
			return err
		}
	}
	return nil
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
			return fmt.Errorf("proposal refers to non existent committee; committee id: %d", p.CommitteeID)
		}

		// validate pubProposal
		if err := p.ValidateBasic(); err != nil {
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

// PackCommittees converts a committee slice to Any slice
func PackCommittees(committees []Committee) ([]*cdctypes.Any, error) {
	committeesAny := make([]*cdctypes.Any, len(committees))
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
func PackCommittee(committee Committee) (*cdctypes.Any, error) {
	msg, ok := committee.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("cannot proto marshal %T", committee)
	}
	any, err := cdctypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}
	return any, nil
}

// UnpackCommittees converts Any slice to Committee slice
func UnpackCommittees(committeesAny []*cdctypes.Any) (Committees, error) {
	committees := make(Committees, len(committeesAny))
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
func UnpackCommittee(committeeAny *cdctypes.Any) (Committee, error) {
	committee, ok := committeeAny.GetCachedValue().(Committee)
	if !ok {
		return nil, fmt.Errorf("unexpected committee when unpacking")
	}
	return committee, nil
}
