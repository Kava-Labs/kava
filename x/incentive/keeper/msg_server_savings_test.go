package keeper_test

import (
	"time"

	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/kava-labs/kava/x/incentive/testutil"
	"github.com/kava-labs/kava/x/incentive/types"
)

func (suite *HandlerTestSuite) TestPayoutSavingsClaimMultiDenom() {
	userAddr := suite.addrs[0]

	authBuilder := suite.authBuilder().
		WithSimpleAccount(userAddr, cs(c("ukava", 1e12), c("busd", 1e12)))

	incentBuilder := suite.incentiveBuilder().
		WithSimpleRewardPeriod(
			types.CLAIM_TYPE_SAVINGS,
			"ukava",
			cs(c("hard", 1e6), c("ukava", 1e6)),
		)

	savingsBuilder := testutil.NewSavingsGenesisBuilder().
		WithSupportedDenoms("ukava")

	suite.SetupWithGenState(authBuilder, savingsBuilder, incentBuilder)

	// deposit into a savings pool
	suite.NoError(
		suite.DeliverSavingsMsgDeposit(userAddr, cs(c("ukava", 1e9))),
	)
	// accumulate some savings rewards
	suite.NextBlockAfter(7 * time.Second)

	preClaimBal := suite.GetBalance(userAddr)

	msg := types.NewMsgClaimSavingsReward(
		userAddr.String(),
		types.Selections{
			types.NewSelection("hard", "small"),
			types.NewSelection("ukava", "large"),
		},
	)

	// Claim rewards
	err := suite.DeliverIncentiveMsg(&msg)
	suite.NoError(err)

	// Check rewards were paid out
	expectedRewardsHard := c("hard", int64(0.2*float64(7*1e6)))
	expectedRewardsUkava := c("ukava", 7*1e6)
	suite.BalanceEquals(userAddr, preClaimBal.Add(expectedRewardsHard, expectedRewardsUkava))

	suite.VestingPeriodsEqual(userAddr, []vestingtypes.Period{
		{Length: (17+31)*secondsPerDay - 7, Amount: cs(expectedRewardsHard)},
		{Length: (28 + 31 + 30 + 31 + 30 + 31 + 31 + 30 + 31 + 30 + 31) * secondsPerDay, Amount: cs(expectedRewardsUkava)}, // second length is stacked on top of the first
	})

	// Check that each claim reward coin's amount has been reset to 0
	suite.RewardEquals(types.CLAIM_TYPE_SAVINGS, userAddr, nil)
}

func (suite *HandlerTestSuite) TestPayoutSavingsClaimSingleDenom() {
	userAddr := suite.addrs[0]

	authBuilder := suite.authBuilder().
		WithSimpleAccount(userAddr, cs(c("ukava", 1e12), c("busd", 1e12)))

	incentBuilder := suite.incentiveBuilder().
		WithSimpleRewardPeriod(
			types.CLAIM_TYPE_SAVINGS,
			"ukava",
			cs(c("hard", 1e6), c("ukava", 1e6)),
		)

	savingsBuilder := testutil.NewSavingsGenesisBuilder().WithSupportedDenoms("ukava")

	suite.SetupWithGenState(authBuilder, savingsBuilder, incentBuilder)

	// deposit into savings
	suite.NoError(
		suite.DeliverSavingsMsgDeposit(userAddr, cs(c("ukava", 1e9))),
	)

	// accumulate some savings rewards
	suite.NextBlockAfter(7 * time.Second)

	preClaimBal := suite.GetBalance(userAddr)

	msg := types.NewMsgClaimSavingsReward(
		userAddr.String(),
		types.Selections{
			types.NewSelection("ukava", "large"),
		},
	)

	// Claim rewards
	err := suite.DeliverIncentiveMsg(&msg)
	suite.NoError(err)

	// Check rewards were paid out
	expectedRewards := c("ukava", 7*1e6)
	suite.BalanceEquals(userAddr, preClaimBal.Add(expectedRewards))

	suite.VestingPeriodsEqual(userAddr, vestingtypes.Periods{
		{Length: (17+31+28+31+30+31+30+31+31+30+31+30+31)*secondsPerDay - 7, Amount: cs(expectedRewards)},
	})

	// Check that claimed coins have been removed from a claim's reward
	suite.RewardEquals(types.CLAIM_TYPE_SAVINGS, userAddr, cs(c("hard", 7*1e6)))
}
