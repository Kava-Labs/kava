package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/x/committee/types"
)

// NewQuerier creates a new gov Querier instance
func NewQuerier(k Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QueryCommittees:
			return queryCommittees(ctx, req, k, legacyQuerierCdc)
		case types.QueryCommittee:
			return queryCommittee(ctx, req, k, legacyQuerierCdc)
		case types.QueryProposals:
			return queryProposals(ctx, req, k, legacyQuerierCdc)
		case types.QueryProposal:
			return queryProposal(ctx, req, k, legacyQuerierCdc)
		case types.QueryVotes:
			return queryVotes(ctx, req, k, legacyQuerierCdc)
		case types.QueryVote:
			return queryVote(ctx, req, k, legacyQuerierCdc)
		case types.QueryTally:
			return queryTally(ctx, req, k, legacyQuerierCdc)
		case types.QueryNextProposalID:
			return queryNextProposalID(ctx, req, k, legacyQuerierCdc)
		case types.QueryRawParams:
			return queryRawParams(ctx, req, k, legacyQuerierCdc)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint", types.ModuleName)
		}
	}
}

// ------------------------------------------
//				Committees
// ------------------------------------------

func queryCommittees(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {

	committees := keeper.GetCommittees(ctx)

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, committees)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

func queryCommittee(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryCommitteeParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	committee, found := keeper.GetCommittee(ctx, params.CommitteeID)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUnknownCommittee, "%d", params.CommitteeID)
	}

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, committee)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

// ------------------------------------------
//				Proposals
// ------------------------------------------

func queryProposals(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryCommitteeParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	proposals := keeper.GetProposalsByCommittee(ctx, params.CommitteeID)

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, proposals)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

func queryProposal(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryProposalParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	proposal, found := keeper.GetProposal(ctx, params.ProposalID)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUnknownProposal, "%d", params.ProposalID)
	}

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, proposal)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

func queryNextProposalID(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	nextProposalID, _ := k.GetNextProposalID(ctx)

	bz, err := types.ModuleCdc.LegacyAmino.MarshalJSON(nextProposalID)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	return bz, nil
}

// ------------------------------------------
//				Votes
// ------------------------------------------

func queryVotes(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryProposalParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)

	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	votes := keeper.GetVotesByProposal(ctx, params.ProposalID)

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, votes)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

func queryVote(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryVoteParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	vote, found := keeper.GetVote(ctx, params.ProposalID, params.Voter)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUnknownVote, "proposal id: %d, voter: %s", params.ProposalID, params.Voter)
	}

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, vote)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

// ------------------------------------------
//				Tally
// ------------------------------------------

func queryTally(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryProposalParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	tally, found := keeper.GetProposalTallyResponse(ctx, params.ProposalID)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrNotFoundProposalTally, "proposal id: %d", params.ProposalID)
	}

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, tally)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

// ------------------------------------------
//				Raw Params
// ------------------------------------------

func queryRawParams(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryRawParamsParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	subspace, found := keeper.paramKeeper.GetSubspace(params.Subspace)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUnknownSubspace, "subspace: %s", params.Subspace)
	}
	rawParams := subspace.GetRaw(ctx, []byte(params.Key))

	// encode the raw params as json, which converts them to a base64 string
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, rawParams)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}
