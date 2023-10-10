package e2e_test

import (
	"context"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/kava-labs/kava/tests/util"
	communitytypes "github.com/kava-labs/kava/x/community/types"
	kavadisttypes "github.com/kava-labs/kava/x/kavadist/types"
)

func (suite *IntegrationTestSuite) TestUpgradeInflation_Disable() {
	afterUpgradeCtx := util.CtxAtHeight(suite.UpgradeHeight)

	// Get x/community for switchover time
	params, err := suite.Kava.Community.Params(afterUpgradeCtx, &communitytypes.QueryParamsRequest{})
	suite.Require().NoError(err)

	// Sleep until switchover time + 6 seconds for extra block
	sleepDuration := time.Until(params.Params.UpgradeTimeDisableInflation.Add(6 * time.Second))
	time.Sleep(sleepDuration)

	suite.Require().Eventually(func() bool {
		communityParams, err := suite.Kava.Community.Params(afterUpgradeCtx, &communitytypes.QueryParamsRequest{})
		suite.Require().NoError(err)

		// After params are set in x/community -- non-zero switchover time
		return !communityParams.Params.UpgradeTimeDisableInflation.Equal(time.Time{})
	}, 20*time.Second, 3*time.Second)

	// Fetch exact block when inflation stop event emitted
	// This is run after the switchover, so we don't need to poll
	_, switchoverHeight, err := suite.Kava.GetBeginBlockEventsFromQuery(
		context.Background(),
		fmt.Sprintf(
			"%s.%s EXISTS",
			communitytypes.EventTypeInflationStop,
			communitytypes.AttributeKeyDisableTime,
		),
	)
	suite.Require().NoError(err)
	suite.Require().NotZero(switchoverHeight)

	afterSwitchoverCtx := util.CtxAtHeight(switchoverHeight)

	suite.Run("x/mint inflation before switchover", func() {
		mintParams, err := suite.Kava.Mint.Params(
			afterUpgradeCtx,
			&minttypes.QueryParamsRequest{},
		)
		suite.NoError(err)
		kavaDistParams, err := suite.Kava.Kavadist.Params(
			afterUpgradeCtx,
			&kavadisttypes.QueryParamsRequest{},
		)
		suite.NoError(err)

		// Use .String() to compare Decs since x/mint uses the deprecated one,
		// mismatch of types but same value.
		suite.Equal(
			sdkmath.LegacyMustNewDecFromStr("0.595000000000000000").String(),
			mintParams.Params.InflationMin.String(),
			"x/mint inflation min should be 59.5%% before switchover",
		)
		suite.Equal(
			sdkmath.LegacyMustNewDecFromStr("0.595000000000000000").String(),
			mintParams.Params.InflationMax.String(),
			"x/mint inflation max should be 59.5%% before switchover",
		)

		suite.True(
			kavaDistParams.Params.Active,
			"x/kavadist should be active before switchover",
		)
	})

	suite.Run("x/mint inflation after switchover", func() {
		mintParams, err := suite.Kava.Mint.Params(
			afterSwitchoverCtx,
			&minttypes.QueryParamsRequest{},
		)
		suite.NoError(err)
		kavaDistParams, err := suite.Kava.Kavadist.Params(
			afterSwitchoverCtx,
			&kavadisttypes.QueryParamsRequest{},
		)
		suite.NoError(err)

		suite.Equal(
			sdkmath.LegacyZeroDec().String(),
			mintParams.Params.InflationMin.String(),
			"x/mint inflation min should be 0% after switchover",
		)
		suite.Equal(
			sdkmath.LegacyZeroDec().String(),
			mintParams.Params.InflationMax.String(),
			"x/mint inflation max should be 0% after switchover",
		)

		suite.False(
			kavaDistParams.Params.Active,
			"x/kavadist should be inactive after switchover",
		)
	})
}
