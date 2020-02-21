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
		case types.QueryGetAtomicSwap:
			return queryAtomicSwap(ctx, req, keeper)
		case types.QueryGetAtomicSwaps:
			return queryAtomicSwaps(ctx, req, keeper)
		case types.QueryGetParams:
			return queryGetParams(ctx, req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown bep3 query endpoint")
		}
	}
}

func queryAtomicSwap(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	// Decode request
	var requestParams types.QueryAtomicSwapByID
	err := keeper.cdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	// Lookup atomic swap
	atomicSwap, found := keeper.GetAtomicSwap(ctx, requestParams.SwapID)
	if !found {
		return nil, sdk.ErrInternal("Not found")
	}

	// Encode results
	bz, err := codec.MarshalJSONIndent(keeper.cdc, atomicSwap)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}

func queryAtomicSwaps(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	var swaps types.AtomicSwaps

	keeper.IterateAtomicSwaps(ctx, func(s types.AtomicSwap) bool {
		swaps = append(swaps, s)
		return false
	})

	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, swaps)
	if err2 != nil {
		return nil, sdk.ErrInternal("could not marshal result to JSON")
	}

	return bz, nil
}

// query params in the bep3 store
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
