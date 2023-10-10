package e2e_test

import (
	"context"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types/tx"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/kava-labs/kava/tests/util"
	communitytypes "github.com/kava-labs/kava/x/community/types"
	kavadisttypes "github.com/kava-labs/kava/x/kavadist/types"
)

func (suite *IntegrationTestSuite) TestUpgradeInflation_Disable() {
	beforeUpgradeCtx := util.CtxAtHeight(suite.UpgradeHeight - 1)
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
	// TODO: Not in a tx, need to query abci beginblock events
	suite.Kava.Tx.GetTxsEvent(context.Background(), &tx.GetTxsEventRequest{
		Events: []string{
			fmt.Sprintf("tm.event=%s", communitytypes.EventTypeInflationStop),
		},
	})

	suite.Run("x/mint inflation before switchover", func() {
		mintParams, err := suite.Kava.Mint.Params(
			context.Background(),
			&minttypes.QueryParamsRequest{},
		)
		suite.NoError(err)
		kavaDistParams, err := suite.Kava.Kavadist.Params(
			context.Background(),
			&kavadisttypes.QueryParamsRequest{},
		)
		suite.NoError(err)

		suite.Equal(
			sdkmath.LegacyMustNewDecFromStr("0.595000000000000000"),
			mintParams.Params.InflationMin,
			"x/mint inflation min should be 59.5%% before switchover",
		)
		suite.Equal(
			sdkmath.LegacyMustNewDecFromStr("0.595000000000000000"),
			mintParams.Params.InflationMax,
			"x/mint inflation max should be 59.5%% before switchover",
		)

		suite.False(
			kavaDistParams.Params.Active,
			"x/kavadist should be inactive after switchover",
		)
	})
}
