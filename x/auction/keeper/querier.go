package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/auction/types"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QueryGetAuction:
			return queryAuction(ctx, req, keeper)
		case types.QueryGetAuctions:
			return queryAuctions(ctx, req, keeper)
		case types.QueryGetParams:
			return queryGetParams(ctx, req, keeper)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint", types.ModuleName)
		}
	}
}

func queryAuction(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	// Decode request
	var requestParams types.QueryAuctionParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	// Lookup auction
	auction, found := keeper.GetAuction(ctx, requestParams.AuctionID)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrAuctionNotFound, "%d", requestParams.AuctionID)
	}

	// Encode results
	bz, err := codec.MarshalJSONIndent(keeper.cdc, auction)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryAuctions(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	var params types.QueryAllAuctionParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	unfilteredAuctions := keeper.GetAllAuctions(ctx)
	auctions := filterAuctions(ctx, unfilteredAuctions, params)
	if auctions == nil {
		auctions = types.Auctions{}
	}

	bz, err := codec.MarshalJSONIndent(keeper.cdc, auctions)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

// query params in the auction store
func queryGetParams(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	// Get params
	params := keeper.GetParams(ctx)

	// Encode results
	bz, err := codec.MarshalJSONIndent(keeper.cdc, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

// filterAuctions retrieves auctions filtered by a given set of params.
// If no filters are provided, all auctions will be returned in paginated form.
func filterAuctions(ctx sdk.Context, auctions types.Auctions, params types.QueryAllAuctionParams) types.Auctions {
	filteredAuctions := make(types.Auctions, 0, len(auctions))

	for _, auc := range auctions {
		matchType, matchDenom, matchPhase := true, true, true

		// match auction type (if supplied)
		if len(params.Type) > 0 {
			matchType = auc.GetType() == params.Type
		}

		// match auction denom (if supplied)
		if len(params.Denom) > 0 {
			matchDenom = auc.GetBid().Denom == params.Denom || auc.GetLot().Denom == params.Denom
		}

		// match auction phase (if supplied)
		if len(params.Phase) > 0 {
			matchPhase = auc.GetPhase() == params.Phase
		}

		if matchType && matchDenom && matchPhase {
			filteredAuctions = append(filteredAuctions, auc)
		}
	}

	start, end := client.Paginate(len(filteredAuctions), params.Page, params.Limit, 100)
	if start < 0 || end < 0 {
		filteredAuctions = types.Auctions{}
	} else {
		filteredAuctions = filteredAuctions[start:end]
	}

	return filteredAuctions
}
