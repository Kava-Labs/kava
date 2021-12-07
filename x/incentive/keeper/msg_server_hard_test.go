package keeper_test

import (
	"time"

	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/kava-labs/kava/x/incentive/types"
)

func (suite *HandlerTestSuite) TestPayoutHardClaimMultiDenom() {
	userAddr, receiverAddr := suite.addrs[0], suite.addrs[1]

	authBulder := suite.authBuilder().
		WithSimpleAccount(userAddr, cs(c("bnb", 1e12))).
		WithSimpleAccount(receiverAddr, nil)

	incentBuilder := suite.incentiveBuilder().
		WithSimpleSupplyRewardPeriod("bnb", cs(c("hard", 1e6), c("swap", 1e6))).
		WithSimpleBorrowRewardPeriod("bnb", cs(c("hard", 1e6), c("swap", 1e6)))

	suite.SetupWithGenState(authBulder, incentBuilder)

	// create a deposit and borrow
	suite.NoError(suite.DeliverHardMsgDeposit(userAddr, cs(c("bnb", 1e11))))
	suite.NoError(suite.DeliverHardMsgBorrow(userAddr, cs(c("bnb", 1e10))))

	// accumulate some rewards
	suite.NextBlockAfter(7 * time.Second)

	preClaimBal := suite.GetBalance(userAddr)

	msg := types.NewMsgClaimHardReward(
		userAddr.String(),
		types.Selections{
			types.NewSelection("hard", "small"),
			types.NewSelection("swap", "medium"),
		},
	)

	// Claim denoms
	err := suite.DeliverIncentiveMsg(&msg)
	suite.NoError(err)

	// Check rewards were paid out
	expectedRewardsHard := c("hard", int64(0.2*float64(2*7*1e6)))
	expectedRewardsSwap := c("swap", int64(0.5*float64(2*7*1e6)))
	suite.BalanceEquals(userAddr, preClaimBal.Add(expectedRewardsHard, expectedRewardsSwap))

	suite.VestingPeriodsEqual(userAddr, vestingtypes.Periods{
		{Length: (17+31)*secondsPerDay - 7, Amount: cs(expectedRewardsHard)},
		{Length: (28 + 31 + 30 + 31 + 30) * secondsPerDay, Amount: cs(expectedRewardsSwap)}, // second length is stacked on top of the first
	})
	// Check that claimed coins have been removed from a claim's reward
	suite.HardRewardEquals(userAddr, nil)
}

func (suite *HandlerTestSuite) TestPayoutHardClaimSingleDenom() {
	userAddr := suite.addrs[0]

	authBulder := suite.authBuilder().
		WithSimpleAccount(userAddr, cs(c("bnb", 1e12)))

	incentBuilder := suite.incentiveBuilder().
		WithSimpleSupplyRewardPeriod("bnb", cs(c("hard", 1e6), c("swap", 1e6))).
		WithSimpleBorrowRewardPeriod("bnb", cs(c("hard", 1e6), c("swap", 1e6)))

	suite.SetupWithGenState(authBulder, incentBuilder)

	// create a deposit and borrow
	suite.NoError(suite.DeliverHardMsgDeposit(userAddr, cs(c("bnb", 1e11))))
	suite.NoError(suite.DeliverHardMsgBorrow(userAddr, cs(c("bnb", 1e10))))

	// accumulate some rewards
	suite.NextBlockAfter(7 * time.Second)

	preClaimBal := suite.GetBalance(userAddr)

	msg := types.NewMsgClaimHardReward(
		userAddr.String(),
		types.Selections{
			types.NewSelection("swap", "large"),
		},
	)

	// Claim rewards
	err := suite.DeliverIncentiveMsg(&msg)
	suite.NoError(err)

	// Check rewards were paid out
	expectedRewards := c("swap", 2*7*1e6)
	suite.BalanceEquals(userAddr, preClaimBal.Add(expectedRewards))

	suite.VestingPeriodsEqual(userAddr, vestingtypes.Periods{
		{Length: (17+31+28+31+30+31+30+31+31+30+31+30+31)*secondsPerDay - 7, Amount: cs(expectedRewards)},
	})

	// Check that claimed coins have been removed from a claim's reward
	suite.HardRewardEquals(userAddr, cs(c("hard", 2*7*1e6)))
}
