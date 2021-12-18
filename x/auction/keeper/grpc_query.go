package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	proto "github.com/gogo/protobuf/proto"

	"github.com/kava-labs/kava/x/auction/types"
)

type queryServer struct {
	keeper Keeper
}

// NewQueryServer creates a new server for handling gRPC queries.
func NewQueryServer(k Keeper) types.QueryServer {
	return &queryServer{keeper: k}
}

// Params implements the gRPC service handler for querying x/auction parameters.
func (q *queryServer) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := q.keeper.GetParams(sdkCtx)

	return &types.QueryParamsResponse{Params: params}, nil
}

// Auction implements the Query/Auction gRPC method
func (q *queryServer) Auction(c context.Context, req *types.QueryAuctionRequest) (*types.QueryAuctionResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	auction, found := q.keeper.GetAuction(ctx, req.AuctionId)
	if !found {
		return &types.QueryAuctionResponse{}, nil
	}

	msg, ok := auction.(proto.Message)
	if !ok {
		return nil, status.Errorf(codes.Internal, "can't protomarshal %T", msg)
	}

	auctionAny, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &types.QueryAuctionResponse{
		Auction: auctionAny,
	}, nil
}

// Auctions implements the Query/Auctions gRPC method
func (q *queryServer) Auctions(c context.Context, req *types.QueryAuctionsRequest) (*types.QueryAuctionsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	var auctions []*codectypes.Any
	auctionStore := prefix.NewStore(ctx.KVStore(q.keeper.storeKey), types.AuctionKeyPrefix)

	pageRes, err := query.FilteredPaginate(auctionStore, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		result, err := q.keeper.UnmarshalAuction(value)
		if err != nil {
			return false, err
		}

		// True if empty owner, otherwise check if auction contains owner
		ownerIsMatch := req.Owner == ""
		if req.Owner != "" {
			if cAuc, ok := result.(*types.CollateralAuction); ok {
				for _, addr := range cAuc.GetLotReturns().Addresses {
					if addr.String() == req.Owner {
						ownerIsMatch = true
						break
					}
				}
			}
		}

		phaseIsMatch := req.Phase == "" || req.Phase == result.GetPhase()
		typeIsMatch := req.Type == "" || req.Type == result.GetType()
		denomIsMatch := req.Denom == "" || req.Denom == result.GetBid().Denom || req.Denom == result.GetLot().Denom

		if ownerIsMatch && phaseIsMatch && typeIsMatch && denomIsMatch {
			if accumulate {
				msg, ok := result.(proto.Message)
				if !ok {
					return false, status.Errorf(codes.Internal, "can't protomarshal %T", msg)
				}

				auctionAny, err := codectypes.NewAnyWithValue(msg)
				if err != nil {
					return false, err
				}
				auctions = append(auctions, auctionAny)
			}

			return true, nil
		}

		return false, nil
	})
	if err != nil {
		return &types.QueryAuctionsResponse{}, err
	}

	return &types.QueryAuctionsResponse{
		Auction:    auctions,
		Pagination: pageRes,
	}, nil
}

// NextAuctionID implements the gRPC service handler for querying x/auction next auction ID.
func (q *queryServer) NextAuctionID(ctx context.Context, req *types.QueryNextAuctionIDRequest) (*types.QueryNextAuctionIDResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	nextAuctionID, err := q.keeper.GetNextAuctionID(sdkCtx)
	if err != nil {
		return &types.QueryNextAuctionIDResponse{}, err
	}

	return &types.QueryNextAuctionIDResponse{Id: nextAuctionID}, nil
}
