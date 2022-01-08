package v0_15

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	v036gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v036"
)

const MaxCommitteeDescriptionLength int = 512

type TallyOption uint64

const (
	NullTallyOption  TallyOption = iota
	FirstPastThePost TallyOption = iota // Votes are tallied each block and the proposal passes as soon as the vote threshold is reached
	Deadline         TallyOption = iota // Votes are tallied exactly once, when the deadline time is reached
)

const (
	BaseCommitteeType   = "kava/BaseCommittee"
	MemberCommitteeType = "kava/MemberCommittee" // Committee is composed of member addresses that vote to enact proposals within their permissions
	TokenCommitteeType  = "kava/TokenCommittee"  // Committee is composed of token holders with voting power determined by total token balance
	BondDenom           = "ukava"
)

// TallyOptionFromString returns a TallyOption from a string. It returns an error
// if the string is invalid.
func TallyOptionFromString(str string) (TallyOption, error) {
	switch strings.ToLower(str) {
	case "firstpastthepost", "fptp":
		return FirstPastThePost, nil

	case "deadline", "d":
		return Deadline, nil

	default:
		return TallyOption(0xff), fmt.Errorf("'%s' is not a valid tally option", str)
	}
}

// Marshal needed for protobuf compatibility.
func (t TallyOption) Marshal() ([]byte, error) {
	return []byte{byte(t)}, nil
}

// Unmarshal needed for protobuf compatibility.
func (t *TallyOption) Unmarshal(data []byte) error {
	*t = TallyOption(data[0])
	return nil
}

// Marshals to JSON using string.
func (t TallyOption) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// UnmarshalJSON decodes from JSON assuming Bech32 encoding.
func (t *TallyOption) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	bz2, err := TallyOptionFromString(s)
	if err != nil {
		return err
	}

	*t = bz2
	return nil
}

// String implements the Stringer interface.
func (t TallyOption) String() string {
	switch t {
	case FirstPastThePost:
		return "FirstPastThePost"
	case Deadline:
		return "Deadline"
	default:
		return ""
	}
}

// Committee is an interface for handling common actions on committees
type Committee interface {
	GetID() uint64
	GetType() string
	GetDescription() string

	GetMembers() []sdk.AccAddress
	SetMembers([]sdk.AccAddress) BaseCommittee
	HasMember(addr sdk.AccAddress) bool

	GetPermissions() []Permission
	SetPermissions([]Permission) Committee

	GetProposalDuration() time.Duration
	SetProposalDuration(time.Duration) BaseCommittee

	GetVoteThreshold() sdk.Dec
	SetVoteThreshold(sdk.Dec) BaseCommittee

	GetTallyOption() TallyOption
	Validate() error
}

var (
	_ Committee = MemberCommittee{}
	_ Committee = TokenCommittee{}
)

// Committees is a slice of committees
type Committees []Committee

// BaseCommittee is a common type shared by all Committees
type BaseCommittee struct {
	ID               uint64           `json:"id" yaml:"id"`
	Description      string           `json:"description" yaml:"description"`
	Members          []sdk.AccAddress `json:"members" yaml:"members"`
	Permissions      []Permission     `json:"permissions" yaml:"permissions"`
	VoteThreshold    sdk.Dec          `json:"vote_threshold" yaml:"vote_threshold"`       // Smallest percentage that must vote for a proposal to pass
	ProposalDuration time.Duration    `json:"proposal_duration" yaml:"proposal_duration"` // The length of time a proposal remains active for. Proposals will close earlier if they get enough votes.
	TallyOption      TallyOption      `json:"tally_option" yaml:"tally_option"`
}

// GetType is a getter for committee type
func (c BaseCommittee) GetType() string { return BaseCommitteeType }

// GetID is a getter for committee ID
func (c BaseCommittee) GetID() uint64 { return c.ID }

// GetDescription is a getter for committee description
func (c BaseCommittee) GetDescription() string { return c.Description }

// GetMembers is a getter for committee members
func (c BaseCommittee) GetMembers() []sdk.AccAddress { return c.Members }

// SetMembers is a setter for committee members
func (c BaseCommittee) SetMembers(members []sdk.AccAddress) BaseCommittee {
	c.Members = members
	return c
}

// HasMember returns if a committee contains a given member address
func (c BaseCommittee) HasMember(addr sdk.AccAddress) bool {
	for _, m := range c.Members {
		if m.Equals(addr) {
			return true
		}
	}
	return false
}

// GetPermissions is a getter for committee permissions
func (c BaseCommittee) GetPermissions() []Permission { return c.Permissions }

// SetPermissions is a setter for committee permissions
func (c BaseCommittee) SetPermissions(permissions []Permission) BaseCommittee {
	c.Permissions = permissions
	return c
}

// GetVoteThreshold is a getter for committee VoteThreshold
func (c BaseCommittee) GetVoteThreshold() sdk.Dec { return c.VoteThreshold }

// SetVoteThreshold is a setter for committee VoteThreshold
func (c BaseCommittee) SetVoteThreshold(voteThreshold sdk.Dec) BaseCommittee {
	c.VoteThreshold = voteThreshold
	return c
}

// GetProposalDuration is a getter for committee ProposalDuration
func (c BaseCommittee) GetProposalDuration() time.Duration { return c.ProposalDuration }

// SetProposalDuration is a setter for committee ProposalDuration
func (c BaseCommittee) SetProposalDuration(proposalDuration time.Duration) BaseCommittee {
	c.ProposalDuration = proposalDuration
	return c
}

// GetTallyOption is a getter for committee TallyOption
func (c BaseCommittee) GetTallyOption() TallyOption { return c.TallyOption }

// Validate validates BaseCommittee fields
func (c BaseCommittee) Validate() error {
	if len(c.Description) > MaxCommitteeDescriptionLength {
		return fmt.Errorf("description length %d longer than max allowed %d", len(c.Description), MaxCommitteeDescriptionLength)
	}

	if len(c.Members) <= 0 {
		return fmt.Errorf("committee must have members")
	}

	addressMap := make(map[string]bool, len(c.Members))
	for _, m := range c.Members {
		// check there are no duplicate members
		if _, ok := addressMap[m.String()]; ok {
			return fmt.Errorf("committee cannot have duplicate members, %s", m)
		}
		// check for valid addresses
		if m.Empty() {
			return fmt.Errorf("committee cannot have empty member address")
		}
		addressMap[m.String()] = true
	}

	for _, p := range c.Permissions {
		if p == nil {
			return fmt.Errorf("committee cannot have a nil permission")
		}
	}

	if c.ProposalDuration < 0 {
		return fmt.Errorf("invalid proposal duration: %s", c.ProposalDuration)
	}

	// threshold must be in the range [0, 1]
	if c.VoteThreshold.IsNil() || c.VoteThreshold.LTE(sdk.ZeroDec()) || c.VoteThreshold.GT(sdk.NewDec(1)) {
		return fmt.Errorf("invalid threshold: %s", c.VoteThreshold)
	}

	if c.TallyOption <= 0 || c.TallyOption > 2 {
		return fmt.Errorf("invalid tally option: %d", c.TallyOption)
	}

	return nil
}

// MemberCommittee is an alias of BaseCommittee
type MemberCommittee struct {
	BaseCommittee `json:"base_committee" yaml:"base_committee"`
}

// NewMemberCommittee instantiates a new instance of MemberCommittee
func NewMemberCommittee(id uint64, description string, members []sdk.AccAddress, permissions []Permission,
	threshold sdk.Dec, duration time.Duration, tallyOption TallyOption) MemberCommittee {
	return MemberCommittee{
		BaseCommittee: BaseCommittee{
			ID:               id,
			Description:      description,
			Members:          members,
			Permissions:      permissions,
			VoteThreshold:    threshold,
			ProposalDuration: duration,
			TallyOption:      tallyOption,
		},
	}
}

// GetType is a getter for committee type
func (c MemberCommittee) GetType() string { return MemberCommitteeType }

// SetPermissions is a setter for committee permissions
func (c MemberCommittee) SetPermissions(permissions []Permission) Committee {
	c.Permissions = permissions
	return c
}

// Validate validates the committee's fields
func (c MemberCommittee) Validate() error {
	return c.BaseCommittee.Validate()
}

// TokenCommittee supports voting on proposals by token holders
type TokenCommittee struct {
	BaseCommittee `json:"base_committee" yaml:"base_committee"`
	Quorum        sdk.Dec `json:"quorum" yaml:"quorum"`
	TallyDenom    string  `json:"tally_denom" yaml:"tally_denom"`
}

// NewTokenCommittee instantiates a new instance of TokenCommittee
func NewTokenCommittee(id uint64, description string, members []sdk.AccAddress, permissions []Permission,
	threshold sdk.Dec, duration time.Duration, tallyOption TallyOption, quorum sdk.Dec, tallyDenom string) TokenCommittee {
	return TokenCommittee{
		BaseCommittee: BaseCommittee{
			ID:               id,
			Description:      description,
			Members:          members,
			Permissions:      permissions,
			VoteThreshold:    threshold,
			ProposalDuration: duration,
			TallyOption:      tallyOption,
		},
		Quorum:     quorum,
		TallyDenom: tallyDenom,
	}
}

// GetType is a getter for committee type
func (c TokenCommittee) GetType() string { return TokenCommitteeType }

// GetQuorum returns the quorum of the committee
func (c TokenCommittee) GetQuorum() sdk.Dec { return c.Quorum }

// GetTallyDenom returns the tally denom of the committee
func (c TokenCommittee) GetTallyDenom() string { return c.TallyDenom }

// SetPermissions is a setter for committee permissions
func (c TokenCommittee) SetPermissions(permissions []Permission) Committee {
	c.Permissions = permissions
	return c
}

// Validate validates the committee's fields
func (c TokenCommittee) Validate() error {
	if c.TallyDenom == BondDenom {
		return fmt.Errorf("invalid tally denom: %s", c.TallyDenom)
	}

	err := sdk.ValidateDenom(c.TallyDenom)
	if err != nil {
		return err
	}

	if c.Quorum.IsNil() || c.Quorum.IsNegative() || c.Quorum.GT(sdk.NewDec(1)) {
		return fmt.Errorf("invalid quorum: %s", c.Quorum)
	}

	return c.BaseCommittee.Validate()
}

// ------------------------------------------
//				Proposals
// ------------------------------------------

// PubProposal is the interface that all proposals must fulfill to be submitted to a committee.
// Proposal types can be created external to this module. For example a ParamChangeProposal, or CommunityPoolSpendProposal.
// It is pinned to the equivalent type in the gov module to create compatibility between proposal types.
type PubProposal v036gov.Content

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

// ------------------------------------------
//				Votes
// ------------------------------------------

type Vote struct {
	ProposalID uint64         `json:"proposal_id" yaml:"proposal_id"`
	Voter      sdk.AccAddress `json:"voter" yaml:"voter"`
	VoteType   VoteType       `json:"vote_type" yaml:"vote_type"`
}

func NewVote(proposalID uint64, voter sdk.AccAddress, voteType VoteType) Vote {
	return Vote{
		ProposalID: proposalID,
		Voter:      voter,
		VoteType:   voteType,
	}
}

func (v Vote) Validate() error {
	if v.Voter.Empty() {
		return fmt.Errorf("voter address cannot be empty")
	}

	return v.VoteType.Validate()
}

type VoteType uint64

const (
	NullVoteType VoteType = iota // 0
	Yes          VoteType = iota // 1
	No           VoteType = iota // 2
	Abstain      VoteType = iota // 3
)

// VoteTypeFromString returns a VoteType from a string. It returns an error
// if the string is invalid.
func VoteTypeFromString(str string) (VoteType, error) {
	switch strings.ToLower(str) {
	case "yes", "y":
		return Yes, nil

	case "abstain", "a":
		return Abstain, nil

	case "no", "n":
		return No, nil

	default:
		return VoteType(0xff), fmt.Errorf("'%s' is not a valid vote type", str)
	}
}

// Marshal needed for protobuf compatibility.
func (vt VoteType) Marshal() ([]byte, error) {
	return []byte{byte(vt)}, nil
}

// Unmarshal needed for protobuf compatibility.
func (vt *VoteType) Unmarshal(data []byte) error {
	*vt = VoteType(data[0])
	return nil
}

// Marshals to JSON using string.
func (vt VoteType) MarshalJSON() ([]byte, error) {
	return json.Marshal(vt.String())
}

// UnmarshalJSON decodes from JSON assuming Bech32 encoding.
func (vt *VoteType) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	bz2, err := VoteTypeFromString(s)
	if err != nil {
		return err
	}

	*vt = bz2
	return nil
}

// String implements the Stringer interface.
func (vt VoteType) String() string {
	switch vt {
	case Yes:
		return "Yes"
	case Abstain:
		return "Abstain"
	case No:
		return "No"
	default:
		return ""
	}
}

func (vt VoteType) Validate() error {
	if vt <= 0 || vt > 3 {
		return fmt.Errorf("invalid vote type: %d", vt)
	}
	return nil
}

// CommitteeChangeProposal is a gov proposal for creating a new committee or modifying an existing one.
type CommitteeChangeProposal struct {
	Title        string    `json:"title" yaml:"title"`
	Description  string    `json:"description" yaml:"description"`
	NewCommittee Committee `json:"new_committee" yaml:"new_committee"`
}

// CommitteeDeleteProposal is a gov proposal for removing a committee.
type CommitteeDeleteProposal struct {
	Title       string `json:"title" yaml:"title"`
	Description string `json:"description" yaml:"description"`
	CommitteeID uint64 `json:"committee_id" yaml:"committee_id"`
}
