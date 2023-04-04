package keeper_test

import (
	"context"
	"testing"

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
