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
	deadline := ctx.BlockTime().Add(com.GetProposalDuration())
	proposalID, err := k.StoreNewProposal(ctx, pubProposal, committeeID, deadline)
	if err != nil {
		return 0, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposalSubmit,
			sdk.NewAttribute(types.AttributeKeyCommitteeID, fmt.Sprintf("%d", com.GetID())),
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposalID)),
		),
	)
	return proposalID, nil
}

// AddVote submits a vote on a proposal.
func (k Keeper) AddVote(ctx sdk.Context, proposalID uint64, voter sdk.AccAddress, voteType types.VoteType) error {
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

	if com.GetType() == types.MemberCommitteeType {
		if !com.HasMember(voter) {
			return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "voter must be a member of committee")
		}
		if voteType != types.Yes {
			return sdkerrors.Wrap(types.ErrInvalidVoteType, "member committees only accept yes votes")
		}
	}

	// Store vote, overwriting any prior vote
	k.SetVote(ctx, types.NewVote(proposalID, voter, voteType))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposalVote,
			sdk.NewAttribute(types.AttributeKeyCommitteeID, fmt.Sprintf("%d", com.GetID())),
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", pr.ID)),
			sdk.NewAttribute(types.AttributeKeyVoter, voter.String()),
			sdk.NewAttribute(types.AttributeKeyVote, fmt.Sprintf("%d", voteType)),
		),
	)
	return nil
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

func (k Keeper) ProcessProposals(ctx sdk.Context) {

	k.IterateProposals(ctx, func(proposal types.Proposal) bool {

		committee, found := k.GetCommittee(ctx, proposal.CommitteeID)
		if !found {
			return false // TODO: If committees can't be removed while having active proposals, should panic here
		}

		if proposal.HasExpiredBy(ctx.BlockTime()) {
			passed := k.GetProposalResult(ctx, proposal.ID, committee)
			outcome := types.Failed
			if passed {
				outcome = types.Passed
				err := k.EnactProposal(ctx, proposal)
				if err != nil {
					outcome = types.Invalid
				}
			}
			k.CloseProposal(ctx, proposal, outcome)
		} else {
			if committee.GetTallyOption() == types.FirstPastThePost {
				passed := k.GetProposalResult(ctx, proposal.ID, committee)
				if passed {
					outcome := types.Passed
					err := k.EnactProposal(ctx, proposal)
					if err != nil {
						outcome = types.Invalid
					}
					k.CloseProposal(ctx, proposal, outcome)
				}
			}
		}
		return false
	})
}

func (k Keeper) GetProposalResult(ctx sdk.Context, proposalID uint64, committee types.Committee) bool {
	switch committee.GetType() {
	case types.MemberCommitteeType:
		return k.GetMemberCommitteeProposalResult(ctx, proposalID, committee)
	case types.TokenCommitteeType:
		tokenCommittee := committee.(types.TokenCommittee) // will panic if type assertion isn't met
		return k.GetTokenCommitteeProposalResult(ctx, proposalID, tokenCommittee)
	default: // Should never hit default case
		return false
	}
}

// GetMemberCommitteeProposalResult gets the result of a member committee proposal
func (k Keeper) GetMemberCommitteeProposalResult(ctx sdk.Context, proposalID uint64, committee types.Committee) bool {
	currVotes, possibleVotes, voteThreshold := k.TallyMemberCommitteeVotes(ctx, proposalID, committee)
	return currVotes.GTE(voteThreshold.Mul(possibleVotes)) // vote threshold requirements
}

// TallyMemberCommitteeVotes returns the polling status of a member committee vote. Returns current votes,
// committee members (possible votes), vote threshold (proposal passes at this percentage)
func (k Keeper) TallyMemberCommitteeVotes(ctx sdk.Context, proposalID uint64,
	committee types.Committee) (sdk.Dec, sdk.Dec, sdk.Dec) {
	votes := k.GetVotesByProposal(ctx, proposalID)
	currVotes := sdk.NewDec(int64(len(votes)))
	possibleVotes := sdk.NewDec(int64(len(committee.GetMembers())))
	voteThreshold := committee.GetVoteThreshold()
	return currVotes, possibleVotes, voteThreshold
}

// GetTokenCommitteeProposalResult gets the result of a token committee proposal
func (k Keeper) GetTokenCommitteeProposalResult(ctx sdk.Context, proposalID uint64, committee types.TokenCommittee) bool {
	yesVotes, currVotes, possibleVotes, voteThreshold, quroum := k.TallyTokenCommitteeVotes(ctx, proposalID, committee)
	if currVotes.Quo(possibleVotes).GTE(quroum) { // quorum requirement
		if yesVotes.Quo(currVotes).GTE(voteThreshold) { // vote threshold requirements
			return true
		}
	}
	return false
}

// TallyMemberCommitteeVotes returns the polling status of a token committee vote. Returns yes votes,
// total current votes, total possible votes (equal to token supply), vote threshold (yes vote ratio
// required for proposal to pass), and quroum (votes tallied at this percentage).
func (k Keeper) TallyTokenCommitteeVotes(ctx sdk.Context, proposalID uint64,
	committee types.TokenCommittee) (sdk.Dec, sdk.Dec, sdk.Dec, sdk.Dec, sdk.Dec) {
	tallyDenom := committee.GetTallyDenom()
	votes := k.GetVotesByProposal(ctx, proposalID)

	currVotes := sdk.ZeroDec()
	yesVotes := sdk.ZeroDec()
	for _, vote := range votes {
		// 1 token = 1 vote
		acc := k.accountKeeper.GetAccount(ctx, vote.Voter)
		// TODO: what if account is nil?
		// if reflect.DeepEqual(acc, authexported.Account{}) {
		// 	continue
		// }
		accNumCoins := acc.GetCoins().AmountOf(tallyDenom)

		// Add votes to counters
		currVotes = currVotes.Add(accNumCoins.ToDec())
		if vote.VoteType == types.Yes {
			yesVotes = yesVotes.Add(accNumCoins.ToDec())
		}
	}

	possibleVotes := k.supplyKeeper.GetSupply(ctx).GetTotal().AmountOf(tallyDenom)
	return yesVotes, currVotes, possibleVotes.ToDec(), committee.GetVoteThreshold(), committee.GetQuorum()
}

// EnactProposal makes the changes proposed in a proposal.
func (k Keeper) EnactProposal(ctx sdk.Context, proposal types.Proposal) error {
	// Check committee still has permissions for the proposal
	// Since the proposal was submitted params could have changed, invalidating the permission of the committee.
	com, found := k.GetCommittee(ctx, proposal.CommitteeID)
	if !found {
		return sdkerrors.Wrapf(types.ErrUnknownCommittee, "%d", proposal.CommitteeID)
	}
	if !com.HasPermissionsFor(ctx, k.cdc, k.ParamKeeper, proposal.PubProposal) {
		return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "committee does not have permissions to enact proposal")
	}

	if err := k.ValidatePubProposal(ctx, proposal.PubProposal); err != nil {
		return err
	}

	// enact the proposal
	handler := k.router.GetRoute(proposal.ProposalRoute())
	if err := handler(ctx, proposal.PubProposal); err != nil {
		// the handler should not error as it was checked in ValidatePubProposal
		panic(fmt.Sprintf("unexpected handler error: %s", err))
	}
	return nil
}

// CloseProposal deletes proposals and their votes, emitting an event denoting the final status of the proposal
func (k Keeper) CloseProposal(ctx sdk.Context, proposal types.Proposal, outcome types.ProposalOutcome) {

	k.DeleteProposalAndVotes(ctx, proposal.ID)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposalClose,
			sdk.NewAttribute(types.AttributeKeyCommitteeID, fmt.Sprintf("%d", proposal.CommitteeID)),
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposal.ID)),
			sdk.NewAttribute(types.AttributeKeyProposalOutcome, outcome.String()),
		),
	)
}
