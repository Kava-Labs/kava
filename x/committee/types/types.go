package types

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
)

// TODO move these into params
var (
	VoteThreshold       sdk.Dec       = sdk.MustNewDecFromStr("0.75")
	MaxProposalDuration time.Duration = time.Hour * 24 * 7
)

// -------- Committees --------

// A Committee is a collection of addresses that are allowed to vote and enact any governance proposal that passes their permissions.
type Committee struct {
	ID uint64 `json:"id" yaml:"id"`
	//Description string           `json:"description" yaml:"description"`
	Members     []sdk.AccAddress `json:"members" yaml:"members"`
	Permissions []Permission     `json:"permissions" yaml:"permissions"`
	// VoteThreshold       sdk.Dec          `json:"vote_threshold" yaml:"vote_threshold"`
	// MaxProposalDuration time.Duration    `json:"max_proposal_duration" yaml:"max_proposal_duration"`
}

func NewCommittee(id uint64, members []sdk.AccAddress, permissions []Permission) Committee {
	return Committee{
		ID:          id,
		Members:     members,
		Permissions: permissions,
	}
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

func (c Committee) Validate() error {

	addressMap := make(map[string]bool, len(c.Members))
	for _, m := range c.Members {
		// check there are no duplicate members
		if _, ok := addressMap[m.String()]; ok {
			return fmt.Errorf("duplicate member found in committee, %s", m)
		}
		// check for valid addresses
		if m.Empty() {
			return fmt.Errorf("committee %d invalid: found empty member address", c.ID)
		}
		addressMap[m.String()] = true

	}

	if len(c.Members) == 0 {
		return fmt.Errorf("committee %d invalid: cannot have zero members", c.ID)
	}
	return nil
}

// Permission is anything with a method that validates whether a proposal is allowed by it or not.
type Permission interface {
	Allows(PubProposal) bool
}

// -------- Proposals --------

// PubProposal is an interface that all gov proposals defined in other modules must satisfy.
type PubProposal = gov.Content // TODO find a better name

type Proposal struct {
	PubProposal `json:"pub_proposal" yaml:"pub_proposal"`
	ID          uint64    `json:"id" yaml:"id"`
	CommitteeID uint64    `json:"committee_id" yaml:"committee_id"`
	Deadline    time.Time `json:"deadline" yaml:"deadline"`
}

// HasExpiredBy calculates if the proposal will have expired by a certain time.
// All votes must be cast before deadline, those cast at time == deadline are not valid
func (p Proposal) HasExpiredBy(time time.Time) bool {
	return !time.Before(p.Deadline)
}

// String implements the fmt.Stringer interface, and importantly overrides the String methods inherited from the embedded PubProposal type.
func (p Proposal) String() string {
	return strings.TrimSpace(fmt.Sprintf(`Proposal:
	PubProposal:
%s
	ID:           %d
	Committee ID: %d
	Deadline:     %s`,
		p.PubProposal,
		p.ID,
		p.CommitteeID,
		p.Deadline,
	))
}

type Vote struct {
	ProposalID uint64         `json:"proposal_id" yaml:"proposal_id"`
	Voter      sdk.AccAddress `json:"voter" yaml:"voter"`
	// Option     byte // TODO for now don't need more than just a yes as options
}
