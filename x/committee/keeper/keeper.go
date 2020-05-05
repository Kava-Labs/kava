package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/kava-labs/kava/x/committee/types"
)

type Keeper struct {
	cdc      *codec.Codec
	storeKey sdk.StoreKey

	// Proposal router
	router govtypes.Router
}

func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey, router govtypes.Router) Keeper {
	// Logic in the keeper methods assume the set of gov handlers is fixed.
	// So the gov router must be sealed so no handlers can be added or removed after the keeper is created.
	router.Seal()

	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,
		router:   router,
	}
}

// ------------------------------------------
//				Committees
// ------------------------------------------

// GetCommittee gets a committee from the store.
func (k Keeper) GetCommittee(ctx sdk.Context, committeeID uint64) (types.Committee, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.CommitteeKeyPrefix)
	bz := store.Get(types.GetKeyFromID(committeeID))
	if bz == nil {
		return types.Committee{}, false
	}
	var committee types.Committee
	k.cdc.MustUnmarshalBinaryBare(bz, &committee)
	return committee, true
}

// SetCommittee puts a committee into the store.
func (k Keeper) SetCommittee(ctx sdk.Context, committee types.Committee) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.CommitteeKeyPrefix)
	bz := k.cdc.MustMarshalBinaryBare(committee)
	store.Set(types.GetKeyFromID(committee.ID), bz)
}

// DeleteCommittee removes a committee from the store.
func (k Keeper) DeleteCommittee(ctx sdk.Context, committeeID uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.CommitteeKeyPrefix)
	store.Delete(types.GetKeyFromID(committeeID))
}

// IterateCommittees provides an iterator over all stored committees.
// For each committee, cb will be called. If cb returns true, the iterator will close and stop.
func (k Keeper) IterateCommittees(ctx sdk.Context, cb func(committee types.Committee) (stop bool)) {
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.CommitteeKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var committee types.Committee
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &committee)

		if cb(committee) {
			break
		}
	}
}

// GetCommittees returns all stored committees.
func (k Keeper) GetCommittees(ctx sdk.Context) []types.Committee {
	results := []types.Committee{}
	k.IterateCommittees(ctx, func(com types.Committee) bool {
		results = append(results, com)
		return false
	})
	return results
}

// ------------------------------------------
//				Proposals
// ------------------------------------------

// SetNextProposalID stores an ID to be used for the next created proposal
func (k Keeper) SetNextProposalID(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.NextProposalIDKey, types.GetKeyFromID(id))
}

// GetNextProposalID reads the next available global ID from store
func (k Keeper) GetNextProposalID(ctx sdk.Context) (uint64, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.NextProposalIDKey)
	if bz == nil {
		return 0, sdkerrors.Wrap(types.ErrInvalidGenesis, "next proposal ID not set at genesis")
	}
	return types.Uint64FromBytes(bz), nil
}

// IncrementNextProposalID increments the next proposal ID in the store by 1.
func (k Keeper) IncrementNextProposalID(ctx sdk.Context) error {
	id, err := k.GetNextProposalID(ctx)
	if err != nil {
		return err
	}
	k.SetNextProposalID(ctx, id+1)
	return nil
}

// StoreNewProposal stores a proposal, adding a new ID
func (k Keeper) StoreNewProposal(ctx sdk.Context, pubProposal types.PubProposal, committeeID uint64, deadline time.Time) (uint64, error) {
	newProposalID, err := k.GetNextProposalID(ctx)
	if err != nil {
		return 0, err
	}
	proposal := types.NewProposal(
		pubProposal,
		newProposalID,
		committeeID,
		deadline,
	)

	k.SetProposal(ctx, proposal)

	err = k.IncrementNextProposalID(ctx)
	if err != nil {
		return 0, err
	}
	return newProposalID, nil
}

// GetProposal gets a proposal from the store.
func (k Keeper) GetProposal(ctx sdk.Context, proposalID uint64) (types.Proposal, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ProposalKeyPrefix)
	bz := store.Get(types.GetKeyFromID(proposalID))
	if bz == nil {
		return types.Proposal{}, false
	}
	var proposal types.Proposal
	k.cdc.MustUnmarshalBinaryBare(bz, &proposal)
	return proposal, true
}

// SetProposal puts a proposal into the store.
func (k Keeper) SetProposal(ctx sdk.Context, proposal types.Proposal) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ProposalKeyPrefix)
	bz := k.cdc.MustMarshalBinaryBare(proposal)
	store.Set(types.GetKeyFromID(proposal.ID), bz)
}

// DeleteProposal removes a proposal from the store.
func (k Keeper) DeleteProposal(ctx sdk.Context, proposalID uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ProposalKeyPrefix)
	store.Delete(types.GetKeyFromID(proposalID))
}

// IterateProposals provides an iterator over all stored proposals.
// For each proposal, cb will be called. If cb returns true, the iterator will close and stop.
func (k Keeper) IterateProposals(ctx sdk.Context, cb func(proposal types.Proposal) (stop bool)) {
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.ProposalKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var proposal types.Proposal
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &proposal)

		if cb(proposal) {
			break
		}
	}
}

// GetProposals returns all stored proposals.
func (k Keeper) GetProposals(ctx sdk.Context) []types.Proposal {
	results := []types.Proposal{}
	k.IterateProposals(ctx, func(prop types.Proposal) bool {
		results = append(results, prop)
		return false
	})
	return results
}

// GetProposalsByCommittee returns all proposals for one committee.
func (k Keeper) GetProposalsByCommittee(ctx sdk.Context, committeeID uint64) []types.Proposal {
	results := []types.Proposal{}
	k.IterateProposals(ctx, func(prop types.Proposal) bool {
		if prop.CommitteeID == committeeID {
			results = append(results, prop)
		}
		return false
	})
	return results
}

// DeleteProposalAndVotes removes a proposal and its associated votes.
func (k Keeper) DeleteProposalAndVotes(ctx sdk.Context, proposalID uint64) {

	votes := k.GetVotesByProposal(ctx, proposalID)

	k.DeleteProposal(ctx, proposalID)
	for _, v := range votes {
		k.DeleteVote(ctx, v.ProposalID, v.Voter)
	}
}

// ------------------------------------------
//				Votes
// ------------------------------------------

// GetVote gets a vote from the store.
func (k Keeper) GetVote(ctx sdk.Context, proposalID uint64, voter sdk.AccAddress) (types.Vote, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.VoteKeyPrefix)
	bz := store.Get(types.GetVoteKey(proposalID, voter))
	if bz == nil {
		return types.Vote{}, false
	}
	var vote types.Vote
	k.cdc.MustUnmarshalBinaryBare(bz, &vote)
	return vote, true
}

// SetVote puts a vote into the store.
func (k Keeper) SetVote(ctx sdk.Context, vote types.Vote) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.VoteKeyPrefix)
	bz := k.cdc.MustMarshalBinaryBare(vote)
	store.Set(types.GetVoteKey(vote.ProposalID, vote.Voter), bz)
}

// DeleteVote removes a Vote from the store.
func (k Keeper) DeleteVote(ctx sdk.Context, proposalID uint64, voter sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.VoteKeyPrefix)
	store.Delete(types.GetVoteKey(proposalID, voter))
}

// IterateVotes provides an iterator over all stored votes.
// For each vote, cb will be called. If cb returns true, the iterator will close and stop.
func (k Keeper) IterateVotes(ctx sdk.Context, cb func(vote types.Vote) (stop bool)) {
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.VoteKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var vote types.Vote
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &vote)

		if cb(vote) {
			break
		}
	}
}

// GetVotes returns all stored votes.
func (k Keeper) GetVotes(ctx sdk.Context) []types.Vote {
	results := []types.Vote{}
	k.IterateVotes(ctx, func(vote types.Vote) bool {
		results = append(results, vote)
		return false
	})
	return results
}

// TODO: Might as well use a more specific iterator with the ProposalID as the prefix
// GetVotesByProposal returns all votes for one proposal.
func (k Keeper) GetVotesByProposal(ctx sdk.Context, proposalID uint64) []types.Vote {
	results := []types.Vote{}
	k.IterateVotes(ctx, func(vote types.Vote) bool {
		if vote.ProposalID == proposalID {
			results = append(results, vote)
		}
		return false
	})
	return results
}
