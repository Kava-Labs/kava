package keeper_test

import (
	"github.com/kava-labs/kava/x/swap/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite keeperTestSuite) TestParams() {
	keeper := suite.Keeper

	params := types.Params{
		AllowedPools: types.AllowedPools{
			types.NewAllowedPool("ukava", "usdx"),
		},
		SwapFee: sdk.MustNewDecFromStr("0.03"),
	}
	keeper.SetParams(suite.Ctx, params)
	suite.Equal(keeper.GetParams(suite.Ctx), params)

	oldParams := params
	params = types.Params{
		AllowedPools: types.AllowedPools{
			types.NewAllowedPool("hard", "ukava"),
		},
		SwapFee: sdk.MustNewDecFromStr("0.01"),
	}
	keeper.SetParams(suite.Ctx, params)
	suite.NotEqual(keeper.GetParams(suite.Ctx), oldParams)
	suite.Equal(keeper.GetParams(suite.Ctx), params)
}
