<!--
order: 3
-->

# Messages

Committee members submit proposals using a `MsgSubmitProposal`

```go
// MsgSubmitProposal is used by committee members to create a new proposal that they can vote on.
type MsgSubmitProposal struct {
  PubProposal PubProposal    `json:"pub_proposal" yaml:"pub_proposal"`
  Proposer    sdk.AccAddress `json:"proposer" yaml:"proposer"`
  CommitteeID uint64         `json:"committee_id" yaml:"committee_id"`
}
```

## State Modifications

- Generate new `ProposalID`
- Create new `Proposal` with deadline equal to the time that the proposal will expire.

Valid votes include 'yes', 'no', and 'abstain'.

```go
// MsgVote is submitted by committee members to vote on proposals.
type MsgVote struct {
	ProposalID uint64         `json:"proposal_id" yaml:"proposal_id"`
	Voter      sdk.AccAddress `json:"voter" yaml:"voter"`
	VoteType   VoteType       `json:"vote_type" yaml:"vote_type"`
}
```

## State Modifications

- Create a new `Vote`
- When the proposal is evaluated:
  - Enact the proposal (passed proposals may cause state modifications)
  - Delete the proposal and associated votes
