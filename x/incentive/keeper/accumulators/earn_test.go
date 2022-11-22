package accumulators_test

import (
	"fmt"
	"time"

	"github.com/kava-labs/kava/x/incentive/keeper/accumulators"
	"github.com/kava-labs/kava/x/incentive/types"
)

func (suite *AccumulateEarnRewardsIntegrationTests) TestEarnAccumulator_OnlyEarnClaimType() {
	period := types.NewMultiRewardPeriod(
		true,
		"bkava",
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(c("earn", 2000), c("ukava", 1000)), // same denoms as in global indexes
	)

	earnKeeper := suite.App.GetEarnKeeper()

	for _, claimTypeValue := range types.ClaimType_value {
		claimType := types.ClaimType(claimTypeValue)

		if claimType == types.CLAIM_TYPE_EARN {
			suite.NotPanics(func() {
				err := accumulators.
					NewEarnAccumulator(suite.keeper.Store, suite.App.GetLiquidKeeper(), &earnKeeper, suite.keeper.Adapters).
					AccumulateRewards(suite.Ctx, claimType, period)
				suite.NoError(err)
			})

			continue
		}

		suite.PanicsWithValue(
			fmt.Sprintf(
				"invalid claim type for earn accumulator, expected %s but got %s",
				types.CLAIM_TYPE_EARN,
				claimType,
			),
			func() {
				err := accumulators.
					NewEarnAccumulator(suite.keeper.Store, suite.App.GetLiquidKeeper(), &earnKeeper, suite.keeper.Adapters).
					AccumulateRewards(suite.Ctx, claimType, period)
				suite.NoError(err)
			},
		)
	}
}
