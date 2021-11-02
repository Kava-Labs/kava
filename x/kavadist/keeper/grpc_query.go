package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/kavadist/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Balance(ctx context.Context, req *types.QueryBalanceRequest) (*types.QueryBalanceResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	acc := k.accountKeeper.GetModuleAccount(sdkCtx, types.KavaDistMacc)
	balance := k.bankKeeper.GetAllBalances(sdkCtx, acc.GetAddress())
	return &types.QueryBalanceResponse{Coins: balance}, nil
}

func (k Keeper) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := k.GetParams(sdkCtx)

	return &types.QueryParamsResponse{Params: params}, nil
}
