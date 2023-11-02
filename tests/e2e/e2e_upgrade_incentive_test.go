package e2e_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/tests/util"
	incentivetypes "github.com/kava-labs/kava/x/incentive/types"
)

func (suite *IntegrationTestSuite) TestUpgradeIncentiveParams() {
	suite.SkipIfUpgradeDisabled()

	beforeUpgradeCtx := util.CtxAtHeight(suite.UpgradeHeight - 1)
	afterUpgradeCtx := util.CtxAtHeight(suite.UpgradeHeight)

	// Before params
	incentiveParamsBefore, err := suite.Kava.Incentive.Params(
		beforeUpgradeCtx,
		&incentivetypes.QueryParamsRequest{},
	)
	suite.NoError(err)

	incentiveParamsAfter, err := suite.Kava.Incentive.Params(
		afterUpgradeCtx,
		&incentivetypes.QueryParamsRequest{},
	)
	suite.NoError(err)

	suite.Run("x/incentive parameters before upgrade", func() {
		suite.Require().Len(
			incentiveParamsBefore.Params.EarnRewardPeriods,
			1,
			"x/incentive should have 1 earn reward period before upgrade",
		)

		suite.Require().Equal(
			sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(159_459))),
			incentiveParamsBefore.Params.EarnRewardPeriods[0].RewardsPerSecond,
		)
	})

	suite.Run("x/incentive parameters after upgrade", func() {
		suite.Require().Len(
			incentiveParamsAfter.Params.EarnRewardPeriods,
			1,
			"x/incentive should have 1 earn reward period before upgrade",
		)

		suite.Require().Equal(
			// Manual calculation of
			// 600,000 * 1000,000 / (365 * 24 * 60 * 60)
			sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(19025))),
			incentiveParamsAfter.Params.EarnRewardPeriods[0].RewardsPerSecond,
		)

		// No other changes
		incentiveParamsAfter.Params.EarnRewardPeriods[0].RewardsPerSecond = incentiveParamsBefore.Params.EarnRewardPeriods[0].RewardsPerSecond
		suite.Require().Equal(
			incentiveParamsBefore,
			incentiveParamsAfter,
			"other param values should not be changed",
		)
	})
}
