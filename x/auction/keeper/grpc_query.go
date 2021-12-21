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

// NewQueryServerImpl creates a new server for handling gRPC queries.
func NewQueryServerImpl(k Keeper) types.QueryServer {
	return &queryServer{keeper: k}
}

var _ types.QueryServer = queryServer{}

// Params implements the gRPC service handler for querying x/auction parameters.
func (s queryServer) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := s.keeper.GetParams(sdkCtx)

	return &types.QueryParamsResponse{Params: params}, nil
}

// Auction implements the Query/Auction gRPC method
func (s queryServer) Auction(c context.Context, req *types.QueryAuctionRequest) (*types.QueryAuctionResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	auction, found := s.keeper.GetAuction(ctx, req.AuctionId)
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
func (s queryServer) Auctions(c context.Context, req *types.QueryAuctionsRequest) (*types.QueryAuctionsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	var auctions []*codectypes.Any
	auctionStore := prefix.NewStore(ctx.KVStore(s.keeper.storeKey), types.AuctionKeyPrefix)

	pageRes, err := query.Paginate(auctionStore, req.Pagination, func(key []byte, value []byte) error {
		result, err := s.keeper.UnmarshalAuction(value)
		if err != nil {
			return err
		}

		msg, ok := result.(proto.Message)
		if !ok {
			return status.Errorf(codes.Internal, "can't protomarshal %T", msg)
		}

		auctionAny, err := codectypes.NewAnyWithValue(msg)
		if err != nil {
			return err
		}
		auctions = append(auctions, auctionAny)
		return nil
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
func (s queryServer) NextAuctionID(ctx context.Context, req *types.QueryNextAuctionIDRequest) (*types.QueryNextAuctionIDResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	nextAuctionID, err := s.keeper.GetNextAuctionID(sdkCtx)
	if err != nil {
		return &types.QueryNextAuctionIDResponse{}, err
	}

	return &types.QueryNextAuctionIDResponse{Id: nextAuctionID}, nil
}
