package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/bep3/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		// TODO: Get individual HTLT by ID
		// case types.QueryGetHTLT:
		// 	return queryHTLT(ctx, req, keeper)
		case types.QueryGetHTLTs:
			return queryHTLTs(ctx, req, keeper)
		case types.QueryGetParams:
			return queryGetParams(ctx, req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown bep3 query endpoint")
		}
	}
}

func queryHTLTs(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	var HTLTs types.HTLTs

	keeper.IterateHTLTs(ctx, func(h types.HTLT) bool {
		HTLTs = append(HTLTs, h)
		return false
	})

	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, HTLTs)
	if err2 != nil {
		return nil, sdk.ErrInternal("could not marshal result to JSON")
	}

	return bz, nil
}

// query params in the htlt store
func queryGetParams(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	// Get params
	params := keeper.GetParams(ctx)

	// Encode results
	bz, err := codec.MarshalJSONIndent(keeper.cdc, params)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}
