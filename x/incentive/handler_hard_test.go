package incentive_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/x/auth/vesting"

	"github.com/kava-labs/kava/x/incentive/types"
)

func (suite *HandlerTestSuite) TestPayoutHardClaim() {
	userAddr, receiverAddr := suite.addrs[0], suite.addrs[1]

	authBulder := suite.authBuilder().
		WithSimpleAccount(userAddr, cs(c("bnb", 1e12))).
		WithSimpleAccount(receiverAddr, nil)

	incentBuilder := suite.incentiveBuilder().
		WithSimpleSupplyRewardPeriod("bnb", cs(c("hard", 1e6))).
		WithSimpleBorrowRewardPeriod("bnb", cs(c("hard", 1e6)))

	suite.SetupWithGenState(authBulder, incentBuilder)

	// create a deposit and borrow
	suite.NoError(suite.DeliverHardMsgDeposit(userAddr, cs(c("bnb", 1e11))))
	suite.NoError(suite.DeliverHardMsgBorrow(userAddr, cs(c("bnb", 1e10))))

	// accumulate some rewards
	suite.NextBlockAfter(7 * time.Second)

	preClaimBal := suite.GetBalance(userAddr)

	// Check rewards cannot be claimed by vvesting claim msgs
	err := suite.DeliverIncentiveMsg(
		types.NewMsgClaimHardRewardVVesting(userAddr, receiverAddr, types.NewSelection("hard", "large")),
	)
	suite.ErrorIs(err, types.ErrInvalidAccountType)

	// Claim a single denom
	err = suite.DeliverIncentiveMsg(
		types.NewMsgClaimHardReward(userAddr, types.NewSelection("hard", "large")),
	)
	suite.NoError(err)

	// Check rewards were paid out
	expectedRewards := cs(c("hard", 2*7*1e6))
	suite.BalanceEquals(userAddr, preClaimBal.Add(expectedRewards...))

	suite.VestingPeriodsEqual(userAddr, vesting.Periods{
		{Length: 33004793, Amount: expectedRewards},
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

	// Check rewards cannot be claimed by vvesting claim msgs
	err := suite.DeliverIncentiveMsg(
		types.NewMsgClaimHardRewardVVesting(userAddr, suite.addrs[1], types.NewSelection("swap", "large")),
	)
	suite.ErrorIs(err, types.ErrInvalidAccountType)

	// Claim rewards
	err = suite.DeliverIncentiveMsg(
		types.NewMsgClaimHardReward(userAddr, types.NewSelection("swap", "large")),
	)
	suite.NoError(err)

	// Check rewards were paid out
	expectedRewards := c("swap", 2*7*1e6)
	suite.BalanceEquals(userAddr, preClaimBal.Add(expectedRewards))

	suite.VestingPeriodsEqual(userAddr, vesting.Periods{
		{Length: 33004793, Amount: cs(expectedRewards)},
	})

	// Check that claimed coins have been removed from a claim's reward
	suite.HardRewardEquals(userAddr, cs(c("hard", 2*7*1e6)))
}

func (suite *HandlerTestSuite) TestPayoutHardClaimVVesting() {
	valAddr, receiverAddr := suite.addrs[0], suite.addrs[1]

	vva := suite.NewValidatorVestingAccountWithBalance(valAddr, cs(c("bnb", 1e12)))

	authBulder := suite.authBuilder().
		WithAccounts(vva).
		WithSimpleAccount(receiverAddr, nil)

	incentBuilder := suite.incentiveBuilder().
		WithSimpleSupplyRewardPeriod("bnb", cs(c("hard", 1e6))).
		WithSimpleBorrowRewardPeriod("bnb", cs(c("hard", 1e6)))

	suite.SetupWithGenState(authBulder, incentBuilder)

	// create a deposit and borrow
	suite.NoError(suite.DeliverHardMsgDeposit(valAddr, cs(c("bnb", 1e11))))
	suite.NoError(suite.DeliverHardMsgBorrow(valAddr, cs(c("bnb", 1e10))))

	// accumulate some rewards
	suite.NextBlockAfter(7 * time.Second)

	preClaimBal := suite.GetBalance(receiverAddr)

	// Check rewards cannot be claimed by normal claim msgs
	err := suite.DeliverIncentiveMsg(
		types.NewMsgClaimHardReward(valAddr, types.NewSelection("hard", "large")),
	)
	suite.ErrorIs(err, types.ErrInvalidAccountType)

	// Claim rewards
	err = suite.DeliverIncentiveMsg(
		types.NewMsgClaimHardRewardVVesting(valAddr, receiverAddr, types.NewSelection("hard", "large")),
	)
	suite.NoError(err)

	// Check rewards were paid out
	expectedRewards := c("hard", 2*7*1e6)
	suite.BalanceEquals(receiverAddr, preClaimBal.Add(expectedRewards))

	suite.VestingPeriodsEqual(receiverAddr, vesting.Periods{
		{Length: 33004793, Amount: cs(expectedRewards)},
	})

	// Check that each claim reward coin's amount has been reset to 0
	suite.HardRewardEquals(valAddr, nil)
}

func (suite *HandlerTestSuite) TestPayoutHardClaimVVestingSingleDenom() {
	valAddr, receiverAddr := suite.addrs[0], suite.addrs[1]

	vva := suite.NewValidatorVestingAccountWithBalance(valAddr, cs(c("bnb", 1e12)))

	authBulder := suite.authBuilder().
		WithAccounts(vva).
		WithSimpleAccount(receiverAddr, nil)

	incentBuilder := suite.incentiveBuilder().
		WithSimpleSupplyRewardPeriod("bnb", cs(c("hard", 1e6), c("swap", 1e6))).
		WithSimpleBorrowRewardPeriod("bnb", cs(c("hard", 1e6), c("swap", 1e6)))

	suite.SetupWithGenState(authBulder, incentBuilder)

	// create a deposit and borrow
	suite.NoError(suite.DeliverHardMsgDeposit(valAddr, cs(c("bnb", 1e11))))
	suite.NoError(suite.DeliverHardMsgBorrow(valAddr, cs(c("bnb", 1e10))))

	// accumulate some rewards
	suite.NextBlockAfter(7 * time.Second)

	preClaimBal := suite.GetBalance(receiverAddr)

	// Check rewards cannot be claimed by normal claim msgs
	err := suite.DeliverIncentiveMsg(
		types.NewMsgClaimHardReward(valAddr, types.NewSelection("swap", "large")),
	)
	suite.ErrorIs(err, types.ErrInvalidAccountType)

	// Claim rewards
	err = suite.DeliverIncentiveMsg(
		types.NewMsgClaimHardRewardVVesting(valAddr, receiverAddr, types.NewSelection("swap", "large")),
	)
	suite.NoError(err)

	// Check rewards were paid out
	expectedRewards := c("swap", 2*7*1e6)
	suite.BalanceEquals(receiverAddr, preClaimBal.Add(expectedRewards))

	suite.VestingPeriodsEqual(receiverAddr, vesting.Periods{
		{Length: 33004793, Amount: cs(expectedRewards)},
	})

	// Check that claimed coins have been removed from a claim's reward
	suite.HardRewardEquals(valAddr, cs(c("hard", 2*7*1e6)))
}
