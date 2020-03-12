package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	// Note: for some reason the gov router panics if it has already been sealed, so a helper func is used to make sealing idempotent.
	sealGovRouterIdempotently(router)

	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,
		router:   router,
	}
}

func sealGovRouterIdempotently(router govtypes.Router) {
	defer func() {
		recover()
	}()
	router.Seal()
}

// ---------- Committees ----------

// GetCommittee gets a committee from the store.
func (k Keeper) GetCommittee(ctx sdk.Context, committeeID uint64) (types.Committee, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.CommitteeKeyPrefix)
	bz := store.Get(types.GetKeyFromID(committeeID))
	if bz == nil {
		return types.Committee{}, false
	}
	var committee types.Committee
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &committee)
	return committee, true
}

// SetCommittee puts a committee into the store.
func (k Keeper) SetCommittee(ctx sdk.Context, committee types.Committee) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.CommitteeKeyPrefix)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(committee)
	store.Set(types.GetKeyFromID(committee.ID), bz)
}

// DeleteCommittee removes a committee from the store.
func (k Keeper) DeleteCommittee(ctx sdk.Context, committeeID uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.CommitteeKeyPrefix)
	store.Delete(types.GetKeyFromID(committeeID))
}

// ---------- Proposals ----------

// SetNextProposalID stores an ID to be used for the next created proposal
func (k Keeper) SetNextProposalID(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.NextProposalIDKey, types.GetKeyFromID(id))
}

// GetNextProposalID reads the next available global ID from store
func (k Keeper) GetNextProposalID(ctx sdk.Context) (uint64, sdk.Error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.NextProposalIDKey)
	if bz == nil {
		return 0, sdk.ErrInternal("proposal ID not set at genesis")
	}
	return types.Uint64FromBytes(bz), nil
}

// IncrementNextProposalID increments the next proposal ID in the store by 1.
func (k Keeper) IncrementNextProposalID(ctx sdk.Context) sdk.Error {
	id, err := k.GetNextProposalID(ctx)
	if err != nil {
		return err
	}
	k.SetNextProposalID(ctx, id+1)
	return nil
}

// StoreNewProposal stores a proposal, adding a new ID
func (k Keeper) StoreNewProposal(ctx sdk.Context, pubProposal types.PubProposal, committeeID uint64, deadline time.Time) (uint64, sdk.Error) {
	newProposalID, err := k.GetNextProposalID(ctx)
	if err != nil {
		return 0, err
	}
	proposal := types.Proposal{
		PubProposal: pubProposal,
		ID:          newProposalID,
		CommitteeID: committeeID,
		Deadline:    deadline,
	}

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
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &proposal)
	return proposal, true
}

// SetProposal puts a proposal into the store.
func (k Keeper) SetProposal(ctx sdk.Context, proposal types.Proposal) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ProposalKeyPrefix)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(proposal)
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
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &proposal)

		if cb(proposal) {
			break
		}
	}
}

// ---------- Votes ----------

// GetVote gets a vote from the store.
func (k Keeper) GetVote(ctx sdk.Context, proposalID uint64, voter sdk.AccAddress) (types.Vote, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.VoteKeyPrefix)
	bz := store.Get(types.GetVoteKey(proposalID, voter))
	if bz == nil {
		return types.Vote{}, false
	}
	var vote types.Vote
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &vote)
	return vote, true
}

// SetVote puts a vote into the store.
func (k Keeper) SetVote(ctx sdk.Context, vote types.Vote) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.VoteKeyPrefix)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(vote)
	store.Set(types.GetVoteKey(vote.ProposalID, vote.Voter), bz)
}

// DeleteVote removes a Vote from the store.
func (k Keeper) DeleteVote(ctx sdk.Context, proposalID uint64, voter sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.VoteKeyPrefix)
	store.Delete(types.GetVoteKey(proposalID, voter))
}

// IterateVotes provides an iterator over all stored votes for a given proposal.
// For each vote, cb will be called. If cb returns true, the iterator will close and stop.
func (k Keeper) IterateVotes(ctx sdk.Context, proposalID uint64, cb func(vote types.Vote) (stop bool)) {
	// iterate over the section of the votes store that has all votes for a particular proposal
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), append(types.VoteKeyPrefix, types.GetKeyFromID(proposalID)...))

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var vote types.Vote
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &vote)

		if cb(vote) {
			break
		}
	}
}
