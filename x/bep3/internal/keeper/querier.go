package keeper

import (
	"github.com/Kava-Labs/kava/x/bep3/internal/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		// TODO: Get individual KHTLT by ID
		// case types.QueryGetKHTLT:
		// 	return queryKHTLT(ctx, req, keeper)
		case types.QueryGetKHTLTs:
			return queryKHTLTs(ctx, req, keeper)
		case types.QueryGetParams:
			return queryGetParams(ctx, req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown bep3 query endpoint")
		}
	}
}

func queryKHTLTs(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	var KHTLTs types.KHTLTs

	keeper.IterateKHTLTs(ctx, func(h types.KHTLT) bool {
		KHTLTs = append(KHTLTs, h)
		return false
	})

	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, KHTLTs)
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
