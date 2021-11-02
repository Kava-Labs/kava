package types

import (
	fmt "fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	types "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	proto "github.com/gogo/protobuf/proto"
	"gopkg.in/yaml.v2"
)

// Committee is an interface for handling common actions on committees
type Committee interface {
	GetId() uint64
	GetType() string
	GetDescription() string

	GetMembers() []sdk.AccAddress
	SetMembers([]sdk.AccAddress) BaseCommittee
	HasMember(addr sdk.AccAddress) bool

	GetPermissions() []Permission
	SetPermissions([]Permission) Committee
	HasPermissionsFor(ctx sdk.Context, appCdc *codec.Codec, pk ParamKeeper, proposal PubProposal) bool

	GetProposalDuration() time.Duration
	SetProposalDuration(time.Duration) BaseCommittee

	GetVoteThreshold() sdk.Dec
	SetVoteThreshold(sdk.Dec) BaseCommittee

	GetTallyOption() TallyOption
	Validate() error
}

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
