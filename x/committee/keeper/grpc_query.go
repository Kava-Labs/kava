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

type queryServer struct {
	keeper Keeper
}

// NewQueryServerImpl creates a new server for handling gRPC queries.
func NewQueryServerImpl(k Keeper) types.QueryServer {
	return &queryServer{keeper: k}
}

var _ types.QueryServer = queryServer{}

// Committees implements the gRPC service handler for querying committees.
func (s queryServer) Committees(ctx context.Context, req *types.QueryCommitteesRequest) (*types.QueryCommitteesResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	committees := s.keeper.GetCommittees(sdkCtx)
	committeesAny, err := types.PackCommittees(committees)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "could not pack committees: %v", err)
	}

	return &types.QueryCommitteesResponse{Committees: committeesAny}, nil
}

// Committee implements the Query/Committee gRPC method.
func (s queryServer) Committee(c context.Context, req *types.QueryCommitteeRequest) (*types.QueryCommitteeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	committee, found := s.keeper.GetCommittee(ctx, req.CommitteeId)
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
func (s queryServer) Proposals(c context.Context, req *types.QueryProposalsRequest) (*types.QueryProposalsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	proposals := s.keeper.GetProposalsByCommittee(ctx, req.CommitteeId)
	var proposalsResp []types.QueryProposalResponse

	for _, proposal := range proposals {
		proposalsResp = append(proposalsResp, s.proposalResponseFromProposal(proposal))
	}

	return &types.QueryProposalsResponse{
		Proposals: proposalsResp,
	}, nil
}

// Proposal implements the Query/Proposal gRPC method
func (s queryServer) Proposal(c context.Context, req *types.QueryProposalRequest) (*types.QueryProposalResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	proposal, found := s.keeper.GetProposal(ctx, req.ProposalId)
	if !found {
		return nil, status.Errorf(codes.NotFound, "cannot find proposal: %v", req.ProposalId)
	}
	proposalResp := s.proposalResponseFromProposal(proposal)
	return &proposalResp, nil
}

// NextProposalID implements the Query/NextProposalID gRPC method
func (s queryServer) NextProposalID(c context.Context, req *types.QueryNextProposalIDRequest) (*types.QueryNextProposalIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	proposalID, err := s.keeper.GetNextProposalID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "cannot find next proposal id: %v", err)
	}

	return &types.QueryNextProposalIDResponse{NextProposalID: proposalID}, nil
}

// Votes implements the Query/Votes gRPC method
func (s queryServer) Votes(c context.Context, req *types.QueryVotesRequest) (*types.QueryVotesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	var queryResults []types.QueryVoteResponse
	store := ctx.KVStore(s.keeper.storeKey)
	votesStore := prefix.NewStore(store, append(types.VoteKeyPrefix, types.GetKeyFromID(req.ProposalId)...))
	pageRes, err := query.Paginate(votesStore, req.Pagination, func(key []byte, value []byte) error {
		var vote types.Vote
		if err := s.keeper.cdc.Unmarshal(value, &vote); err != nil {
			return err
		}

		queryResults = append(queryResults, s.votesResponseFromVote(vote))
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
func (s queryServer) Vote(c context.Context, req *types.QueryVoteRequest) (*types.QueryVoteResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	voter, err := sdk.AccAddressFromBech32(req.Voter)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid voter address: %v", err)
	}
	vote, found := s.keeper.GetVote(ctx, req.ProposalId, voter)
	if !found {
		return nil, status.Errorf(codes.NotFound, "proposal id: %d, voter: %s", req.ProposalId, req.Voter)
	}
	voteResp := s.votesResponseFromVote(vote)
	return &voteResp, nil
}

// Tally implements the Query/Tally gRPC method
func (s queryServer) Tally(c context.Context, req *types.QueryTallyRequest) (*types.QueryTallyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	tally, found := s.keeper.GetProposalTallyResponse(ctx, req.ProposalId)
	if !found {
		return nil, status.Errorf(codes.NotFound, "proposal id: %d", req.ProposalId)
	}
	return tally, nil
}

// RawParams implements the Query/RawParams gRPC method
func (s queryServer) RawParams(c context.Context, req *types.QueryRawParamsRequest) (*types.QueryRawParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	subspace, found := s.keeper.paramKeeper.GetSubspace(req.Subspace)
	if !found {
		return nil, status.Errorf(codes.NotFound, "subspace not found: %s", req.Subspace)
	}
	rawParams := subspace.GetRaw(ctx, []byte(req.Key))
	return &types.QueryRawParamsResponse{RawData: string(rawParams)}, nil
}

func (s queryServer) proposalResponseFromProposal(proposal types.Proposal) types.QueryProposalResponse {
	return types.QueryProposalResponse{
		PubProposal: proposal.Content,
		ID:          proposal.ID,
		CommitteeID: proposal.CommitteeID,
		Deadline:    proposal.Deadline,
	}
}

func (s queryServer) votesResponseFromVote(vote types.Vote) types.QueryVoteResponse {
	return types.QueryVoteResponse{
		ProposalID: vote.ProposalID,
		Voter:      vote.Voter.String(),
		VoteType:   vote.VoteType,
	}
}
