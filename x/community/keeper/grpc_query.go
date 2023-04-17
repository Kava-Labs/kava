package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
