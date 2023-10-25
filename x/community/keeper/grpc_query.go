package keeper

import (
	"context"

	sdkmath "cosmossdk.io/math"
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

// AnnualizedRewards calculates the annualized rewards for the chain.
// This method is backported from v0.25.x to allow for early migration.
func (s queryServer) AnnualizedRewards(
	c context.Context,
	req *types.QueryAnnualizedRewardsRequest,
) (*types.QueryAnnualizedRewardsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	bondDenom := s.keeper.stakingKeeper.BondDenom(ctx)

	totalSupply := s.keeper.bankKeeper.GetSupply(ctx, bondDenom).Amount
	totalBonded := s.keeper.stakingKeeper.TotalBondedTokens(ctx)
	rewardsPerSecond := sdkmath.LegacyZeroDec() // always zero. this method is backported from v0.25.x to allow for early migration.
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
