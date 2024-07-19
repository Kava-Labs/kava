package keeper_test

import (
	"context"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/precisebank/keeper"
	"github.com/kava-labs/kava/x/precisebank/testutil"
	"github.com/kava-labs/kava/x/precisebank/types"
)

type grpcQueryTestSuite struct {
	testutil.Suite

	queryClient types.QueryClient
}

func (suite *grpcQueryTestSuite) SetupTest() {
	suite.Suite.SetupTest()

	queryHelper := baseapp.NewQueryServerTestHelper(suite.Ctx, suite.App.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServerImpl(suite.Keeper))

	suite.queryClient = types.NewQueryClient(queryHelper)
}

func TestGrpcQueryTestSuite(t *testing.T) {
	suite.Run(t, new(grpcQueryTestSuite))
}

func (suite *grpcQueryTestSuite) TestQueryTotalFractionalBalance() {
	testCases := []struct {
		name         string
		giveBalances []sdkmath.Int
	}{
		{
			"empty",
			[]sdkmath.Int{},
		},
		{
			"min amount",
			[]sdkmath.Int{
				types.ConversionFactor().QuoRaw(2),
				types.ConversionFactor().QuoRaw(2),
			},
		},
		{
			"exceeds conversion factor",
			[]sdkmath.Int{
				// 4 accounts * 0.5 == 2
				types.ConversionFactor().QuoRaw(2),
				types.ConversionFactor().QuoRaw(2),
				types.ConversionFactor().QuoRaw(2),
				types.ConversionFactor().QuoRaw(2),
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			total := sdkmath.ZeroInt()
			for i, balance := range tc.giveBalances {
				addr := sdk.AccAddress([]byte{byte(i)})
				suite.Keeper.SetFractionalBalance(suite.Ctx, addr, balance)

				total = total.Add(balance)
			}

			res, err := suite.queryClient.TotalFractionalBalances(
				context.Background(),
				&types.QueryTotalFractionalBalancesRequest{},
			)
			suite.Require().NoError(err)

			suite.Require().Equal(total, res.Total)
		})
	}
}
