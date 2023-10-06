package e2e_test

import (
	"context"
	"time"

	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/tests/util"
	communitytypes "github.com/kava-labs/kava/x/community/types"
	kavadisttypes "github.com/kava-labs/kava/x/kavadist/types"
)

func (suite *IntegrationTestSuite) TestDisableInflationOnUpgrade() {
	beforeUpgradeCtx := util.CtxAtHeight(suite.UpgradeHeight - 1)
	afterUpgradeCtx := util.CtxAtHeight(suite.UpgradeHeight)

	// Before balances - pending community pool fund consolidation
	/*
		kavaDistBalBefore, err := suite.Kava.Kavadist.Balance(beforeUpgradeCtx, &kavadisttypes.QueryBalanceRequest{})
		suite.NoError(err)
		distrBalBefore, err := suite.Kava.Distribution.CommunityPool(beforeUpgradeCtx, &distrtypes.QueryCommunityPoolRequest{})
		suite.NoError(err)
		distrBalCoinsBefore, distrBalDustBefore := distrBalBefore.Pool.TruncateDecimal()
		beforeCommPoolBalance, err := suite.Kava.Community.Balance(beforeUpgradeCtx, &communitytypes.QueryBalanceRequest{})
		suite.NoError(err)
	*/

	// Before params
	kavaDistParamsBefore, err := suite.Kava.Kavadist.Params(beforeUpgradeCtx, &kavadisttypes.QueryParamsRequest{})
	suite.NoError(err)
	mintParamsBefore, err := suite.Kava.Mint.Params(beforeUpgradeCtx, &minttypes.QueryParamsRequest{})
	suite.NoError(err)

	// Before parameters
	suite.Run("x/distribution and x/kavadist parameters before upgrade", func() {
		_, err = suite.Kava.Community.Params(beforeUpgradeCtx, &communitytypes.QueryParamsRequest{})
		suite.Error(err, "x/community should not have params before upgrade")

		suite.Require().True(
			kavaDistParamsBefore.Params.Active,
			"x/kavadist should be active before upgrade",
		)

		suite.Require().True(
			mintParamsBefore.Params.InflationMax.IsPositive(),
			"x/mint inflation max should be positive before upgrade",
		)
		suite.Require().True(
			mintParamsBefore.Params.InflationMin.IsPositive(),
			"x/mint inflation min should be positive before upgrade",
		)
	})

	// After upgrade, Before switchover - parameters
	suite.Run("x/distribution and x/kavadist parameters after upgrade, before switchover", func() {
		kavaDistParamsAfter, err := suite.Kava.Kavadist.Params(afterUpgradeCtx, &kavadisttypes.QueryParamsRequest{})
		suite.NoError(err)
		mintParamsAfter, err := suite.Kava.Mint.Params(afterUpgradeCtx, &minttypes.QueryParamsRequest{})
		suite.NoError(err)
		communityParamsAfter, err := suite.Kava.Community.Params(afterUpgradeCtx, &communitytypes.QueryParamsRequest{})
		suite.NoError(err)

		suite.Equal(
			kavaDistParamsBefore.Params,
			kavaDistParamsAfter.Params,
			"x/kavadist should be unaffected after upgrade",
		)

		suite.Equal(
			mintParamsBefore.Params,
			mintParamsAfter.Params,
			"x/mint params should be unaffected after upgrade",
		)

		expectedParams := app.CommunityParams_E2E
		// Make UpgradeTimeDisableInflation match so that we ignore it, because
		// referencing app.CommunityParams_E2E in this test files is different
		// from the one set in the upgrade handler. At least check that it is
		// set to a non-zero value in the assertion below
		expectedParams.UpgradeTimeDisableInflation = communityParamsAfter.Params.UpgradeTimeDisableInflation

		suite.False(
			communityParamsAfter.Params.UpgradeTimeDisableInflation.IsZero(),
			"x/community switchover time should be set after upgrade",
		)
		suite.Equal(
			expectedParams,
			communityParamsAfter.Params,
			"x/community params should be set to E2E params after upgrade",
		)
	})

	// Get x/community for switchover time
	params, err := suite.Kava.Community.Params(afterUpgradeCtx, &communitytypes.QueryParamsRequest{})
	suite.Require().NoError(err)

	// Sleep until switchover time + 6 seconds for extra block
	sleepDuration := time.Until(params.Params.UpgradeTimeDisableInflation.Add(6 * time.Second))
	time.Sleep(sleepDuration)

	suite.Run("x/distribution and x/kavadist parameters after upgrade, after switchover", func() {
		kavaDistParamsAfter, err := suite.Kava.Kavadist.Params(
			context.Background(),
			&kavadisttypes.QueryParamsRequest{},
		)
		suite.NoError(err)
		mintParamsAfter, err := suite.Kava.Mint.Params(
			context.Background(),
			&minttypes.QueryParamsRequest{},
		)
		suite.NoError(err)
		communityParamsAfter, err := suite.Kava.Community.Params(
			context.Background(),
			&communitytypes.QueryParamsRequest{},
		)
		suite.NoError(err)

		suite.False(
			kavaDistParamsAfter.Params.Active,
			"x/kavadist should be disabled after upgrade",
		)

		suite.True(
			mintParamsAfter.Params.InflationMax.IsZero(),
			"x/mint inflation max should be zero after switchover",
		)
		suite.True(
			mintParamsAfter.Params.InflationMin.IsZero(),
			"x/mint inflation min should be zero after switchover",
		)

		suite.Equal(
			time.Time{},
			communityParamsAfter.Params.UpgradeTimeDisableInflation,
			"x/community switchover time should be reset",
		)

		suite.Equal(
			communityParamsAfter.Params.UpgradeTimeSetStakingRewardsPerSecond,
			communityParamsAfter.Params.StakingRewardsPerSecond,
			"x/community staking rewards per second should match upgrade time staking rewards per second",
		)
	})

	/* TODO: Pending community pool fund consolidation
	suite.Run("x/distribution and x/kavadist balances after switchover", func() {
		// After balances
		kavaDistBalAfter, err := suite.Kava.Kavadist.Balance(
			context.Background(),
			&kavadisttypes.QueryBalanceRequest{},
		)
		suite.NoError(err)
		distrBalAfter, err := suite.Kava.Distribution.CommunityPool(
			context.Background(),
			&distrtypes.QueryCommunityPoolRequest{},
		)
		suite.NoError(err)
		afterCommPoolBalance, err := suite.Kava.Community.Balance(
			context.Background(),
			&communitytypes.QueryBalanceRequest{},
		)
		suite.NoError(err)

		// expect empty balances after (ignoring dust in x/distribution)
		suite.Equal(sdk.NewCoins(), kavaDistBalAfter.Coins)
		distrCoinsAfter, distrBalDustAfter := distrBalAfter.Pool.TruncateDecimal()
		suite.Equal(sdk.NewCoins(), distrCoinsAfter)

		// x/kavadist and x/distribution community pools should be moved to x/community
		suite.Equal(
			beforeCommPoolBalance.Coins.
				Add(kavaDistBalBefore.Coins...).
				Add(distrBalCoinsBefore...),
			afterCommPoolBalance.Coins,
		)

		// x/distribution dust should stay in x/distribution
		suite.Equal(distrBalDustBefore, distrBalDustAfter)
	})
	*/
}
