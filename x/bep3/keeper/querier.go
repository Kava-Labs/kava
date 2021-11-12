package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/bep3/types"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QueryGetAssetSupply:
			return queryAssetSupply(ctx, req, keeper, legacyQuerierCdc)
		case types.QueryGetAssetSupplies:
			return queryAssetSupplies(ctx, req, keeper, legacyQuerierCdc)
		case types.QueryGetAtomicSwap:
			return queryAtomicSwap(ctx, req, keeper, legacyQuerierCdc)
		case types.QueryGetAtomicSwaps:
			return queryAtomicSwaps(ctx, req, keeper, legacyQuerierCdc)
		case types.QueryGetParams:
			return queryGetParams(ctx, req, keeper, legacyQuerierCdc)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint", types.ModuleName)
		}
	}
}

func queryAssetSupply(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	// Decode request
	var requestParams types.QueryAssetSupply
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	assetSupply, found := keeper.GetAssetSupply(ctx, requestParams.Denom)
	if !found {
		return nil, sdkerrors.Wrap(types.ErrAssetSupplyNotFound, string(requestParams.Denom))
	}

	// Encode results
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, assetSupply)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryAssetSupplies(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) (res []byte, err error) {
	assets := keeper.GetAllAssetSupplies(ctx)
	if assets == nil {
		assets = []types.AssetSupply{}
	}

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, assets)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryAtomicSwap(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	// Decode request
	var requestParams types.QueryAtomicSwapByID
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	// Lookup atomic swap
	atomicSwap, found := keeper.GetAtomicSwap(ctx, requestParams.SwapID)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrAtomicSwapNotFound, "%d", requestParams.SwapID)
	}

	augmentedAtomicSwap := types.NewAugmentedAtomicSwap(atomicSwap)

	// Encode results
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, augmentedAtomicSwap)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryAtomicSwaps(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryAtomicSwaps
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	unfilteredSwaps := keeper.GetAllAtomicSwaps(ctx)
	swaps := filterAtomicSwaps(ctx, unfilteredSwaps, params)
	if swaps == nil {
		swaps = []types.AtomicSwap{}
	}

	augmentedSwaps := []types.LegacyAugmentedAtomicSwap{}

	for _, swap := range swaps {
		augmentedSwaps = append(augmentedSwaps, types.NewLegacyAugmentedAtomicSwap(swap))
	}

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, augmentedSwaps)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

// query params in the bep3 store
func queryGetParams(ctx sdk.Context, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	// Get params
	params := keeper.GetParams(ctx)

	// Encode results
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

// filterAtomicSwaps retrieves atomic swaps filtered by a given set of params.
// If no filters are provided, all atomic swaps will be returned in paginated form.
func filterAtomicSwaps(ctx sdk.Context, swaps []types.AtomicSwap, params types.QueryAtomicSwaps) []types.AtomicSwap {
	filteredSwaps := make([]types.AtomicSwap, 0, len(swaps))

	for _, s := range swaps {
		if legacyAtomicSwapIsMatch(s, params) {
			filteredSwaps = append(filteredSwaps, s)
		}
	}

	start, end := client.Paginate(len(filteredSwaps), params.Page, params.Limit, 100)
	if start < 0 || end < 0 {
		filteredSwaps = []types.AtomicSwap{}
	} else {
		filteredSwaps = filteredSwaps[start:end]
	}

	return filteredSwaps
}

func legacyAtomicSwapIsMatch(swap types.AtomicSwap, params types.QueryAtomicSwaps) bool {
	// match involved address (if supplied)
	if len(params.Involve) > 0 {
		if !swap.Sender.Equals(params.Involve) && !swap.Recipient.Equals(params.Involve) {
			return false
		}
	}

	// match expiration block limit (if supplied)
	if params.Expiration > 0 {
		if swap.ExpireHeight > params.Expiration {
			return false
		}
	}

	// match status (if supplied/valid)
	if params.Status.IsValid() {
		if swap.Status != params.Status {
			return false
		}
	}

	// match direction (if supplied/valid)
	if params.Direction.IsValid() {
		if swap.Direction != params.Direction {
			return false
		}
	}

	return true
}
