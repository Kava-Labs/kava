package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/bep3/types"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QueryGetAssetSupply:
			return queryAssetSupply(ctx, req, keeper)
		case types.QueryGetAssetSupplies:
			return queryAssetSupplies(ctx, req, keeper)
		case types.QueryGetAtomicSwap:
			return queryAtomicSwap(ctx, req, keeper)
		case types.QueryGetAtomicSwaps:
			return queryAtomicSwaps(ctx, req, keeper)
		case types.QueryGetParams:
			return queryGetParams(ctx, req, keeper)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint", types.ModuleName)
		}
	}
}

func queryAssetSupply(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	// Decode request
	var requestParams types.QueryAssetSupply
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	assetSupply, found := keeper.GetAssetSupply(ctx, []byte(requestParams.Denom))
	if !found {
		return nil, sdkerrors.Wrap(types.ErrAssetSupplyNotFound, string(requestParams.Denom))
	}

	// Encode results
	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, assetSupply)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryAssetSupplies(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) (res []byte, err error) {
	var assets types.AssetSupplies

	keeper.IterateAssetSupplies(ctx, func(a types.AssetSupply) bool {
		assets = append(assets, a)
		return false
	})

	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, assets)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryAtomicSwap(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	// Decode request
	var requestParams types.QueryAtomicSwapByID
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	// Lookup atomic swap
	atomicSwap, found := keeper.GetAtomicSwap(ctx, requestParams.SwapID)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrAtomicSwapNotFound, "%d", requestParams.SwapID)
	}

	// Encode results
	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, atomicSwap)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryAtomicSwaps(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	var params types.QueryAtomicSwaps
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	swaps := keeper.GetAtomicSwapsFiltered(ctx, params)
	if swaps == nil {
		swaps = types.AtomicSwaps{}
	}

	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, swaps)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

// query params in the bep3 store
func queryGetParams(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	// Get params
	params := keeper.GetParams(ctx)

	// Encode results
	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}
