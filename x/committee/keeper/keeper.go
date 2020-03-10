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

/* TODO keeper methods - very similar to gov

- SubmitProposal validate and store a proposal, additionally setting things like timeout
- GetProposal
- SetProposal

- AddVote - add a vote to a particular proposal from a member
- GetVote
- SetVote

- GetCommittee
- SetCommittee

*/

func (k Keeper) SubmitProposal(ctx sdk.Context, proposal types.Proposal) sdk.Error {
	// TODO Limit proposals to only be submitted by group members

	// Check group has permissions to enact proposal. As long as one permission allows the proposal then it goes through. Its the OR of all permissions.
	committee, _ := k.GetCommittee(ctx, proposal.CommitteeID)
	hasPermissions := false
	for _, p := range committee.Permissions {
		if p.Allows(proposal) {
			hasPermissions = true
			break
		}
	}
	if !hasPermissions {
		return sdk.ErrInternal("committee does not have permissions to enact proposal")
	}

	// TODO validate proposal by running it with cached context like how gov does it

	// TODO store the proposal, probably put it in a queue

	return nil
}

func (k Keeper) AddVote(ctx sdk.Context, msg types.MsgVote) sdk.Error {
	/* TODO
	- validate vote
	- store vote
	*/
	return nil
}

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

func (k Keeper) SetCommittee(ctx sdk.Context, committee types.Committee) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.CommitteeKeyPrefix)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(committee)
	store.Set(types.GetKeyFromID(committee.ID), bz)
}

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

func (k Keeper) SetProposal(ctx sdk.Context, proposal types.Proposal) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ProposalKeyPrefix)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(proposal)
	store.Set(types.GetKeyFromID(proposal.ID), bz)
}

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

func (k Keeper) SetVote(ctx sdk.Context, vote types.Vote) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.VoteKeyPrefix)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(vote)
	store.Set(types.GetVoteKey(vote.ProposalID, vote.Voter), bz)
}
