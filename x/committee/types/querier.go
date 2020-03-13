package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Query endpoints supported by the Querier
const (
	//QueryParams     = "params"
	QueryCommittees = "committees"
	QueryCommittee  = "committee"
	QueryProposals  = "proposals"
	QueryProposal   = "proposal"
	QueryVotes      = "votes"
	QueryVote       = "vote"
	QueryTally      = "tally"
)

type QueryCommitteeParams struct {
	CommitteeID uint64
}

func NewQueryCommitteeParams(committeeID uint64) QueryCommitteeParams {
	return QueryCommitteeParams{
		CommitteeID: committeeID,
	}
}

type QueryProposalParams struct {
	ProposalID uint64
}

func NewQueryProposalParams(proposalID uint64) QueryProposalParams {
	return QueryProposalParams{
		ProposalID: proposalID,
	}
}

type QueryVoteParams struct {
	ProposalID uint64
	Voter      sdk.AccAddress
}

func NewQueryVoteParams(proposalID uint64, voter sdk.AccAddress) QueryVoteParams {
	return QueryVoteParams{
		ProposalID: proposalID,
		Voter:      voter,
	}
}
