package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/committee/types"
)

// SubmitProposal adds a proposal to a committee so that it can be voted on.
func (k Keeper) SubmitProposal(ctx sdk.Context, proposer sdk.AccAddress, committeeID uint64, pubProposal types.PubProposal) (uint64, error) {
	// Limit proposals to only be submitted by committee members
	com, found := k.GetCommittee(ctx, committeeID)
	if !found {
		return 0, sdkerrors.Wrapf(types.ErrUnknownCommittee, "%d", committeeID)
	}
	if !com.HasMember(proposer) {
		return 0, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "proposer not member of committee")
	}

	// Check committee has permissions to enact proposal.
	if !com.HasPermissionsFor(ctx, k.cdc, k.ParamKeeper, pubProposal) {
		return 0, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "committee does not have permissions to enact proposal")
	}

	// Check proposal is valid
	if err := k.ValidatePubProposal(ctx, pubProposal); err != nil {
		return 0, err
	}

	// Get a new ID and store the proposal
	deadline := ctx.BlockTime().Add(com.ProposalDuration)
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
func (k Keeper) AddVote(ctx sdk.Context, proposalID uint64, voter sdk.AccAddress) error {
	// Validate
	pr, found := k.GetProposal(ctx, proposalID)
	if !found {
		return sdkerrors.Wrapf(types.ErrUnknownProposal, "%d", proposalID)
	}
	if pr.HasExpiredBy(ctx.BlockTime()) {
		return sdkerrors.Wrapf(types.ErrProposalExpired, "%s ≥ %s", ctx.BlockTime(), pr.Deadline)

	}
	com, found := k.GetCommittee(ctx, pr.CommitteeID)
	if !found {
		return sdkerrors.Wrapf(types.ErrUnknownCommittee, "%d", pr.CommitteeID)
	}
	if !com.HasMember(voter) {
		return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "voter must be a member of committee")
	}

	// Store vote, overwriting any prior vote
	k.SetVote(ctx, types.NewVote(proposalID, voter))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposalVote,
			sdk.NewAttribute(types.AttributeKeyCommitteeID, fmt.Sprintf("%d", com.ID)),
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", pr.ID)),
			sdk.NewAttribute(types.AttributeKeyVoter, voter.String()),
		),
	)
	return nil
}

// GetProposalResult calculates if a proposal currently has enough votes to pass.
func (k Keeper) GetProposalResult(ctx sdk.Context, proposalID uint64) (bool, error) {
	pr, found := k.GetProposal(ctx, proposalID)
	if !found {
		return false, sdkerrors.Wrapf(types.ErrUnknownProposal, "%d", proposalID)
	}
	com, found := k.GetCommittee(ctx, pr.CommitteeID)
	if !found {
		return false, sdkerrors.Wrapf(types.ErrUnknownCommittee, "%d", pr.CommitteeID)
	}

	numVotes := k.TallyVotes(ctx, proposalID)

	proposalResult := sdk.NewDec(numVotes).GTE(com.VoteThreshold.MulInt64(int64(len(com.Members))))

	return proposalResult, nil
}

// TallyVotes counts all the votes on a proposal
func (k Keeper) TallyVotes(ctx sdk.Context, proposalID uint64) int64 {

	votes := k.GetVotesByProposal(ctx, proposalID)

	return int64(len(votes))
}

// EnactProposal makes the changes proposed in a proposal.
func (k Keeper) EnactProposal(ctx sdk.Context, proposalID uint64) error {
	pr, found := k.GetProposal(ctx, proposalID)
	if !found {
		return sdkerrors.Wrapf(types.ErrUnknownProposal, "%d", proposalID)
	}
	// Check committee still has permissions for the proposal
	// Since the proposal was submitted params could have changed, invalidating the permission of the committee.
	com, found := k.GetCommittee(ctx, pr.CommitteeID)
	if !found {
		return sdkerrors.Wrapf(types.ErrUnknownCommittee, "%d", pr.CommitteeID)
	}
	if !com.HasPermissionsFor(ctx, k.cdc, k.ParamKeeper, pr.PubProposal) {
		return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "committee does not have permissions to enact proposal")
	}

	if err := k.ValidatePubProposal(ctx, pr.PubProposal); err != nil {
		return err
	}

	// enact the proposal
	handler := k.router.GetRoute(pr.ProposalRoute())
	if err := handler(ctx, pr.PubProposal); err != nil {
		// the handler should not error as it was checked in ValidatePubProposal
		panic(fmt.Sprintf("unexpected handler error: %s", err))
	}
	return nil
}

// EnactPassedProposals puts in place the changes proposed in any proposal that has enough votes
func (k Keeper) EnactPassedProposals(ctx sdk.Context) {
	k.IterateProposals(ctx, func(proposal types.Proposal) bool {
		passes, err := k.GetProposalResult(ctx, proposal.ID)
		if err != nil {
			panic(err)
		}
		if !passes {
			return false
		}

		err = k.EnactProposal(ctx, proposal.ID)
		outcome := types.AttributeValueProposalPassed
		if err != nil {
			outcome = types.AttributeValueProposalFailed
		}

		k.DeleteProposalAndVotes(ctx, proposal.ID)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeProposalClose,
				sdk.NewAttribute(types.AttributeKeyCommitteeID, fmt.Sprintf("%d", proposal.CommitteeID)),
				sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposal.ID)),
				sdk.NewAttribute(types.AttributeKeyProposalCloseStatus, outcome),
			),
		)
		return false
	})
}

// CloseExpiredProposals removes proposals (and associated votes) that have past their deadline.
func (k Keeper) CloseExpiredProposals(ctx sdk.Context) {

	k.IterateProposals(ctx, func(proposal types.Proposal) bool {
		if !proposal.HasExpiredBy(ctx.BlockTime()) {
			return false
		}

		k.DeleteProposalAndVotes(ctx, proposal.ID)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeProposalClose,
				sdk.NewAttribute(types.AttributeKeyCommitteeID, fmt.Sprintf("%d", proposal.CommitteeID)),
				sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposal.ID)),
				sdk.NewAttribute(types.AttributeKeyProposalCloseStatus, types.AttributeValueProposalTimeout),
			),
		)
		return false
	})
}

// ValidatePubProposal checks if a pubproposal is valid.
func (k Keeper) ValidatePubProposal(ctx sdk.Context, pubProposal types.PubProposal) (returnErr error) {
	if pubProposal == nil {
		return sdkerrors.Wrap(types.ErrInvalidPubProposal, "pub proposal cannot be nil")
	}
	if err := pubProposal.ValidateBasic(); err != nil {
		return err
	}

	if !k.router.HasRoute(pubProposal.ProposalRoute()) {
		return sdkerrors.Wrapf(types.ErrNoProposalHandlerExists, "%T", pubProposal)
	}

	// Run the proposal's changes through the associated handler using a cached version of state to ensure changes are not permanent.
	cacheCtx, _ := ctx.CacheContext()
	handler := k.router.GetRoute(pubProposal.ProposalRoute())

	// Handle an edge case where a param change proposal causes the proposal handler to panic.
	// A param change proposal with a registered subspace value but unregistered key value will cause a panic in the param change proposal handler.
	// This defer will catch panics and return a normal error: `recover()` gets the panic value, then the enclosing function's return value is swapped for an error.
	// reference: https://stackoverflow.com/questions/33167282/how-to-return-a-value-in-a-go-function-that-panics?noredirect=1&lq=1
	defer func() {
		if r := recover(); r != nil {
			returnErr = sdkerrors.Wrapf(types.ErrInvalidPubProposal, "proposal handler panicked: %s", r)
		}
	}()

	if err := handler(cacheCtx, pubProposal); err != nil {
		return err
	}
	return nil
}
