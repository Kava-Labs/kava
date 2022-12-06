package keeper_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/staking"

	"github.com/kava-labs/kava/x/kavamint/keeper"
	"github.com/kava-labs/kava/x/kavamint/testutil"
	"github.com/kava-labs/kava/x/kavamint/types"
)

type grpcQueryTestSuite struct {
	testutil.KavamintTestSuite

	queryClient     types.QueryClient
	mintQueryClient minttypes.QueryClient
}

func (suite *grpcQueryTestSuite) SetupTest() {
	suite.KavamintTestSuite.SetupTest()

	queryHelper := baseapp.NewQueryServerTestHelper(suite.Ctx, suite.App.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.Keeper)
	suite.queryClient = types.NewQueryClient(queryHelper)

	mintQueryServer := keeper.NewMintQueryServer(suite.Keeper)
	minttypes.RegisterQueryServer(queryHelper, mintQueryServer)
	suite.mintQueryClient = minttypes.NewQueryClient(queryHelper)
}

func TestGRPCQueryTestSuite(t *testing.T) {
	suite.Run(t, new(grpcQueryTestSuite))
}

func (suite *grpcQueryTestSuite) TestGRPCQueryParams() {
	app, ctx, queryClient := suite.App, suite.Ctx, suite.queryClient

	kavamintKeeper := app.GetKavamintKeeper()

	params, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(params.Params, kavamintKeeper.GetParams(ctx))
}

func (suite *grpcQueryTestSuite) TestGRPCInflationQuery() {
	testCases := []struct {
		name               string
		communityInflation sdk.Dec
		stakingApy         sdk.Dec
		bondedRatio        sdk.Dec
		expectedInflation  sdk.Dec
	}{
		{
			name:               "no community inflation, no staking apy = no inflation",
			communityInflation: sdk.NewDec(0),
			stakingApy:         sdk.NewDec(0),
			bondedRatio:        sdk.NewDecWithPrec(40, 2),
			expectedInflation:  sdk.NewDec(0),
		},
		{
			name:               "no community inflation means only staking contributes",
			communityInflation: sdk.NewDec(0),
			stakingApy:         sdk.NewDec(1),
			bondedRatio:        sdk.NewDecWithPrec(34, 2),
			expectedInflation:  sdk.NewDecWithPrec(34, 2),
		},
		{
			name:               "no staking apy means only inflation contributes",
			communityInflation: sdk.NewDecWithPrec(75, 2),
			stakingApy:         sdk.NewDec(0),
			bondedRatio:        sdk.NewDecWithPrec(40, 2),
			expectedInflation:  sdk.NewDecWithPrec(75, 2),
		},
		{
			name:               "staking and community inflation combines (100 percent bonded)",
			communityInflation: sdk.NewDec(1),
			stakingApy:         sdk.NewDecWithPrec(50, 2),
			bondedRatio:        sdk.NewDec(1),
			expectedInflation:  sdk.NewDecWithPrec(150, 2),
		},
		{
			name:               "staking and community inflation combines (40 percent bonded)",
			communityInflation: sdk.NewDecWithPrec(90, 2),
			stakingApy:         sdk.NewDecWithPrec(25, 2),
			bondedRatio:        sdk.NewDecWithPrec(40, 2),
			// 90 + .4*25 = 100
			expectedInflation: sdk.NewDec(1),
		},
		{
			name:               "staking and community inflation combines (25 percent bonded)",
			communityInflation: sdk.NewDecWithPrec(90, 2),
			stakingApy:         sdk.NewDecWithPrec(20, 2),
			bondedRatio:        sdk.NewDecWithPrec(25, 2),
			// 90 + .25*20 = 95
			expectedInflation: sdk.NewDecWithPrec(95, 2),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			app, ctx, queryClient := suite.App, suite.Ctx, suite.queryClient

			kavamintKeeper := app.GetKavamintKeeper()

			// set desired params
			kavamintKeeper.SetParams(ctx, types.NewParams(tc.communityInflation, tc.stakingApy))

			// set bonded token ratio
			suite.SetBondedTokenRatio(tc.bondedRatio)
			staking.EndBlocker(ctx, suite.StakingKeeper)

			// query inflation & check for expected results
			inflation, err := queryClient.Inflation(context.Background(), &types.QueryInflationRequest{})
			suite.Require().NoError(err)
			suite.Require().Equal(inflation.Inflation, kavamintKeeper.CumulativeInflation(ctx))
			suite.Require().Equal(inflation.Inflation, tc.expectedInflation)

			// ensure overridden x/mint query for inflation returns the adjusted inflation
			// the adjusted inflation is the inflation that returns the correct staking apy for the
			// standard staking apy calculation used by third parties:
			// staking_apy = (inflation - community_tax) * total_supply / total_bonded
			// => inflation = staking_apy * total_bonded / total_supply
			expectedAdjustedInflation := tc.stakingApy.Mul(tc.bondedRatio)
			mintQ, err := suite.mintQueryClient.Inflation(
				context.Background(),
				&minttypes.QueryInflationRequest{},
			)
			suite.NoError(err)
			suite.Equal(expectedAdjustedInflation, mintQ.Inflation)
		})
	}
}

func (suite *grpcQueryTestSuite) TestUnimplementedMintQueries() {
	suite.SetupTest()

	suite.Run("Params is unimplemented", func() {
		_, err := suite.mintQueryClient.Params(context.Background(), nil)
		suite.Error(err)
		suite.Equal(codes.Unimplemented, status.Code(err))
	})

	suite.Run("AnnualProvisions is unimplemented", func() {
		_, err := suite.mintQueryClient.AnnualProvisions(context.Background(), nil)
		suite.Error(err)
		suite.Equal(codes.Unimplemented, status.Code(err))
	})
}
