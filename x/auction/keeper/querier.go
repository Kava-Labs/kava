package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

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
			return nil, sdk.ErrUnknownRequest("unknown auction query endpoint")
		}
	}
}

func queryAuction(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	// Decode request
	var requestParams types.QueryAuctionParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	// Lookup auction
	auction, found := keeper.GetAuction(ctx, requestParams.AuctionID)
	if !found {
		return nil, types.ErrAuctionNotFound(types.DefaultCodespace, requestParams.AuctionID)
	}

	// Encode results
	bz, err := codec.MarshalJSONIndent(keeper.cdc, auction)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}

func queryAuctions(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	// Get all auctions
	auctionsList := types.Auctions{}
	keeper.IterateAuctions(ctx, func(a types.Auction) bool {
		auctionsList = append(auctionsList, a)
		return false
	})

	// Encode Results
	bz, err := codec.MarshalJSONIndent(keeper.cdc, auctionsList)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
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
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}
