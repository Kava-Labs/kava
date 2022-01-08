package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/kavadist/types"
)

type queryServer struct {
	keeper Keeper
}

// NewQueryServerImpl creates a new server for handling gRPC queries.
func NewQueryServerImpl(k Keeper) types.QueryServer {
	return &queryServer{keeper: k}
}

var _ types.QueryServer = queryServer{}

func (s queryServer) Balance(ctx context.Context, req *types.QueryBalanceRequest) (*types.QueryBalanceResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	acc := s.keeper.accountKeeper.GetModuleAccount(sdkCtx, types.KavaDistMacc)
	balance := s.keeper.bankKeeper.GetAllBalances(sdkCtx, acc.GetAddress())
	return &types.QueryBalanceResponse{Coins: balance}, nil
}

func (s queryServer) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := s.keeper.GetParams(sdkCtx)

	return &types.QueryParamsResponse{Params: params}, nil
}
