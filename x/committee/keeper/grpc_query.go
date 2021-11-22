package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/kava-labs/kava/x/committee/types"
)

var _ types.QueryServer = QueryHandler{}

type QueryHandler struct {
	keeper *Keeper
}

// NewQueryHandler returns a new QueryHandler instance
func NewQueryHandler(k *Keeper) QueryHandler {
	return QueryHandler{keeper: k}
}

// Committees implements the gRPC service handler for querying committees.
func (q QueryHandler) Committees(ctx context.Context, req *types.QueryCommitteesRequest) (*types.QueryCommitteesResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	committees := q.keeper.GetCommittees(sdkCtx)
	committeesAny, err := types.PackCommittees(committees)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "could not pack committees: %v", err)
	}

	return &types.QueryCommitteesResponse{Committees: committeesAny}, nil
}

// Committee implements the Query/Committee gRPC method.
func (q QueryHandler) Committee(c context.Context, req *types.QueryCommitteeRequest) (*types.QueryCommitteeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	committee, found := q.keeper.GetCommittee(ctx, req.CommitteeId)
	if !found {
		return nil, status.Errorf(codes.NotFound, "could not find committee for id: %v", req.CommitteeId)
	}
	committeeAny, err := types.PackCommittee(committee)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not pack committees: %v", err)
	}
	return &types.QueryCommitteeResponse{Committee: committeeAny}, nil
}

// Proposals implements the Query/Proposals gRPC method
func (q QueryHandler) Proposals(c context.Context, req *types.QueryProposalsRequest) (*types.QueryProposalsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	proposals := q.keeper.GetProposalsByCommittee(ctx, req.CommitteeId)
	proposalsResp := types.QueryProposalsResponse{
		Proposals: make([]types.QueryProposalResponse, len(proposals)),
	}
	for i, proposal := range proposals {
		proposalsResp.Proposals[i] = q.proposalResponseFromProposal(proposal)
	}

	return &proposalsResp, nil
}

// Proposal implements the Query/Proposal gRPC method
func (q QueryHandler) Proposal(c context.Context, req *types.QueryProposalRequest) (*types.QueryProposalResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	proposal, found := q.keeper.GetProposal(ctx, req.ProposalId)
	if !found {
		return nil, status.Errorf(codes.NotFound, "cannot find proposal: %v", req.ProposalId)
	}
	proposalResp := q.proposalResponseFromProposal(proposal)
	return &proposalResp, nil
}

// NextProposalID implements the Query/NextProposalID gRPC method
func (q QueryHandler) NextProposalID(c context.Context, req *types.QueryNextProposalIDRequest) (*types.QueryNextProposalIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	proposalID, err := q.keeper.GetNextProposalID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "cannot find next proposal id: %v", err)
	}

	return &types.QueryNextProposalIDResponse{NextProposalID: proposalID}, nil
}

// Votes implements the Query/Votes gRPC method
func (q QueryHandler) Votes(c context.Context, req *types.QueryVotesRequest) (*types.QueryVotesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	votes := q.keeper.GetVotesByProposal(ctx, req.ProposalId)
	votesResp := types.QueryVotesResponse{
		Votes: make([]types.QueryVoteResponse, len(votes)),
	}
	for i, vote := range votes {
		votesResp.Votes[i] = q.votesResponseFromVote(vote)
	}

	var queryResults []types.QueryVoteResponse
	store := ctx.KVStore(q.keeper.storeKey)
	votesStore := prefix.NewStore(store, append(types.VoteKeyPrefix, types.GetKeyFromID(req.ProposalId)...))
	pageRes, err := query.Paginate(votesStore, req.Pagination, func(key []byte, value []byte) error {
		var vote types.Vote
		if err := q.keeper.cdc.Unmarshal(value, &vote); err != nil {
			return err
		}

		votes = append(votes, vote)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryVotesResponse{
		Votes:      queryResults,
		Pagination: pageRes,
	}, nil
}

// Vote implements the Query/Vote gRPC method
func (q QueryHandler) Vote(c context.Context, req *types.QueryVoteRequest) (*types.QueryVoteResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	voter, err := sdk.AccAddressFromBech32(req.Voter)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid voter address: %v", err)
	}
	vote, found := q.keeper.GetVote(ctx, req.ProposalId, voter)
	if !found {
		return nil, status.Errorf(codes.NotFound, "proposal id: %d, voter: %s", req.ProposalId, req.Voter)
	}
	voteResp := q.votesResponseFromVote(vote)
	return &voteResp, nil
}

// Tally implements the Query/Tally gRPC method
func (q QueryHandler) Tally(c context.Context, req *types.QueryTallyRequest) (*types.QueryTallyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	tally, found := q.keeper.GetProposalTallyResponse(ctx, req.ProposalId)
	if !found {
		return nil, status.Errorf(codes.NotFound, "proposal id: %d", req.ProposalId)
	}
	return tally, nil
}

// RawParams implements the Query/RawParams gRPC method
func (q QueryHandler) RawParams(c context.Context, req *types.QueryRawParamsRequest) (*types.QueryRawParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	subspace, found := q.keeper.paramKeeper.GetSubspace(req.Subspace)
	if !found {
		return nil, status.Errorf(codes.NotFound, "subspace not found: %s", req.Subspace)
	}
	rawParams := subspace.GetRaw(ctx, []byte(req.Key))
	return &types.QueryRawParamsResponse{RawData: string(rawParams)}, nil
}

func (q QueryHandler) proposalResponseFromProposal(proposal types.Proposal) types.QueryProposalResponse {
	return types.QueryProposalResponse{
		PubProposal: proposal.Content,
		ID:          proposal.ID,
		CommitteeID: proposal.CommitteeID,
		Deadline:    proposal.Deadline,
	}
}

func (q QueryHandler) votesResponseFromVote(vote types.Vote) types.QueryVoteResponse {
	return types.QueryVoteResponse{
		ProposalID: vote.ProposalID,
		Voter:      vote.Voter.String(),
		VoteType:   vote.VoteType,
	}
}
