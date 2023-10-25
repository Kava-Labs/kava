package e2e_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/kava-labs/kava/tests/util"
)

func (suite *IntegrationTestSuite) TestValMinCommission() {
	suite.SkipIfUpgradeDisabled()

	beforeUpgradeCtx := util.CtxAtHeight(suite.UpgradeHeight - 1)
	afterUpgradeCtx := util.CtxAtHeight(suite.UpgradeHeight)

	suite.Run("before upgrade", func() {
		// Before params
		beforeParams, err := suite.Kava.Staking.Params(beforeUpgradeCtx, &types.QueryParamsRequest{})
		suite.Require().NoError(err)

		suite.Require().Equal(
			sdkmath.LegacyZeroDec().String(),
			beforeParams.Params.MinCommissionRate.String(),
			"min commission rate should be 0%% before upgrade",
		)

		// Before validators
		beforeValidators, err := suite.Kava.Staking.Validators(beforeUpgradeCtx, &types.QueryValidatorsRequest{})
		suite.Require().NoError(err)

		for _, val := range beforeValidators.Validators {
			expectedRate := sdkmath.LegacyMustNewDecFromStr("0.1")

			suite.Require().Equalf(
				expectedRate.String(),
				val.Commission.CommissionRates.Rate.String(),
				"validator %s should have commission rate of %s before upgrade",
				val.OperatorAddress,
				expectedRate,
			)
		}
	})

	suite.Run("after upgrade", func() {
		// After params
		afterParams, err := suite.Kava.Staking.Params(afterUpgradeCtx, &types.QueryParamsRequest{})
		suite.Require().NoError(err)

		suite.Require().Equal(
			sdkmath.LegacyMustNewDecFromStr("0.05").String(),
			afterParams.Params.MinCommissionRate.String(),
			"min commission rate should be 5%% after upgrade",
		)

		// After validators
		afterValidators, err := suite.Kava.Staking.Validators(afterUpgradeCtx, &types.QueryValidatorsRequest{})
		suite.Require().NoError(err)

		for _, val := range afterValidators.Validators {
			suite.Require().Truef(
				val.Commission.CommissionRates.Rate.GTE(sdk.MustNewDecFromStr("0.05")),
				"validator %s should have commission rate of at least 5%%",
				val.OperatorAddress,
			)
		}
	})
}
