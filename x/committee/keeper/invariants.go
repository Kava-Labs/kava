package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/committee/types"
)

// RegisterInvariants registers all staking invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {

	ir.RegisterRoute(types.ModuleName, "valid-committees",
		ValidCommitteesInvariant(k))
	ir.RegisterRoute(types.ModuleName, "valid-proposals",
		ValidProposalsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "valid-votes",
		ValidVotesInvariant(k))
}

// ValidCommitteesInvariant verifies that all committees in the store are independently valid
func ValidCommitteesInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {

		var validationErr error
		var invalidCommittee types.Committee
		k.IterateCommittees(ctx, func(com types.Committee) bool {

			if err := com.Validate(); err != nil {
				validationErr = err
				invalidCommittee = com
				return true
			}
			return false
		})

		broken := validationErr != nil
		invariantMessage := sdk.FormatInvariant(
			types.ModuleName,
			"valid committees",
			fmt.Sprintf(
				"\tfound invalid committee, reason: %s\n"+
					"\tcommittee:\n\t%+v\n",
				validationErr, invalidCommittee),
		)
		return invariantMessage, broken
	}
}

// ValidProposalsInvariant verifies that all proposals in the store are valid
func ValidProposalsInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {

		var validationErr error
		var invalidProposal types.Proposal
		k.IterateProposals(ctx, func(proposal types.Proposal) bool {
			invalidProposal = proposal

			if err := proposal.PubProposal.ValidateBasic(); err != nil {
				validationErr = err
				return true
			}

			currentTime := ctx.BlockTime()
			if !currentTime.Equal(time.Time{}) { // this avoids a simulator bug where app.InitGenesis is called with blockTime=0 instead of the correct time
				if proposal.Deadline.Before(currentTime) {
					validationErr = fmt.Errorf("deadline after current block time %s", currentTime)
					return true
				}
			}

			com, found := k.GetCommittee(ctx, proposal.CommitteeID)
			if !found {
				validationErr = fmt.Errorf("proposal has no committee %d", proposal.CommitteeID)
				return true
			}

			if !com.HasPermissionsFor(proposal.PubProposal) {
				validationErr = fmt.Errorf("proposal not permitted for committee %+v", com)
				return true
			}

			return false
		})

		broken := validationErr != nil
		invariantMessage := sdk.FormatInvariant(
			types.ModuleName,
			"valid proposals",
			fmt.Sprintf(
				"\tfound invalid proposal, reason: %s\n"+
					"\tproposal:\n\t%s\n",
				validationErr, invalidProposal),
		)
		return invariantMessage, broken
	}
}

// ValidVotesInvariant verifies that all votes in the store are valid
func ValidVotesInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {

		var validationErr error
		var invalidVote types.Vote
		k.IterateVotes(ctx, func(vote types.Vote) bool {
			invalidVote = vote

			if vote.Voter.Empty() {
				validationErr = fmt.Errorf("empty voter address")
				return true
			}

			proposal, found := k.GetProposal(ctx, vote.ProposalID)
			if !found {
				validationErr = fmt.Errorf("vote has no proposal %d", vote.ProposalID)
				return true
			}

			com, found := k.GetCommittee(ctx, proposal.CommitteeID)
			if !found {
				validationErr = fmt.Errorf("vote's proposal has no committee %d", proposal.CommitteeID)
				return true
			}
			if !com.HasMember(vote.Voter) {
				validationErr = fmt.Errorf("voter is not a member of committee %+v", com)
				return true
			}

			return false
		})

		broken := validationErr != nil
		invariantMessage := sdk.FormatInvariant(
			types.ModuleName,
			"valid votes",
			fmt.Sprintf(
				"\tfound invalid vote, reason: %s\n"+
					"\tvote:\n\t%+v\n",
				validationErr, invalidVote),
		)
		return invariantMessage, broken
	}
}
