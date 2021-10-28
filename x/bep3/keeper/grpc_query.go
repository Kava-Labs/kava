package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/kava-labs/kava/x/bep3/types"
)

var _ types.QueryServer = Keeper{}

// Params queries module params
func (k Keeper) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := k.GetParams(sdkCtx)

	return &types.QueryParamsResponse{Params: params}, nil
}

// AssetSupply queries info about an asset's supply
func (k Keeper) AssetSupply(ctx context.Context, req *types.QueryAssetSupplyRequest) (*types.QueryAssetSupplyResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	assetSupply, ok := k.GetAssetSupply(sdkCtx, req.Denom)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "denom not found")
	}

	return &types.QueryAssetSupplyResponse{AssetSupply: assetSupply}, nil
}

// AssetSupplies queries a list of asset supplies
func (k Keeper) AssetSupplies(ctx context.Context, req *types.QueryAssetSuppliesRequest) (*types.QueryAssetSuppliesResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(k.key), types.AssetSupplyPrefix)

	var queryResults []types.AssetSupply
	pageRes, err := query.FilteredPaginate(store, req.Pagination, func(_, value []byte, shouldAccumulate bool) (bool, error) {
		var assetSupply types.AssetSupply
		err := k.cdc.UnmarshalLengthPrefixed(value, &assetSupply)
		if err != nil {
			return false, err
		}

		if shouldAccumulate {
			queryResults = append(queryResults, assetSupply)
		}
		return true, nil
	})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "paginate: %v", err)
	}

	return &types.QueryAssetSuppliesResponse{
		AssetSupplies: queryResults,
		Pagination:    pageRes,
	}, nil
}

// AtomicSwap queries info about an atomic swap
func (k Keeper) AtomicSwap(ctx context.Context, req *types.QueryAtomicSwapRequest) (*types.QueryAtomicSwapResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	atomicSwap, ok := k.GetAtomicSwap(sdkCtx, req.SwapId)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "invalid atomic swap")
	}

	return &types.QueryAtomicSwapResponse{
		ID:         atomicSwap.GetSwapID().String(),
		AtomicSwap: atomicSwap,
	}, nil
}

// AtomicSwaps queries a list of atomic swaps
func (k Keeper) AtomicSwaps(ctx context.Context, req *types.QueryAtomicSwapsRequest) (*types.QueryAtomicSwapsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(k.key), types.AtomicSwapKeyPrefix)

	var queryResults []types.AugmentedAtomicSwap
	pageRes, err := query.FilteredPaginate(store, req.Pagination, func(_, value []byte, shouldAccumulate bool) (bool, error) {
		var atomicSwap types.AtomicSwap
		err := k.cdc.UnmarshalLengthPrefixed(value, &atomicSwap)
		if err != nil {
			return false, err
		}

		if len(req.Involve) > 0 {
			if atomicSwap.Sender != req.Involve && atomicSwap.Recipient != req.Involve {
				return false, nil
			}
		}

		// match expiration block limit (if supplied)
		if req.Expiration > 0 {
			if atomicSwap.ExpireHeight > req.Expiration {
				return false, nil
			}
		}

		// match status (if supplied/valid)
		if req.Status.IsValid() {
			if atomicSwap.Status != req.Status {
				return false, nil
			}
		}

		// match direction (if supplied/valid)
		if req.Direction.IsValid() {
			if atomicSwap.Direction != req.Direction {
				return false, nil
			}
		}

		if shouldAccumulate {
			augmented := types.NewAugmentedAtomicSwap(atomicSwap)
			queryResults = append(queryResults, augmented)
		}
		return true, nil
	})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "paginate: %v", err)
	}

	return &types.QueryAtomicSwapsResponse{
		AtomicSwap: queryResults,
		Pagination: pageRes,
	}, nil
}
