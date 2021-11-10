package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/committee/types"
)

var _ types.QueryServer = QueryHandler{}

type QueryHandler struct {
	*Keeper
}

// Committees implements the gRPC service handler for querying committees.
func (k QueryHandler) Committees(ctx context.Context, req *types.QueryCommitteesRequest) (*types.QueryCommitteesResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	committees := k.GetCommittees(sdkCtx)
	committeesAny, err := types.PackCommittees(committees)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "could not pack committees: %v", err)
	}

	return &types.QueryCommitteesResponse{Committees: committeesAny}, nil
}

// Committee implements the Query/Committee gRPC method.
func (k QueryHandler) Committee(c context.Context, req *types.QueryCommitteeRequest) (*types.QueryCommitteeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	committee, found := k.GetCommittee(ctx, req.CommitteeID)
	if !found {
		return nil, status.Errorf(codes.NotFound, "could not find committee for id: %v", req.CommitteeID)
	}
	committeeAny, err := types.PackCommittee(committee)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not pack committees: %v", err)
	}
	return &types.QueryCommitteeResponse{Committee: committeeAny}, nil
}

// Proposals implements the Query/Proposals gRPC method
func (k QueryHandler) Proposals(c context.Context, req *types.QueryProposalsRequest) (*types.QueryProposalsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	proposals := k.GetProposalsByCommittee(ctx, req.CommitteeID)
	proposalsResp := types.QueryProposalsResponse{
		Proposals: make([]types.QueryProposalResponse, len(proposals)),
	}
	for i, proposal := range proposals {
		proposalsResp.Proposals[i] = k.proposalResponseFromProposal(proposal)
	}

	return &proposalsResp, nil
}

// Proposal implements the Query/Proposal gRPC method
func (k QueryHandler) Proposal(c context.Context, req *types.QueryProposalRequest) (*types.QueryProposalResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	proposal, found := k.GetProposal(ctx, req.ProposalID)
	if !found {
		return nil, status.Errorf(codes.NotFound, "cannot find proposal: %v", req.ProposalID)
	}
	proposalResp := k.proposalResponseFromProposal(proposal)
	return &proposalResp, nil
}

// NextProposalID implements the Query/NextProposalID gRPC method
func (k QueryHandler) NextProposalID(c context.Context, req *types.QueryNextProposalIDRequest) (*types.QueryNextProposalIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	proposalID, err := k.GetNextProposalID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "cannot find next proposal id: %v", err)
	}

	return &types.QueryNextProposalIDResponse{NextProposalID: proposalID}, nil
}

// Votes implements the Query/Votes gRPC method
func (k QueryHandler) Votes(c context.Context, req *types.QueryVotesRequest) (*types.QueryVotesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	votes := k.GetVotesByProposal(ctx, req.ProposalID)
	votesResp := types.QueryVotesResponse{
		Votes: make([]types.QueryVoteResponse, len(votes)),
	}
	for i, vote := range votes {
		votesResp.Votes[i] = k.votesResponseFromVote(vote)
	}
	return &votesResp, nil
}

// Vote implements the Query/Vote gRPC method
func (k QueryHandler) Vote(c context.Context, req *types.QueryVoteRequest) (*types.QueryVoteResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	voter, err := sdk.AccAddressFromBech32(req.Voter)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid voter address: %v", err)
	}
	vote, found := k.GetVote(ctx, req.ProposalID, voter)
	if !found {
		return nil, status.Errorf(codes.NotFound, "proposal id: %d, voter: %s", req.ProposalID, req.Voter)
	}
	voteResp := k.votesResponseFromVote(vote)
	return &voteResp, nil
}

// Tally implements the Query/Tally gRPC method
func (k QueryHandler) Tally(c context.Context, req *types.QueryTallyRequest) (*types.QueryTallyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	tally, found := k.GetProposalTallyResponse(ctx, req.ProposalID)
	if !found {
		return nil, status.Errorf(codes.NotFound, "proposal id: %d", req.ProposalID)
	}
	return tally, nil
}

// RawParams implements the Query/RawParams gRPC method
func (k QueryHandler) RawParams(c context.Context, req *types.QueryRawParamsRequest) (*types.QueryRawParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	subspace, found := k.paramKeeper.GetSubspace(req.Subspace)
	if !found {
		return nil, status.Errorf(codes.NotFound, "subspace not found: %s", req.Subspace)
	}
	rawParams := subspace.GetRaw(ctx, []byte(req.Key))
	return &types.QueryRawParamsResponse{RawData: string(rawParams)}, nil
}

func (k QueryHandler) proposalResponseFromProposal(proposal types.Proposal) types.QueryProposalResponse {
	return types.QueryProposalResponse{
		PubProposal: proposal.Content,
		ID:          proposal.ID,
		CommitteeID: proposal.CommitteeID,
		Deadline:    proposal.Deadline,
	}
}

func (k QueryHandler) votesResponseFromVote(vote types.Vote) types.QueryVoteResponse {
	return types.QueryVoteResponse{
		ProposalID: vote.ProposalID,
		Voter:      vote.Voter.String(),
		VoteType:   vote.VoteType,
	}
}
