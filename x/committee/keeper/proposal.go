package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/committee/types"
)

func (k Keeper) SubmitProposal(ctx sdk.Context, proposer sdk.AccAddress, committeeID uint64, pubProposal types.PubProposal) (uint64, sdk.Error) {
	// Limit proposals to only be submitted by committee members
	com, found := k.GetCommittee(ctx, committeeID)
	if !found {
		return 0, sdk.ErrInternal("committee doesn't exist")
	}
	if !com.HasMember(proposer) {
		return 0, sdk.ErrInternal("only member can propose proposals")
	}

	// Check committee has permissions to enact proposal.
	if !com.HasPermissionsFor(pubProposal) {
		return 0, sdk.ErrInternal("committee does not have permissions to enact proposal")
	}

	// Check proposal is valid
	if err := k.ValidatePubProposal(ctx, pubProposal); err != nil {
		return 0, err
	}

	// Get a new ID and store the proposal
	deadline := ctx.BlockTime().Add(types.MaxProposalDuration)
	return k.StoreNewProposal(ctx, pubProposal, committeeID, deadline)
}

func (k Keeper) AddVote(ctx sdk.Context, proposalID uint64, voter sdk.AccAddress) sdk.Error {
	// Validate
	pr, found := k.GetProposal(ctx, proposalID)
	if !found {
		return sdk.ErrInternal("proposal not found")
	}
	if pr.HasExpiredBy(ctx.BlockTime()) {
		return sdk.ErrInternal("proposal expired")
	}
	com, found := k.GetCommittee(ctx, pr.CommitteeID)
	if !found {
		return sdk.ErrInternal("committee disbanded")
	}
	if !com.HasMember(voter) {
		return sdk.ErrInternal("not authorized to vote on proposal")
	}

	// Store vote, overwriting any prior vote
	k.SetVote(ctx, types.Vote{ProposalID: proposalID, Voter: voter})

	return nil
}

func (k Keeper) CloseOutProposal(ctx sdk.Context, proposalID uint64) sdk.Error {
	pr, found := k.GetProposal(ctx, proposalID)
	if !found {
		return sdk.ErrInternal("proposal not found")
	}
	com, found := k.GetCommittee(ctx, pr.CommitteeID)
	if !found {
		return sdk.ErrInternal("committee disbanded")
	}

	var votes []types.Vote
	k.IterateVotes(ctx, proposalID, func(vote types.Vote) bool {
		votes = append(votes, vote)
		return false
	})
	proposalPasses := sdk.NewDec(int64(len(votes))).GTE(types.VoteThreshold.MulInt64(int64(len(com.Members))))

	if proposalPasses {
		// eneact vote
		// The proposal handler may execute state mutating logic depending
		// on the proposal content. If the handler fails, no state mutation
		// is written and the error message is logged.
		handler := k.router.GetRoute(pr.ProposalRoute())
		cacheCtx, writeCache := ctx.CacheContext()
		err := handler(cacheCtx, pr.PubProposal) // need to pass pubProposal as the handlers type assert it into the concrete types
		if err == nil {
			// write state to the underlying multi-store
			writeCache()
		} // if handler returns error, then still delete the proposal - it's still over, but send an event
	}
	if proposalPasses || pr.HasExpiredBy(ctx.BlockTime()) {

		// delete proposal and votes
		k.DeleteProposal(ctx, proposalID)
		for _, v := range votes {
			k.DeleteVote(ctx, v.ProposalID, v.Voter)
		}
		return nil
	}
	return sdk.ErrInternal("note enough votes to close proposal")
}

func (k Keeper) ValidatePubProposal(ctx sdk.Context, pubProposal types.PubProposal) sdk.Error {
	if pubProposal == nil {
		return sdk.ErrInternal("proposal is empty")
	}
	if err := pubProposal.ValidateBasic(); err != nil {
		return err
	}

	if !k.router.HasRoute(pubProposal.ProposalRoute()) {
		return sdk.ErrInternal("no handler found for proposal")
	}

	// Execute the proposal content in a cache-wrapped context to validate the
	// actual parameter changes before the proposal proceeds through the
	// governance process. State is not persisted.
	cacheCtx, _ := ctx.CacheContext()
	handler := k.router.GetRoute(pubProposal.ProposalRoute())
	if err := handler(cacheCtx, pubProposal); err != nil {
		return err
	}

	return nil
}
