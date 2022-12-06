package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/kava-labs/kava/x/kavamint/types"
)

var _ types.QueryServer = Keeper{}

// Params returns params of the mint module.
func (k Keeper) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)

	return &types.QueryParamsResponse{Params: params}, nil
}

// Inflation returns minter.Inflation of the mint module.
func (k Keeper) Inflation(c context.Context, _ *types.QueryInflationRequest) (*types.QueryInflationResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	inflation := k.CumulativeInflation(ctx)

	return &types.QueryInflationResponse{Inflation: inflation}, nil
}

// MintQueryServer implements cosmos sdk's x/mint querier.
// x/mint was removed from kava, but the standard inflation endpoint is still registered
// for easier third party integration and backwards compatibility.
type MintQueryServer struct {
	keeper Keeper
}

// NewMintQueryServer returns a service that implements x/mint's QueryServer
func NewMintQueryServer(kavamintKeeper Keeper) MintQueryServer {
	return MintQueryServer{kavamintKeeper}
}

var _ minttypes.QueryServer = MintQueryServer{}

// Params is not implemented. There is no mint module.
func (MintQueryServer) Params(
	_ context.Context, _ *minttypes.QueryParamsRequest,
) (*minttypes.QueryParamsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "x/mint has been replaced by x/kavamint")
}

// Inflation returns an adjusted inflation rate.
// The `/cosmos/mint/v1beta1/inflation` endpoint is used by third parties to calculate staking APY.
// The usual staking APY calculation takes the inflation and determines the portion of it devoted
// to staking rewards after adjusting for the bonded ratio and x/distribution community_tax.
// staking_apy = (inflation - community_tax) * total_supply / total_bonded
// Staking APY is not set directly via the x/kavamint staking_rewards_apy param.
// This endpoint returns the inflation that makes the above calculation equal to the param:
// inflation = staking_apy * total_bonded / total_supply
// NOTE: assumes x/distribution community_tax = 0
func (mq MintQueryServer) Inflation(
	c context.Context, _ *minttypes.QueryInflationRequest,
) (*minttypes.QueryInflationResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	stakingApy := mq.keeper.GetParams(ctx).StakingRewardsApy
	totalBonded := mq.keeper.TotalBondedTokens(ctx)
	totalSupply := mq.keeper.TotalSupply(ctx)

	// inflation = staking_apy * total_bonded / total_supply
	inflation := stakingApy.MulInt(totalBonded).QuoInt(totalSupply)

	return &minttypes.QueryInflationResponse{
		Inflation: inflation,
	}, nil
}

// AnnualProvisions is not implemented.
func (MintQueryServer) AnnualProvisions(
	_ context.Context, _ *minttypes.QueryAnnualProvisionsRequest,
) (*minttypes.QueryAnnualProvisionsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "x/mint has been replaced by x/kavamint")
}
