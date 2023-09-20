package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/kava-labs/kava/x/community/types"
)

type queryServer struct {
	keeper Keeper
}

var _ types.QueryServer = queryServer{}

// NewQueryServerImpl creates a new server for handling gRPC queries.
func NewQueryServerImpl(k Keeper) types.QueryServer {
	return &queryServer{keeper: k}
}

// Params implements the gRPC service handler for querying x/community params.
func (s queryServer) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	params, found := s.keeper.GetParams(ctx)
	if !found {
		return nil, status.Error(codes.NotFound, "params not found")
	}

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}

// Balance implements the gRPC service handler for querying x/community balance.
func (s queryServer) Balance(c context.Context, _ *types.QueryBalanceRequest) (*types.QueryBalanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	return &types.QueryBalanceResponse{
		Coins: s.keeper.GetModuleAccountBalance(ctx),
	}, nil
}

// CommunityPool queries the community pool coins
func (s queryServer) TotalBalance(
	c context.Context,
	req *types.QueryTotalBalanceRequest,
) (*types.QueryTotalBalanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	// x/distribution community pool balance
	nativePoolBalance := s.keeper.distrKeeper.GetFeePoolCommunityCoins(ctx)

	// x/community pool balance
	communityPoolBalance := s.keeper.GetModuleAccountBalance(ctx)

	totalBalance := nativePoolBalance.Add(sdk.NewDecCoinsFromCoins(communityPoolBalance...)...)

	return &types.QueryTotalBalanceResponse{
		Pool: totalBalance,
	}, nil
}
