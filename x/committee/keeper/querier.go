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
		// case types.QueryParams:
		// 	return queryParams(ctx, path[1:], req, keeper)

		default:
			return nil, sdk.ErrUnknownRequest(fmt.Sprintf("unknown %s query endpoint", types.ModuleName))
		}
	}
}

// ---------- Committees ----------

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
		return nil, sdk.ErrInternal("not found") ///types.ErrUnknownProposal(types.DefaultCodespace, params.ProposalID)
	}

	bz, err := codec.MarshalJSONIndent(keeper.cdc, committee)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

// ---------- Proposals ----------

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
		return nil, sdk.ErrInternal("not found") // TODO types.ErrUnknownProposal(types.DefaultCodespace, params.ProposalID)
	}

	bz, err := codec.MarshalJSONIndent(keeper.cdc, proposal)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

// ---------- Votes ----------

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
		return nil, sdk.ErrInternal("not found")
	}

	bz, err := codec.MarshalJSONIndent(keeper.cdc, vote)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

// ---------- Tally ----------

func queryTally(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params types.QueryProposalParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	// TODO split tally and process result logic so tally logic can be used here
	pr, found := keeper.GetProposal(ctx, params.ProposalID)
	if !found {
		return nil, sdk.ErrInternal("proposal not found")
	}
	com, found := keeper.GetCommittee(ctx, pr.CommitteeID)
	if !found {
		return nil, sdk.ErrInternal("committee disbanded")
	}
	votes := []types.Vote{}
	keeper.IterateVotes(ctx, params.ProposalID, func(vote types.Vote) bool {
		votes = append(votes, vote)
		return false
	})
	proposalPasses := sdk.NewDec(int64(len(votes))).GTE(types.VoteThreshold.MulInt64(int64(len(com.Members))))
	// TODO return some kind of tally object, rather than just a bool

	bz, err := codec.MarshalJSONIndent(keeper.cdc, proposalPasses)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

// ---------- Params ----------

// func queryParams(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
// 	switch path[0] {
// 	case types.ParamDeposit:
// 		bz, err := codec.MarshalJSONIndent(keeper.cdc, keeper.GetDepositParams(ctx))
// 		if err != nil {
// 			return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
// 		}
// 		return bz, nil
// 	case types.ParamVoting:
// 		bz, err := codec.MarshalJSONIndent(keeper.cdc, keeper.GetVotingParams(ctx))
// 		if err != nil {
// 			return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
// 		}
// 		return bz, nil
// 	case types.ParamTallying:
// 		bz, err := codec.MarshalJSONIndent(keeper.cdc, keeper.GetTallyParams(ctx))
// 		if err != nil {
// 			return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
// 		}
// 		return bz, nil
// 	default:
// 		return nil, sdk.ErrUnknownRequest(fmt.Sprintf("%s is not a valid query request path", req.Path))
// 	}
// }
