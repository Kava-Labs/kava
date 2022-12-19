package keeper_test

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/community/keeper"
	"github.com/kava-labs/kava/x/community/types"
)

const legacyCommunityPoolAddr = "kava1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8m2splc"

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
					sdk.NewCoin("ukava", sdk.NewInt(100)),
					sdk.NewCoin("usdx", sdk.NewInt(1000)),
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

func (suite *grpcQueryTestSuite) TestLegacyCommunityPoolBalance() {
	// watch for regressions in name of account holding community pool funds
	suite.Equal(types.LegacyCommunityPoolModuleName, distrtypes.ModuleName)

	testCases := []struct {
		name    string
		balance sdk.DecCoins
	}{
		{
			name: "success - nonzero balance, single denom",
			balance: sdk.NewDecCoins(
				sdk.NewDecCoinFromDec("ukava", sdk.MustNewDecFromStr("1234567.89")),
				sdk.NewDecCoinFromDec("usdx", sdk.NewDec(1e5)),
				sdk.NewDecCoinFromDec("other-denom", sdk.MustNewDecFromStr("0.00000000003")),
			),
		},
		{
			name: "success - nonzero balance, multiple denoms",
			balance: sdk.NewDecCoins(
				sdk.NewDecCoinFromDec("ukava", sdk.MustNewDecFromStr("1234567.89")),
			),
		},
		{
			name:    "success - zero balance",
			balance: sdk.NewDecCoins(),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			distrKeeper := suite.App.GetDistrKeeper()

			// fund the fee pool
			if !tc.balance.IsZero() {
				feePool := distrKeeper.GetFeePool(suite.Ctx)
				feePool.CommunityPool = feePool.CommunityPool.Add(tc.balance...)
				distrKeeper.SetFeePool(suite.Ctx, feePool)
			}

			// query legacy community pool
			res, err := suite.queryClient.LegacyCommunityPool(
				context.Background(), &types.QueryLegacyCommunityPoolRequest{},
			)
			suite.NoError(err)
			suite.True(tc.balance.IsEqual(res.Balance))
			suite.Equal(legacyCommunityPoolAddr, res.Address)
		})
	}
}
