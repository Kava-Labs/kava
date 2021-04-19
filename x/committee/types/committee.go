package types

import (
	"fmt"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const MaxCommitteeDescriptionLength int = 512

// ------------------------------------------
//				Committees
// ------------------------------------------

type VotingOption int

const (
	MembersOnly VotingOption = iota // 0
	CoinVoting  VotingOption = iota // 1
)

type TallyOption int

const (
	FirstPastThePost VotingOption = iota // 0
	Deadline         VotingOption = iota // 1
)

// A Committee is a collection of addresses that are allowed to vote and enact any governance proposal that passes their permissions.
type Committee struct {
	ID               uint64           `json:"id" yaml:"id"`
	Description      string           `json:"description" yaml:"description"`
	Members          []sdk.AccAddress `json:"members" yaml:"members"` // Only used in MembersOnly committees
	Permissions      []Permission     `json:"permissions" yaml:"permissions"`
	VoteThreshold    sdk.Dec          `json:"vote_threshold" yaml:"vote_threshold"`       // Smallest percentage of members that must vote for a proposal to pass. Only used in MembersOnly committees.
	ProposalDuration time.Duration    `json:"proposal_duration" yaml:"proposal_duration"` // The length of time a proposal remains active for. Proposals will close earlier if they get enough votes.
	VotingOption     VotingOption     `json:"voting_option" yaml:"voting_option"`
	Quorum           sdk.Dec          `json:"quorum" yaml:"quorum"`           // Only used in CoinVoting committees
	TallyDenom       string           `json:"tally_denom" yaml:"tally_denom"` // Only used in CoinVoting committees
	TallyOption      TallyOption      `json:"tally_option" yaml:"tally_option"`
}

func NewCommittee(id uint64, description string, members []sdk.AccAddress, permissions []Permission,
	threshold sdk.Dec, duration time.Duration, votingOption VotingOption, quorum sdk.Dec, tallyOption TallyOption) Committee {
	return Committee{
		ID:               id,
		Description:      description,
		Members:          members,
		Permissions:      permissions,
		VoteThreshold:    threshold,
		ProposalDuration: duration,
		VotingOption:     votingOption,
		Quorum:           quorum,
		TallyOption:      tallyOption,
	}
}

func (c Committee) HasMember(addr sdk.AccAddress) bool {
	if c.VotingOption == CoinVoting {
		return false
	}
	for _, m := range c.Members {
		if m.Equals(addr) {
			return true
		}
	}
	return false
}

// HasPermissionsFor returns whether the committee is authorized to enact a proposal.
// As long as one permission allows the proposal then it goes through. Its the OR of all permissions.
func (c Committee) HasPermissionsFor(ctx sdk.Context, appCdc *codec.Codec, pk ParamKeeper, proposal PubProposal) bool {
	for _, p := range c.Permissions {
		if p.Allows(ctx, appCdc, pk, proposal) {
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
			return fmt.Errorf("committe cannot have duplicate members, %s", m)
		}
		// check for valid addresses
		if m.Empty() {
			return fmt.Errorf("committee cannot have empty member address")
		}
		addressMap[m.String()] = true
	}

	if len(c.Description) > MaxCommitteeDescriptionLength {
		return fmt.Errorf("description length %d longer than max allowed %d", len(c.Description), MaxCommitteeDescriptionLength)
	}

	for _, p := range c.Permissions {
		if p == nil {
			return fmt.Errorf("committee cannot have a nil permission")
		}
	}

	if c.ProposalDuration < 0 {
		return fmt.Errorf("invalid proposal duration: %s", c.ProposalDuration)
	}

	switch c.VotingOption {
	case MembersOnly:
		if len(c.Members) == 0 {
			return fmt.Errorf("committee cannot have zero members")
		}

		// threshold must be in the range (0,1]
		if c.VoteThreshold.IsNil() || c.VoteThreshold.LTE(sdk.ZeroDec()) || c.VoteThreshold.GT(sdk.NewDec(1)) {
			return fmt.Errorf("invalid threshold: %s", c.VoteThreshold)
		}
	case CoinVoting:
		err := sdk.ValidateDenom(c.TallyDenom)
		if err != nil {
			return err
		}

		if len(c.Members) > 0 {
			return fmt.Errorf("invalid members")
		}

		if c.Quorum.IsNegative() {
			return fmt.Errorf("invalid quroum percentage: %s", c.Quorum)
		}
	}

	return nil
}

// ------------------------------------------
//				Proposals
// ------------------------------------------

// PubProposal is the interface that all proposals must fulfill to be submitted to a committee.
// Proposal types can be created external to this module. For example a ParamChangeProposal, or CommunityPoolSpendProposal.
// It is pinned to the equivalent type in the gov module to create compatibility between proposal types.
type PubProposal govtypes.Content

// Proposal is an internal record of a governance proposal submitted to a committee.
type Proposal struct {
	PubProposal `json:"pub_proposal" yaml:"pub_proposal"`
	ID          uint64    `json:"id" yaml:"id"`
	CommitteeID uint64    `json:"committee_id" yaml:"committee_id"`
	Deadline    time.Time `json:"deadline" yaml:"deadline"`
}

func NewProposal(pubProposal PubProposal, id uint64, committeeID uint64, deadline time.Time) Proposal {
	return Proposal{
		PubProposal: pubProposal,
		ID:          id,
		CommitteeID: committeeID,
		Deadline:    deadline,
	}
}

// HasExpiredBy calculates if the proposal will have expired by a certain time.
// All votes must be cast before deadline, those cast at time == deadline are not valid
func (p Proposal) HasExpiredBy(time time.Time) bool {
	return !time.Before(p.Deadline)
}

// String implements the fmt.Stringer interface, and importantly overrides the String methods inherited from the embedded PubProposal type.
func (p Proposal) String() string {
	bz, _ := yaml.Marshal(p)
	return string(bz)
}

// ------------------------------------------
//				Votes
// ------------------------------------------

type Vote struct {
	ProposalID uint64         `json:"proposal_id" yaml:"proposal_id"`
	Voter      sdk.AccAddress `json:"voter" yaml:"voter"`
}

func NewVote(proposalID uint64, voter sdk.AccAddress) Vote {
	return Vote{
		ProposalID: proposalID,
		Voter:      voter,
	}
}

func (v Vote) Validate() error {
	if v.Voter.Empty() {
		return fmt.Errorf("voter address cannot be empty")
	}
	return nil
}
