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

// LegacyCommunityPool implements the gRPC service handler for querying the legacy community pool balance.
func (s queryServer) LegacyCommunityPool(c context.Context, _ *types.QueryLegacyCommunityPoolRequest) (*types.QueryLegacyCommunityPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	balance := s.keeper.distrKeeper.GetFeePoolCommunityCoins(ctx)
	return &types.QueryLegacyCommunityPoolResponse{
		Address: s.keeper.legacyCommunityPoolAddress.String(),
		Balance: balance,
	}, nil
}
