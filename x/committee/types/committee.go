package types

import (
	fmt "fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	proto "github.com/gogo/protobuf/proto"
	"gopkg.in/yaml.v2"
)

const MaxCommitteeDescriptionLength int = 512

const (
	BaseCommitteeType   = "kava/BaseCommittee"
	MemberCommitteeType = "kava/MemberCommittee" // Committee is composed of member addresses that vote to enact proposals within their permissions
	TokenCommitteeType  = "kava/TokenCommittee"  // Committee is composed of token holders with voting power determined by total token balance
	BondDenom           = "ukava"
)

// Committee is an interface for handling common actions on committees
type Committee interface {
	proto.Message

	GetId() uint64
	GetType() string
	GetDescription() string

	GetMembers() []sdk.AccAddress
	SetMembers([]sdk.AccAddress)
	HasMember(addr sdk.AccAddress) bool

	GetPermissions() []Permission
	SetPermissions([]Permission)
	HasPermissionsFor(ctx sdk.Context, appCdc *codec.Codec, pk ParamKeeper, proposal PubProposal) bool

	GetProposalDuration() time.Duration
	SetProposalDuration(time.Duration)

	GetVoteThreshold() sdk.Dec
	SetVoteThreshold(sdk.Dec)

	GetTallyOption() TallyOption
	Validate() error
}

var (
	_ Committee                          = &BaseCommittee{}
	_ codectypes.UnpackInterfacesMessage = &BaseCommittee{}
)

// GetType is a getter for committee type
func (c *BaseCommittee) GetType() string { return BaseCommitteeType }

// GetID is a getter for committee ID
func (c *BaseCommittee) GetId() uint64 { return c.Id }

// GetDescription is a getter for committee description
func (c *BaseCommittee) GetDescription() string { return c.Description }

// GetMembers is a getter for committee members
func (b BaseCommittee) GetMembers() []sdk.AccAddress {
	addresses := make([]sdk.AccAddress, len(b.Members))
	for _, m := range b.Members {
		address, err := sdk.AccAddressFromBech32(m)
		if err != nil {
			panic(err)
		}
		addresses = append(addresses, address)
	}
	return addresses
}

// SetMembers is a setter for committee members
func (c *BaseCommittee) SetMembers(addresses []sdk.AccAddress) {
	members := make([]string, len(addresses))
	for i, addr := range addresses {
		members[i] = addr.String()
	}
	c.Members = members
}

// HasMember returns if a committee contains a given member address
func (c *BaseCommittee) HasMember(addr sdk.AccAddress) bool {
	for _, m := range c.GetMembers() {
		if m.Equals(addr) {
			return true
		}
	}
	return false
}

// GetPermissions is a getter for committee permissions
func (c *BaseCommittee) GetPermissions() []Permission {
	permissions := make([]Permission, len(c.Permissions))
	for i, any := range c.Permissions {
		permission, ok := any.GetCachedValue().(Permission)
		if !ok {
			panic("expected base committee permission")
		}
		permissions[i] = permission
	}

	return permissions
}

// SetPermissions is a setter for committee permissions
func (c *BaseCommittee) SetPermissions(permissions []Permission) {
	c.Permissions = PackPermissions(permissions)
}

// HasPermissionsFor returns whether the committee is authorized to enact a proposal.
// As long as one permission allows the proposal then it goes through. Its the OR of all permissions.
func (c BaseCommittee) HasPermissionsFor(ctx sdk.Context, appCdc *codec.Codec, pk ParamKeeper, proposal PubProposal) bool {
	for _, p := range c.GetPermissions() {
		if p.Allows(ctx, appCdc, pk, proposal) {
			return true
		}
	}
	return false
}

// GetVoteThreshold is a getter for committee VoteThreshold
func (c BaseCommittee) GetVoteThreshold() sdk.Dec { return c.VoteThreshold }

// SetVoteThreshold is a setter for committee VoteThreshold
func (c *BaseCommittee) SetVoteThreshold(voteThreshold sdk.Dec) {
	c.VoteThreshold = voteThreshold
}

// GetProposalDuration is a getter for committee ProposalDuration
func (c BaseCommittee) GetProposalDuration() time.Duration { return c.ProposalDuration }

// SetProposalDuration is a setter for committee ProposalDuration
func (c *BaseCommittee) SetProposalDuration(proposalDuration time.Duration) {
	c.ProposalDuration = proposalDuration
}

// GetTallyOption is a getter for committee TallyOption
func (c BaseCommittee) GetTallyOption() TallyOption { return c.TallyOption }

// String implements Stringer interface
func (c BaseCommittee) String() string {
	// TODO: Question, should we just use yaml to output or use our previous implementation?
	out, _ := yaml.Marshal(c)
	return string(out)
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (c BaseCommittee) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for _, any := range c.Permissions {
		var permission []Permission
		err := unpacker.UnpackAny(any, &permission)
		if err != nil {
			return err
		}
	}
	return nil
}

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
		if _, ok := addressMap[m]; ok {
			return fmt.Errorf("committee cannot have duplicate members, %s", m)
		}
		// check for valid addresses
		if _, err := sdk.AccAddressFromBech32(m); err != nil {
			return err
		}
		addressMap[m] = true
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

// NewMemberCommittee instantiates a new instance of MemberCommittee
func NewMemberCommittee(id uint64, description string, members []string, permissions []Permission,
	threshold sdk.Dec, duration time.Duration, tallyOption TallyOption) *MemberCommittee {
	permissionsAny := PackPermissions(permissions)
	return &MemberCommittee{
		BaseCommittee: &BaseCommittee{
			Id:               id,
			Description:      description,
			Members:          members,
			Permissions:      permissionsAny,
			VoteThreshold:    threshold,
			ProposalDuration: duration,
			TallyOption:      tallyOption,
		},
	}
}

// GetType is a getter for committee type
func (c MemberCommittee) GetType() string { return MemberCommitteeType }

// ------------------------------------------
//				Proposals
// ------------------------------------------

// PubProposal is the interface that all proposals must fulfill to be submitted to a committee.
// Proposal types can be created external to this module. For example a ParamChangeProposal, or CommunityPoolSpendProposal.
// It is pinned to the equivalent type in the gov module to create compatibility between proposal types.
type PubProposal govtypes.Content

func NewProposal(pubProposal PubProposal, id uint64, committeeId uint64, deadline time.Time) Proposal {
	msg, ok := pubProposal.(proto.Message)
	if !ok {
		panic(fmt.Errorf("%T does not implement proto.Message", pubProposal))
	}
	pubProposalAny, err := types.NewAnyWithValue(msg)
	if err != nil {
		panic(err)
	}
	return Proposal{
		Any:         pubProposalAny,
		Id:          id,
		CommitteeId: committeeId,
		Deadline:    deadline,
	}
}

func (p Proposal) GetPubProposal() PubProposal {
	if p.Any == nil {
		return nil
	}
	pubProposal, ok := p.Any.GetCachedValue().(PubProposal)
	if !ok {
		return nil
	}
	return pubProposal
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

func NewVote(proposalId uint64, voter string, voteType VoteType) Vote {
	return Vote{
		ProposalId: proposalId,
		Voter:      voter,
		VoteType:   voteType,
	}
}

func (v Vote) Validate() error {
	if v.Voter == "" {
		return fmt.Errorf("voter address cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(v.Voter); err != nil {
		return err
	}

	return v.VoteType.Validate()
}
