package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/x/committee/types"
)

// NewQuerier creates a new gov Querier instance
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {

		case types.QueryCommittees:
			return queryCommittees(ctx, path[1:], req, keeper)
		case types.QueryCommittee:
			return queryCommittee(ctx, path[1:], req, keeper)
		case types.QueryProposals:
			return queryProposals(ctx, path[1:], req, keeper)
		case types.QueryProposal:
			return queryProposal(ctx, path[1:], req, keeper)
		case types.QueryVotes:
			return queryVotes(ctx, path[1:], req, keeper)
		case types.QueryVote:
			return queryVote(ctx, path[1:], req, keeper)
		case types.QueryTally:
			return queryTally(ctx, path[1:], req, keeper)

		default:
			return nil, sdk.ErrUnknownRequest(fmt.Sprintf("unknown %s query endpoint", types.ModuleName))
		}
	}
}

// ------------------------------------------
//				Committees
// ------------------------------------------

func queryCommittees(ctx sdk.Context, path []string, _ abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {

	committees := []types.Committee{}
	keeper.IterateCommittees(ctx, func(com types.Committee) bool {
		committees = append(committees, com)
		return false
	})

	bz, err := codec.MarshalJSONIndent(keeper.cdc, committees)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

func queryCommittee(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params types.QueryCommitteeParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	committee, found := keeper.GetCommittee(ctx, params.CommitteeID)
	if !found {
		return nil, types.ErrUnknownCommittee(types.DefaultCodespace, params.CommitteeID)
	}

	bz, err := codec.MarshalJSONIndent(keeper.cdc, committee)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

// ------------------------------------------
//				Proposals
// ------------------------------------------

func queryProposals(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params types.QueryCommitteeParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	proposals := []types.Proposal{}
	keeper.IterateProposals(ctx, func(p types.Proposal) bool {
		if p.CommitteeID == params.CommitteeID {
			proposals = append(proposals, p)
		}
		return false
	})

	bz, err := codec.MarshalJSONIndent(keeper.cdc, proposals)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

func queryProposal(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params types.QueryProposalParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	proposal, found := keeper.GetProposal(ctx, params.ProposalID)
	if !found {
		return nil, types.ErrUnknownProposal(types.DefaultCodespace, params.ProposalID)
	}

	bz, err := codec.MarshalJSONIndent(keeper.cdc, proposal)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

// ------------------------------------------
//				Votes
// ------------------------------------------

func queryVotes(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params types.QueryProposalParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &params)

	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	votes := []types.Vote{}
	keeper.IterateVotes(ctx, params.ProposalID, func(v types.Vote) bool {
		votes = append(votes, v)
		return false
	})

	bz, err := codec.MarshalJSONIndent(keeper.cdc, votes)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

func queryVote(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params types.QueryVoteParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	vote, found := keeper.GetVote(ctx, params.ProposalID, params.Voter)
	if !found {
		return nil, types.ErrUnknownVote(types.DefaultCodespace, params.ProposalID, params.Voter)
	}

	bz, err := codec.MarshalJSONIndent(keeper.cdc, vote)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

// ------------------------------------------
//				Tally
// ------------------------------------------

func queryTally(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params types.QueryProposalParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	_, found := keeper.GetProposal(ctx, params.ProposalID)
	if !found {
		return nil, types.ErrUnknownProposal(types.DefaultCodespace, params.ProposalID)
	}
	numVotes := keeper.TallyVotes(ctx, params.ProposalID)

	bz, err := codec.MarshalJSONIndent(keeper.cdc, numVotes)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}
