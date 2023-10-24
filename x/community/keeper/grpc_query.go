package keeper

import (
	"context"

	sdkmath "cosmossdk.io/math"
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

// AnnualizedRewards calculates the annualized rewards for the chain.
func (s queryServer) AnnualizedRewards(
	c context.Context,
	req *types.QueryAnnualizedRewardsRequest,
) (*types.QueryAnnualizedRewardsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	// staking rewards come from one of two sources depending on if inflation is enabled or not.
	// at any given time, only one source will contribute to the staking rewards. the other will be zero.
	// this method adds both sources together so it is accurate in both cases.

	params := s.keeper.mustGetParams(ctx)
	bondDenom := s.keeper.stakingKeeper.BondDenom(ctx)

	totalSupply := s.keeper.bankKeeper.GetSupply(ctx, bondDenom).Amount
	totalBonded := s.keeper.stakingKeeper.TotalBondedTokens(ctx)
	rewardsPerSecond := params.StakingRewardsPerSecond
	// need to convert these from sdk.Dec to sdkmath.LegacyDec
	inflationRate := convertDecToLegacyDec(s.keeper.mintKeeper.GetMinter(ctx).Inflation)
	communityTax := convertDecToLegacyDec(s.keeper.distrKeeper.GetCommunityTax(ctx))

	return &types.QueryAnnualizedRewardsResponse{
		StakingRewards: CalculateStakingAnnualPercentage(totalSupply, totalBonded, inflationRate, communityTax, rewardsPerSecond),
	}, nil
}

// convertDecToLegacyDec is a helper method for converting between new and old Dec types
// current version of cosmos-sdk in this repo uses sdk.Dec
// this module uses sdkmath.LegacyDec in its parameters
// TODO: remove me after upgrade to cosmos-sdk v50 (LegacyDec is everywhere)
func convertDecToLegacyDec(in sdk.Dec) sdkmath.LegacyDec {
	return sdkmath.LegacyNewDecFromBigIntWithPrec(in.BigInt(), sdk.Precision)
}
