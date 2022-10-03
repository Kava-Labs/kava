package keeper

import (
	"context"

	"github.com/kava-labs/kava/x/liquid/types"
)

type queryServer struct {
	keeper Keeper
}

// NewQueryServerImpl creates a new server for handling gRPC queries.
func NewQueryServerImpl(k Keeper) types.QueryServer {
	return &queryServer{keeper: k}
}

var _ types.QueryServer = queryServer{}

func (s queryServer) DelegatedBalance(
	ctx context.Context,
	req *types.QueryDelegatedBalanceRequest,
) (*types.QueryDelegatedBalanceResponse, error) {
	return &types.QueryDelegatedBalanceResponse{}, nil
}
