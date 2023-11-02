package e2e_test

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/tests/util"
	communitytypes "github.com/kava-labs/kava/x/community/types"
	kavadisttypes "github.com/kava-labs/kava/x/kavadist/types"
)

func (suite *IntegrationTestSuite) TestUpgradeCommunityParams() {
	suite.SkipIfUpgradeDisabled()

	beforeUpgradeCtx := util.CtxAtHeight(suite.UpgradeHeight - 1)
	afterUpgradeCtx := util.CtxAtHeight(suite.UpgradeHeight)

	// Before params
	kavaDistParamsBefore, err := suite.Kava.Kavadist.Params(beforeUpgradeCtx, &kavadisttypes.QueryParamsRequest{})
	suite.NoError(err)
	mintParamsBefore, err := suite.Kava.Mint.Params(beforeUpgradeCtx, &minttypes.QueryParamsRequest{})
	suite.NoError(err)

	// Before parameters
	suite.Run("x/community and x/kavadist parameters before upgrade", func() {
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
	suite.Run("x/kavadist, x/mint, x/community parameters after upgrade, before switchover", func() {
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

	suite.Require().Eventually(
		func() bool {
			// Get x/community for switchover time
			params, err := suite.Kava.Community.Params(
				context.Background(),
				&communitytypes.QueryParamsRequest{},
			)
			suite.Require().NoError(err)

			// Check that switchover time is set to zero, e.g. switchover happened
			return params.Params.UpgradeTimeDisableInflation.Equal(time.Time{})
		},
		20*time.Second, 1*time.Second,
		"switchover should happen and x/community params should be updated",
	)

	// Fetch exact block when inflation stop event emitted
	_, switchoverHeight, err := suite.Kava.GetBeginBlockEventsFromQuery(
		context.Background(),
		fmt.Sprintf(
			"%s.%s EXISTS",
			communitytypes.EventTypeInflationStop,
			communitytypes.AttributeKeyInflationDisableTime,
		),
	)
	suite.Require().NoError(err)
	suite.Require().NotZero(switchoverHeight)

	beforeSwitchoverCtx := util.CtxAtHeight(switchoverHeight - 1)
	afterSwitchoverCtx := util.CtxAtHeight(switchoverHeight)

	suite.Run("x/kavadist, x/mint, x/community parameters after upgrade, after switchover", func() {
		kavaDistParamsAfter, err := suite.Kava.Kavadist.Params(
			afterSwitchoverCtx,
			&kavadisttypes.QueryParamsRequest{},
		)
		suite.NoError(err)
		mintParamsAfter, err := suite.Kava.Mint.Params(
			afterSwitchoverCtx,
			&minttypes.QueryParamsRequest{},
		)
		suite.NoError(err)
		communityParamsAfter, err := suite.Kava.Community.Params(
			afterSwitchoverCtx,
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

	suite.Run("x/kavadist, x/distribution, x/community balances after switchover", func() {
		// Before balances - community pool fund consolidation
		kavaDistBalBefore, err := suite.Kava.Kavadist.Balance(
			beforeSwitchoverCtx,
			&kavadisttypes.QueryBalanceRequest{},
		)
		suite.NoError(err)
		distrBalBefore, err := suite.Kava.Distribution.CommunityPool(
			beforeSwitchoverCtx,
			&distrtypes.QueryCommunityPoolRequest{},
		)
		suite.NoError(err)
		distrBalCoinsBefore, distrDustBefore := distrBalBefore.Pool.TruncateDecimal()
		beforeCommPoolBalance, err := suite.Kava.Community.Balance(
			beforeSwitchoverCtx,
			&communitytypes.QueryBalanceRequest{},
		)
		suite.NoError(err)

		// After balances
		kavaDistBalAfter, err := suite.Kava.Kavadist.Balance(
			afterSwitchoverCtx,
			&kavadisttypes.QueryBalanceRequest{},
		)
		suite.NoError(err)
		distrBalAfter, err := suite.Kava.Distribution.CommunityPool(
			afterSwitchoverCtx,
			&distrtypes.QueryCommunityPoolRequest{},
		)
		suite.NoError(err)
		afterCommPoolBalance, err := suite.Kava.Community.Balance(
			afterSwitchoverCtx,
			&communitytypes.QueryBalanceRequest{},
		)
		suite.NoError(err)

		expectedKavadistBal := sdk.NewCoins(sdk.NewCoin(
			"ukava",
			kavaDistBalBefore.Coins.AmountOf("ukava"),
		))
		suite.Equal(
			expectedKavadistBal,
			kavaDistBalAfter.Coins,
			"x/kavadist balance should persist the ukava amount and move all other funds",
		)
		expectedKavadistTransferred := kavaDistBalBefore.Coins.Sub(expectedKavadistBal...)

		// very low ukava balance after (ignoring dust in x/distribution)
		// a small amount of tx fees can still end up here.
		// dust should stay in x/distribution, but may not be the same so it's unchecked
		distrCoinsAfter, distrDustAfter := distrBalAfter.Pool.TruncateDecimal()
		suite.Empty(distrCoinsAfter, "expected no coins in x/distribution community pool")

		// Fetch block results for paid staking rewards in the block
		blockRes, err := suite.Kava.TmSignClient.BlockResults(
			context.Background(),
			&switchoverHeight,
		)
		suite.Require().NoError(err)

		stakingRewardPaidEvents := util.FilterEventsByType(
			blockRes.BeginBlockEvents,
			communitytypes.EventTypeStakingRewardsPaid,
		)
		suite.Require().Len(stakingRewardPaidEvents, 1, "there should be only 1 staking reward paid event")
		stakingRewardAmount := sdk.NewCoins()
		for _, attr := range stakingRewardPaidEvents[0].Attributes {
			if string(attr.Key) == communitytypes.AttributeKeyStakingRewardAmount {
				stakingRewardAmount, err = sdk.ParseCoinsNormalized(string(attr.Value))
				suite.Require().NoError(err)

				break
			}
		}

		expectedCommunityBal := beforeCommPoolBalance.Coins.
			Add(distrBalCoinsBefore...).
			Add(expectedKavadistTransferred...).
			Sub(stakingRewardAmount...) // Remove staking rewards paid in the block

		// x/kavadist and x/distribution community pools should be moved to x/community
		suite.Equal(
			expectedCommunityBal,
			afterCommPoolBalance.Coins,
		)

		suite.Equal(
			distrDustBefore,
			distrDustAfter,
			"x/distribution community pool dust should be unchanged",
		)
	})
}
