package keeper_test

import (
	"context"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/kavadist/types"
)

func (suite *keeperTestSuite) TestGRPCParams() {
	ctx, keeper, queryClient := suite.Ctx, suite.Keeper, suite.QueryClient

	var (
		params    types.Params
		req       *types.QueryParamsRequest
		expParams types.Params
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"response with default params",
			func() {
				expParams = types.DefaultParams()
				keeper.SetParams(ctx, expParams)
				req = &types.QueryParamsRequest{}
			},
			true,
		},
		{
			"response with params",
			func() {
				params = types.Params{
					Active:  true,
					Periods: suite.TestPeriods,
				}
				keeper.SetParams(ctx, params)
				req = &types.QueryParamsRequest{}
				expParams = params
			},
			true,
		},
	}

	for _, testCase := range testCases {
		suite.Run(fmt.Sprintf("Case %s", testCase.msg), func() {
			testCase.malleate()

			paramsRes, err := queryClient.Params(context.Background(), req)

			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(paramsRes)
				suite.Require().True(expParams.Equal(paramsRes.Params))
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *keeperTestSuite) TestGRPCBalance() {
	ctx, queryClient := suite.Ctx, suite.QueryClient

	var (
		req      *types.QueryBalanceRequest
		expCoins sdk.Coins
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"response with no balance",
			func() {
				expCoins = sdk.Coins{}
				req = &types.QueryBalanceRequest{}
			},
			true,
		},
		{
			"response with balance",
			func() {
				expCoins = sdk.Coins{
					sdk.NewCoin("ukava", sdkmath.NewInt(100)),
				}
				suite.App.FundModuleAccount(ctx, types.ModuleName, expCoins)
				req = &types.QueryBalanceRequest{}
			},
			true,
		},
	}

	for _, testCase := range testCases {
		suite.Run(fmt.Sprintf("Case %s", testCase.msg), func() {
			testCase.malleate()

			res, err := queryClient.Balance(context.Background(), req)

			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().True(expCoins.IsEqual(res.Coins))
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
