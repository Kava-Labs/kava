<!--
order: 2
-->

# State

## Genesis state

`GenesisState` defines the state that must be persisted when the blockchain stops/restarts in order for normal function of the committee module to resume.

```go
// GenesisState is state that must be provided at chain genesis.
  type GenesisState struct {
  NextProposalID uint64      `json:"next_proposal_id" yaml:"next_proposal_id"`
  Committees     []Committee `json:"committees" yaml:"committees"`
  Proposals      []Proposal  `json:"proposals" yaml:"proposals"`
  Votes          []Vote      `json:"votes" yaml:"votes"`
  }
```

## Committees

Each committee conforms to the `Committee` interface and is defined as either a `MemberCommittee` or a `TokenCommittee`:

```go
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
	HasPermissionsFor(ctx sdk.Context, appCdc *codec.Codec, pk ParamKeeper, proposal PubProposal) bool

	GetProposalDuration() time.Duration
	SetProposalDuration(time.Duration) BaseCommittee

	GetVoteThreshold() sdk.Dec
	SetVoteThreshold(sdk.Dec) BaseCommittee

	GetTallyOption() TallyOption
	Validate() error
}

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

// MemberCommittee is an alias of BaseCommittee
type MemberCommittee struct {
	BaseCommittee `json:"base_committee" yaml:"base_committee"`
}

// TokenCommittee supports voting on proposals by token holders
type TokenCommittee struct {
	BaseCommittee `json:"base_committee" yaml:"base_committee"`
	Quorum        sdk.Dec `json:"quorum" yaml:"quorum"`
	TallyDenom    string  `json:"tally_denom" yaml:"tally_denom"`
}
```



## Store

For complete implementation details for how items are stored, see [keys.go](../types/keys.go). The committee module store state consists of committees, proposals, and votes. When a proposal expires or passes, the proposal and associated votes are deleted from state.
