package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Query endpoints supported by the Querier
const (
	QueryCommittees = "committees"
	QueryCommittee  = "committee"
	QueryProposals  = "proposals"
	QueryProposal   = "proposal"
	QueryVotes      = "votes"
	QueryVote       = "vote"
	QueryTally      = "tally"
)

type QueryCommitteeParams struct {
	CommitteeID uint64 `json:"committee_id" yaml:"committee_id"`
}

func NewQueryCommitteeParams(committeeID uint64) QueryCommitteeParams {
	return QueryCommitteeParams{
		CommitteeID: committeeID,
	}
}

type QueryProposalParams struct {
	ProposalID uint64 `json:"proposal_id" yaml:"proposal_id"`
}

func NewQueryProposalParams(proposalID uint64) QueryProposalParams {
	return QueryProposalParams{
		ProposalID: proposalID,
	}
}

type QueryVoteParams struct {
	ProposalID uint64         `json:"proposal_id" yaml:"proposal_id"`
	Voter      sdk.AccAddress `json:"voter" yaml:"voter"`
}

func NewQueryVoteParams(proposalID uint64, voter sdk.AccAddress) QueryVoteParams {
	return QueryVoteParams{
		ProposalID: proposalID,
		Voter:      voter,
	}
}
