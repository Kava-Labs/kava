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

## Store

For complete implementation details for how items are stored, see [keys.go](../types/keys.go). The committee module store state consists of committees, proposals, and votes. When a proposal expires or passes, the proposal and associated votes are deleted from state.
