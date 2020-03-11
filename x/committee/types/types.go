package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
)

// -------- Committees --------

var VoteThreshold sdk.Dec = sdk.MustNewDecFromStr("0.75")

// A Committee is a collection of addresses that are allowed to vote and enact any governance proposal that passes their permissions.
type Committee struct {
	ID          uint64 // TODO or a name?
	Members     []sdk.AccAddress
	Permissions []Permission
}

func (c Committee) HasMember(addr sdk.AccAddress) bool {
	for _, m := range c.Members {
		if m.Equals(addr) {
			return true
		}
	}
	return false
}

// HasPermissionsFor returns whether the committee is authorized to enact a proposal.
// As long as one permission allows the proposal then it goes through. Its the OR of all permissions.
func (c Committee) HasPermissionsFor(proposal PubProposal) bool {
	for _, p := range c.Permissions {
		if p.Allows(proposal) {
			return true
		}
	}
	return false
}

// Permission is anything with a method that validates whether a proposal is allowed by it or not.
type Permission interface {
	Allows(PubProposal) bool
}

// -------- Proposals --------

// PubProposal is an interface that all gov proposals defined in other modules must satisfy.
type PubProposal = gov.Content // TODO find a better name

type Proposal struct {
	PubProposal
	ID          uint64
	CommitteeID uint64
}

type Vote struct {
	ProposalID uint64
	Voter      sdk.AccAddress
	// Option     byte // TODO for now don't need more than just a yes as options
}
