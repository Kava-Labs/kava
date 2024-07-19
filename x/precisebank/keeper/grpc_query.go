package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/precisebank/types"
)

type queryServer struct {
	keeper Keeper
}

// NewQueryServerImpl creates a new server for handling gRPC queries.
func NewQueryServerImpl(k Keeper) types.QueryServer {
	return &queryServer{keeper: k}
}

var _ types.QueryServer = queryServer{}

// TotalFractionalBalances returns the total sum of all fractional balances.
// This is mostly for external verification of the total fractional balances,
// being a multiple of the conversion factor and backed by the reserve.
func (s queryServer) TotalFractionalBalances(
	goCtx context.Context,
	req *types.QueryTotalFractionalBalancesRequest,
) (*types.QueryTotalFractionalBalancesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	totalAmount := s.keeper.GetTotalSumFractionalBalances(ctx)

	totalCoin := sdk.NewCoin(types.ExtendedCoinDenom, totalAmount)

	return &types.QueryTotalFractionalBalancesResponse{
		Total: totalCoin,
	}, nil
}
