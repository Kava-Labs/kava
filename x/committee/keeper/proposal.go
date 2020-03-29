package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/committee/types"
)

// SubmitProposal adds a proposal to a committee so that it can be voted on.
func (k Keeper) SubmitProposal(ctx sdk.Context, proposer sdk.AccAddress, committeeID uint64, pubProposal types.PubProposal) (uint64, sdk.Error) {
	// Limit proposals to only be submitted by committee members
	com, found := k.GetCommittee(ctx, committeeID)
	if !found {
		return 0, types.ErrUnknownCommittee(k.codespace, committeeID)
	}
	if !com.HasMember(proposer) {
		return 0, sdk.ErrUnauthorized("proposer not member of committee")
	}

	// Check committee has permissions to enact proposal.
	if !com.HasPermissionsFor(pubProposal) {
		return 0, sdk.ErrUnauthorized("committee does not have permissions to enact proposal")
	}

	// Check proposal is valid
	if err := k.ValidatePubProposal(ctx, pubProposal); err != nil {
		return 0, err
	}

	// Get a new ID and store the proposal
	deadline := ctx.BlockTime().Add(com.MaxProposalDuration)
	proposalID, err := k.StoreNewProposal(ctx, pubProposal, committeeID, deadline)
	if err != nil {
		return 0, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposalSubmit,
			sdk.NewAttribute(types.AttributeKeyCommitteeID, fmt.Sprintf("%d", com.ID)),
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposalID)),
		),
	)
	return proposalID, nil
}

// AddVote submits a vote on a proposal.
func (k Keeper) AddVote(ctx sdk.Context, proposalID uint64, voter sdk.AccAddress) sdk.Error {
	// Validate
	pr, found := k.GetProposal(ctx, proposalID)
	if !found {
		return types.ErrUnknownProposal(k.codespace, proposalID)
	}
	if pr.HasExpiredBy(ctx.BlockTime()) {
		return types.ErrProposalExpired(k.codespace, ctx.BlockTime(), pr.Deadline)
	}
	com, found := k.GetCommittee(ctx, pr.CommitteeID)
	if !found {
		return types.ErrUnknownCommittee(k.codespace, pr.CommitteeID)
	}
	if !com.HasMember(voter) {
		return sdk.ErrUnauthorized("voter must be a member of committee")
	}

	// Store vote, overwriting any prior vote
	k.SetVote(ctx, types.Vote{ProposalID: proposalID, Voter: voter})

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposalVote,
			sdk.NewAttribute(types.AttributeKeyCommitteeID, fmt.Sprintf("%d", com.ID)),
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", pr.ID)),
		),
	)
	return nil
}

// GetProposalResult calculates if a proposal currently has enough votes to pass.
func (k Keeper) GetProposalResult(ctx sdk.Context, proposalID uint64) (bool, sdk.Error) {
	pr, found := k.GetProposal(ctx, proposalID)
	if !found {
		return false, types.ErrUnknownProposal(k.codespace, proposalID)
	}
	com, found := k.GetCommittee(ctx, pr.CommitteeID)
	if !found {
		return false, types.ErrUnknownCommittee(k.codespace, pr.CommitteeID)
	}

	numVotes := k.TallyVotes(ctx, proposalID)

	proposalResult := sdk.NewDec(numVotes).GTE(com.VoteThreshold.MulInt64(int64(len(com.Members))))

	return proposalResult, nil
}

// TallyVotes counts all the votes on a proposal
func (k Keeper) TallyVotes(ctx sdk.Context, proposalID uint64) int64 {

	var votes []types.Vote
	k.IterateVotes(ctx, proposalID, func(vote types.Vote) bool {
		votes = append(votes, vote)
		return false
	})

	return int64(len(votes))
}

// EnactProposal makes the changes proposed in a proposal.
func (k Keeper) EnactProposal(ctx sdk.Context, proposalID uint64) sdk.Error {
	pr, found := k.GetProposal(ctx, proposalID)
	if !found {
		return types.ErrUnknownProposal(k.codespace, proposalID)
	}

	// Run the proposal's changes through the associated handler, but using a cached version of state to ensure changes are not permanent if an error occurs.
	handler := k.router.GetRoute(pr.ProposalRoute())
	cacheCtx, writeCache := ctx.CacheContext()
	if err := handler(cacheCtx, pr.PubProposal); err != nil {
		return err
	}
	// write state to the underlying multi-store
	writeCache()
	return nil
}

// ValidatePubProposal checks if a pubproposal is valid.
func (k Keeper) ValidatePubProposal(ctx sdk.Context, pubProposal types.PubProposal) sdk.Error {
	if pubProposal == nil {
		return types.ErrInvalidPubProposal(k.codespace, "pub proposal cannot be nil")
	}
	if err := pubProposal.ValidateBasic(); err != nil {
		return err
	}

	if !k.router.HasRoute(pubProposal.ProposalRoute()) {
		return types.ErrNoProposalHandlerExists(k.codespace, pubProposal)
	}

	// Run the proposal's changes through the associated handler using a cached version of state to ensure changes are not permanent.
	cacheCtx, _ := ctx.CacheContext()
	handler := k.router.GetRoute(pubProposal.ProposalRoute())
	if err := handler(cacheCtx, pubProposal); err != nil {
		return err
	}
	return nil
}

// DeleteProposalAndVotes removes a proposal and its associated votes.
func (k Keeper) DeleteProposalAndVotes(ctx sdk.Context, proposalID uint64) {
	var votes []types.Vote
	k.IterateVotes(ctx, proposalID, func(vote types.Vote) bool {
		votes = append(votes, vote)
		return false
	})

	k.DeleteProposal(ctx, proposalID)
	for _, v := range votes {
		k.DeleteVote(ctx, v.ProposalID, v.Voter)
	}
}
