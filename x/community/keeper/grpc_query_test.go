package keeper_test

import (
	"context"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/community/keeper"
	"github.com/kava-labs/kava/x/community/types"
)

type grpcQueryTestSuite struct {
	KeeperTestSuite

	queryClient types.QueryClient
}

func (suite *grpcQueryTestSuite) SetupTest() {
	suite.KeeperTestSuite.SetupTest()

	queryHelper := baseapp.NewQueryServerTestHelper(suite.Ctx, suite.App.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServerImpl(suite.Keeper))

	suite.queryClient = types.NewQueryClient(queryHelper)
}

func TestGrpcQueryTestSuite(t *testing.T) {
	suite.Run(t, new(grpcQueryTestSuite))
}

func (suite *grpcQueryTestSuite) TestGrpcQueryParams() {
	p := types.NewParams(
		time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
		sdkmath.LegacyNewDec(1000),
		sdkmath.LegacyNewDec(1000),
	)
	suite.Keeper.SetParams(suite.Ctx, p)

	res, err := suite.queryClient.Params(context.Background(), &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Equal(
		types.QueryParamsResponse{
			Params: p,
		},
		*res,
	)
}

func (suite *grpcQueryTestSuite) TestGrpcQueryBalance() {
	var expCoins sdk.Coins

	testCases := []struct {
		name  string
		setup func()
	}{
		{
			name:  "handles response with no balance",
			setup: func() { expCoins = sdk.Coins{} },
		},
		{
			name: "handles response with balance",
			setup: func() {
				expCoins = sdk.NewCoins(
					sdk.NewCoin("ukava", sdkmath.NewInt(100)),
					sdk.NewCoin("usdx", sdkmath.NewInt(1000)),
				)
				suite.App.FundModuleAccount(suite.Ctx, types.ModuleName, expCoins)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.setup()
			res, err := suite.queryClient.Balance(context.Background(), &types.QueryBalanceRequest{})
			suite.Require().NoError(err)
			suite.Require().True(expCoins.IsEqual(res.Coins))
		})
	}
}

func (suite *grpcQueryTestSuite) TestGrpcQueryTotalBalance() {
	var expCoins sdk.DecCoins

	testCases := []struct {
		name  string
		setup func()
	}{
		{
			name:  "handles response with no balance",
			setup: func() { expCoins = sdk.DecCoins{} },
		},
		{
			name: "handles response with balance",
			setup: func() {
				expCoins = sdk.NewDecCoins(
					sdk.NewDecCoin("ukava", sdkmath.NewInt(100)),
					sdk.NewDecCoin("usdx", sdkmath.NewInt(1000)),
				)

				coins, _ := expCoins.TruncateDecimal()

				suite.App.FundModuleAccount(suite.Ctx, types.ModuleName, coins)
			},
		},
		{
			name: "handles response with both x/community + x/distribution balance",
			setup: func() {
				decCoins1 := sdk.NewDecCoins(
					sdk.NewDecCoin("ukava", sdkmath.NewInt(100)),
					sdk.NewDecCoin("usdx", sdkmath.NewInt(1000)),
				)

				coins, _ := decCoins1.TruncateDecimal()

				err := suite.App.FundModuleAccount(suite.Ctx, types.ModuleName, coins)
				suite.Require().NoError(err)

				decCoins2 := sdk.NewDecCoins(
					sdk.NewDecCoin("ukava", sdkmath.NewInt(100)),
					sdk.NewDecCoin("usdc", sdkmath.NewInt(1000)),
				)

				// Add to x/distribution community pool (just state, not actual coins)
				dk := suite.App.GetDistrKeeper()
				feePool := dk.GetFeePool(suite.Ctx)
				feePool.CommunityPool = feePool.CommunityPool.Add(decCoins2...)
				dk.SetFeePool(suite.Ctx, feePool)

				expCoins = decCoins1.Add(decCoins2...)
			},
		},
		{
			name: "handles response with only x/distribution balance",
			setup: func() {
				expCoins = sdk.NewDecCoins(
					sdk.NewDecCoin("ukava", sdkmath.NewInt(100)),
					sdk.NewDecCoin("usdc", sdkmath.NewInt(1000)),
				)

				// Add to x/distribution community pool (just state, not actual coins)
				dk := suite.App.GetDistrKeeper()
				feePool := dk.GetFeePool(suite.Ctx)
				feePool.CommunityPool = feePool.CommunityPool.Add(expCoins...)
				dk.SetFeePool(suite.Ctx, feePool)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.setup()
			res, err := suite.queryClient.TotalBalance(context.Background(), &types.QueryTotalBalanceRequest{})
			suite.Require().NoError(err)
			suite.Require().True(expCoins.IsEqual(res.Pool))
		})
	}
}
