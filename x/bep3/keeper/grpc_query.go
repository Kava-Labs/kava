package keeper

import (
	"context"
	"encoding/hex"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/kava-labs/kava/x/bep3/types"
)

type queryServer struct {
	keeper Keeper
}

// NewQueryServerImpl creates a new server for handling gRPC queries.
func NewQueryServerImpl(k Keeper) types.QueryServer {
	return &queryServer{keeper: k}
}

var _ types.QueryServer = queryServer{}

// Params queries module params
func (s queryServer) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := s.keeper.GetParams(sdkCtx)

	return &types.QueryParamsResponse{Params: params}, nil
}

// AssetSupply queries info about an asset's supply
func (s queryServer) AssetSupply(ctx context.Context, req *types.QueryAssetSupplyRequest) (*types.QueryAssetSupplyResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	assetSupply, ok := s.keeper.GetAssetSupply(sdkCtx, req.Denom)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "denom not found")
	}

	return &types.QueryAssetSupplyResponse{AssetSupply: mapAssetSupplyToResponse(assetSupply)}, nil
}

// AssetSupplies queries a list of asset supplies
func (s queryServer) AssetSupplies(ctx context.Context, req *types.QueryAssetSuppliesRequest) (*types.QueryAssetSuppliesResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	var queryResults []types.AssetSupplyResponse
	s.keeper.IterateAssetSupplies(sdkCtx, func(assetSupply types.AssetSupply) bool {
		queryResults = append(queryResults, mapAssetSupplyToResponse(assetSupply))
		return false
	})

	return &types.QueryAssetSuppliesResponse{
		AssetSupplies: queryResults,
	}, nil
}

// AtomicSwap queries info about an atomic swap
func (s queryServer) AtomicSwap(ctx context.Context, req *types.QueryAtomicSwapRequest) (*types.QueryAtomicSwapResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	swapId, err := hex.DecodeString(req.SwapId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "invalid atomic swap id")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	atomicSwap, ok := s.keeper.GetAtomicSwap(sdkCtx, swapId)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "invalid atomic swap")
	}

	return &types.QueryAtomicSwapResponse{
		AtomicSwap: mapAtomicSwapToResponse(atomicSwap),
	}, nil
}

// AtomicSwaps queries a list of atomic swaps
func (s queryServer) AtomicSwaps(ctx context.Context, req *types.QueryAtomicSwapsRequest) (*types.QueryAtomicSwapsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(s.keeper.key), types.AtomicSwapKeyPrefix)

	var queryResults []types.AtomicSwapResponse
	pageRes, err := query.FilteredPaginate(store, req.Pagination, func(_, value []byte, shouldAccumulate bool) (bool, error) {
		var atomicSwap types.AtomicSwap
		err := s.keeper.cdc.Unmarshal(value, &atomicSwap)
		if err != nil {
			return false, err
		}

		if len(req.Involve) > 0 {
			if atomicSwap.Sender.String() != req.Involve && atomicSwap.Recipient.String() != req.Involve {
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
			queryResults = append(queryResults, mapAtomicSwapToResponse(atomicSwap))
		}
		return true, nil
	})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "paginate: %v", err)
	}

	return &types.QueryAtomicSwapsResponse{
		AtomicSwaps: queryResults,
		Pagination:  pageRes,
	}, nil
}

func mapAssetSupplyToResponse(assetSupply types.AssetSupply) types.AssetSupplyResponse {
	return types.AssetSupplyResponse{
		IncomingSupply:           assetSupply.IncomingSupply,
		OutgoingSupply:           assetSupply.OutgoingSupply,
		CurrentSupply:            assetSupply.CurrentSupply,
		TimeLimitedCurrentSupply: assetSupply.TimeLimitedCurrentSupply,
		TimeElapsed:              assetSupply.TimeElapsed,
	}
}

func mapAtomicSwapToResponse(atomicSwap types.AtomicSwap) types.AtomicSwapResponse {
	return types.AtomicSwapResponse{
		Id:                  atomicSwap.GetSwapID().String(),
		Amount:              atomicSwap.Amount,
		RandomNumberHash:    atomicSwap.RandomNumberHash.String(),
		ExpireHeight:        atomicSwap.ExpireHeight,
		Timestamp:           atomicSwap.Timestamp,
		Sender:              atomicSwap.Sender.String(),
		Recipient:           atomicSwap.Recipient.String(),
		SenderOtherChain:    atomicSwap.SenderOtherChain,
		RecipientOtherChain: atomicSwap.RecipientOtherChain,
		ClosedBlock:         atomicSwap.ClosedBlock,
		Status:              atomicSwap.Status,
		CrossChain:          atomicSwap.CrossChain,
		Direction:           atomicSwap.Direction,
	}
}
