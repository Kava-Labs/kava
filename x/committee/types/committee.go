package types

import (
	fmt "fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	proto "github.com/gogo/protobuf/proto"
	"sigs.k8s.io/yaml"
)

const MaxCommitteeDescriptionLength int = 512

const (
	BaseCommitteeType   = "kava/BaseCommittee"
	MemberCommitteeType = "kava/MemberCommittee" // Committee is composed of member addresses that vote to enact proposals within their permissions
	TokenCommitteeType  = "kava/TokenCommittee"  // Committee is composed of token holders with voting power determined by total token balance
	BondDenom           = "ukava"
)

// Marshal needed for protobuf compatibility.
func (t TallyOption) Marshal() ([]byte, error) {
	return []byte{byte(t)}, nil
}

// Unmarshal needed for protobuf compatibility.
func (t *TallyOption) Unmarshal(data []byte) error {
	*t = TallyOption(data[0])
	return nil
}

// Committee is an interface for handling common actions on committees
type Committee interface {
	codec.ProtoMarshaler
	codectypes.UnpackInterfacesMessage

	GetID() uint64
	GetType() string
	GetDescription() string

	GetMembers() []sdk.AccAddress
	SetMembers([]sdk.AccAddress)
	HasMember(addr sdk.AccAddress) bool

	GetPermissions() []Permission
	SetPermissions([]Permission)
	HasPermissionsFor(ctx sdk.Context, appCdc codec.Codec, pk ParamKeeper, proposal PubProposal) bool

	GetProposalDuration() time.Duration
	SetProposalDuration(time.Duration)

	GetVoteThreshold() sdk.Dec
	SetVoteThreshold(sdk.Dec)

	GetTallyOption() TallyOption
	Validate() error

	String() string
}

var (
	_ Committee                          = &BaseCommittee{}
	_ codectypes.UnpackInterfacesMessage = &Committees{}
)

type Committees []Committee

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (c Committees) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for _, committee := range c {
		if err := committee.UnpackInterfaces(unpacker); err != nil {
			return err
		}
	}
	return nil
}

// GetType is a getter for committee type
func (c *BaseCommittee) GetType() string { return BaseCommitteeType }

// GetID is a getter for committee ID
func (c *BaseCommittee) GetID() uint64 { return c.ID }

// GetDescription is a getter for committee description
func (c *BaseCommittee) GetDescription() string { return c.Description }

// GetMembers is a getter for committee members
func (c BaseCommittee) GetMembers() []sdk.AccAddress { return c.Members }

// SetMembers is a setter for committee members
func (c *BaseCommittee) SetMembers(members []sdk.AccAddress) { c.Members = members }

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
	permissions, err := UnpackPermissions(c.Permissions)
	if err != nil {
		panic(err)
	}
	return permissions
}

// SetPermissions is a setter for committee permissions
func (c *BaseCommittee) SetPermissions(permissions []Permission) {
	if len(permissions) == 0 {
		c.Permissions = nil
	}
	permissionsAny, err := PackPermissions(permissions)
	if err != nil {
		panic(err)
	}
	c.Permissions = permissionsAny
}

// HasPermissionsFor returns whether the committee is authorized to enact a proposal.
// As long as one permission allows the proposal then it goes through. Its the OR of all permissions.
func (c BaseCommittee) HasPermissionsFor(ctx sdk.Context, appCdc codec.Codec, pk ParamKeeper, proposal PubProposal) bool {
	for _, p := range c.GetPermissions() {
		if p.Allows(ctx, pk, proposal) {
			return true
		}
	}
	return false
}

// String implements fmt.Stringer
func (c BaseCommittee) String() string {
	return fmt.Sprintf(`Committee %d:
	Description:              %s
	Members:               %s
  	Permissions:               			%s
  	VoteThreshold:            		  %s
	ProposalDuration:        						%s
	TallyOption:   						%s`,
		c.ID, c.Description, c.GetMembers(), c.Permissions,
		c.VoteThreshold.String(), c.ProposalDuration.String(),
		c.TallyOption.String(),
	)
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

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (c BaseCommittee) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for _, any := range c.Permissions {
		var permission Permission
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
		if _, ok := addressMap[m.String()]; ok {
			return fmt.Errorf("committee cannot have duplicate members, %s", m)
		}
		// check for valid addresses
		if m.Empty() {
			return fmt.Errorf("committee cannot have empty member address")
		}
		addressMap[m.String()] = true
	}

	// validate permissions
	permissions, err := UnpackPermissions(c.Permissions)
	if err != nil {
		return err
	}
	for _, p := range permissions {
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
func NewMemberCommittee(id uint64, description string, members []sdk.AccAddress, permissions []Permission,
	threshold sdk.Dec, duration time.Duration, tallyOption TallyOption,
) (*MemberCommittee, error) {
	permissionsAny, err := PackPermissions(permissions)
	if err != nil {
		return nil, err
	}
	return &MemberCommittee{
		BaseCommittee: &BaseCommittee{
			ID:               id,
			Description:      description,
			Members:          members,
			Permissions:      permissionsAny,
			VoteThreshold:    threshold,
			ProposalDuration: duration,
			TallyOption:      tallyOption,
		},
	}, nil
}

// MustNewMemberCommittee instantiates a new instance of MemberCommittee and panics on error
func MustNewMemberCommittee(id uint64, description string, members []sdk.AccAddress, permissions []Permission,
	threshold sdk.Dec, duration time.Duration, tallyOption TallyOption,
) *MemberCommittee {
	committee, err := NewMemberCommittee(id, description, members, permissions, threshold, duration, tallyOption)
	if err != nil {
		panic(err)
	}
	return committee
}

// GetType is a getter for committee type
func (c MemberCommittee) GetType() string { return MemberCommitteeType }

// NewTokenCommittee instantiates a new instance of TokenCommittee
func NewTokenCommittee(id uint64, description string, members []sdk.AccAddress, permissions []Permission,
	threshold sdk.Dec, duration time.Duration, tallyOption TallyOption, quorum sdk.Dec, tallyDenom string,
) (*TokenCommittee, error) {
	permissionsAny, err := PackPermissions(permissions)
	if err != nil {
		return nil, err
	}
	return &TokenCommittee{
		BaseCommittee: &BaseCommittee{
			ID:               id,
			Description:      description,
			Members:          members,
			Permissions:      permissionsAny,
			VoteThreshold:    threshold,
			ProposalDuration: duration,
			TallyOption:      tallyOption,
		},
		Quorum:     quorum,
		TallyDenom: tallyDenom,
	}, nil
}

// MustNewTokenCommittee instantiates a new instance of TokenCommittee and panics on error
func MustNewTokenCommittee(id uint64, description string, members []sdk.AccAddress, permissions []Permission,
	threshold sdk.Dec, duration time.Duration, tallyOption TallyOption, quorum sdk.Dec, tallyDenom string,
) *TokenCommittee {
	committee, err := NewTokenCommittee(id, description, members, permissions, threshold, duration, tallyOption, quorum, tallyDenom)
	if err != nil {
		panic(err)
	}
	return committee
}

// GetType is a getter for committee type
func (c TokenCommittee) GetType() string { return TokenCommitteeType }

// GetQuorum returns the quorum of the committee
func (c TokenCommittee) GetQuorum() sdk.Dec { return c.Quorum }

// GetTallyDenom returns the tally denom of the committee
func (c TokenCommittee) GetTallyDenom() string { return c.TallyDenom }

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
type PubProposal govv1beta1.Content

var (
	_ PubProposal                        = Proposal{}
	_ codectypes.UnpackInterfacesMessage = Proposals{}
)

type Proposals []Proposal

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (p Proposals) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for _, committee := range p {
		if err := committee.UnpackInterfaces(unpacker); err != nil {
			return err
		}
	}
	return nil
}

// NewProposal instantiates a new instance of Proposal
func NewProposal(pubProposal PubProposal, id uint64, committeeID uint64, deadline time.Time) (Proposal, error) {
	msg, ok := pubProposal.(proto.Message)
	if !ok {
		return Proposal{}, fmt.Errorf("%T does not implement proto.Message", pubProposal)
	}
	proposalAny, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return Proposal{}, err
	}
	return Proposal{
		Content:     proposalAny,
		ID:          id,
		CommitteeID: committeeID,
		Deadline:    deadline,
	}, nil
}

// MustNewProposal instantiates a new instance of Proposal and panics if there is an error
func MustNewProposal(pubProposal PubProposal, id uint64, committeeID uint64, deadline time.Time) Proposal {
	proposal, err := NewProposal(pubProposal, id, committeeID, deadline)
	if err != nil {
		panic(err)
	}
	return proposal
}

// GetPubProposal returns the PubProposal (govtypes.Content)
func (p Proposal) GetContent() PubProposal {
	content, ok := p.Content.GetCachedValue().(PubProposal)
	if !ok {
		return nil
	}
	return content
}

// String implements the fmt.Stringer interface.
func (p Proposal) String() string {
	bz, _ := yaml.Marshal(p)
	return string(bz)
}

func (p Proposal) GetTitle() string {
	content := p.GetContent()
	if content == nil {
		return ""
	}
	return content.GetTitle()
}

func (p Proposal) GetDescription() string {
	content := p.GetContent()
	if content == nil {
		return ""
	}
	return content.GetDescription()
}

func (p Proposal) ProposalRoute() string {
	content := p.GetContent()
	if content == nil {
		return ""
	}
	return content.ProposalRoute()
}

func (p Proposal) ProposalType() string {
	content := p.GetContent()
	if content == nil {
		return ""
	}
	return content.ProposalType()
}

func (p Proposal) ValidateBasic() error {
	content := p.GetContent()
	if content == nil {
		return nil
	}
	return content.ValidateBasic()
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (p Proposal) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var content PubProposal
	return unpacker.UnpackAny(p.Content, &content)
}

// HasExpiredBy calculates if the proposal will have expired by a certain time.
// All votes must be cast before deadline, those cast at time == deadline are not valid
func (p Proposal) HasExpiredBy(time time.Time) bool {
	return !time.Before(p.Deadline)
}

// NewVote instantiates a new instance of Vote
func NewVote(proposalID uint64, voter sdk.AccAddress, voteType VoteType) Vote {
	return Vote{
		ProposalID: proposalID,
		Voter:      voter,
		VoteType:   voteType,
	}
}

// Validates Vote fields
func (v Vote) Validate() error {
	if v.Voter.Empty() {
		return fmt.Errorf("voter address cannot be empty")
	}

	return v.VoteType.Validate()
}
