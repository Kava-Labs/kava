package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	//govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/kava-labs/kava/x/committee/types"
)

type Keeper struct {
	cdc      *codec.Codec
	storeKey sdk.StoreKey

	// TODO Proposal router
	//router govtypes.Router
}

func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey) Keeper {
	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,
	}
}

func (k Keeper) SubmitProposal(ctx sdk.Context, proposer sdk.AccAddress, proposal types.Proposal) (uint64, sdk.Error) {
	// Limit proposals to only be submitted by committee members
	com, found := k.GetCommittee(ctx, proposal.CommitteeID)
	if !found {
		return 0, sdk.ErrInternal("committee doesn't exist")
	}
	if !com.HasMember(proposer) {
		return 0, sdk.ErrInternal("only member can propose proposals")
	}

	// Check proposal is valid
	if err := proposal.ValidateBasic(); err != nil {
		return 0, err
	}

	// Check group has permissions to enact proposal.
	if !com.HasPermissionsFor(proposal) {
		return 0, sdk.ErrInternal("committee does not have permissions to enact proposal")
	}

	// TODO validate proposal by running it with cached context like how gov does it
	// what if it's not valid now but will be in the future?

	// Get a new ID and store the proposal
	return k.StoreNewProposal(ctx, proposal)
}

func (k Keeper) AddVote(ctx sdk.Context, proposalID uint64, voter sdk.AccAddress) sdk.Error {
	// Validate
	proposal, found := k.GetProposal(ctx, proposalID)
	if !found {
		return sdk.ErrInternal("proposal not found")
	}
	com, found := k.GetCommittee(ctx, proposal.CommitteeID)
	if !found {
		return sdk.ErrInternal("committee disbanded")
	}
	if !com.HasMember(voter) {
		return sdk.ErrInternal("not authorized to vote on proposal")
	}

	// Store vote, overwriting any prior vote
	k.SetVote(ctx, types.Vote{ProposalID: proposalID, Voter: voter})

	// TODO close vote if tally has been reached
	return nil
}

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
func (k Keeper) StoreNewProposal(ctx sdk.Context, proposal types.Proposal) (uint64, sdk.Error) {
	newProposalID, err := k.GetNextProposalID(ctx)
	if err != nil {
		return 0, err
	}
	proposal.ID = newProposalID

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
