package incentive_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"

	"github.com/kava-labs/kava/x/incentive/types"
)

func (suite *HandlerTestSuite) TestPayoutDelegatorClaim() {
	userAddr := suite.addrs[0]
	receiverAddr := suite.addrs[1]

	authBulder := suite.authBuilder().
		WithSimpleAccount(userAddr, cs(c("ukava", 1e12))).
		WithSimpleAccount(receiverAddr, nil)

	incentBuilder := suite.incentiveBuilder().
		WithSimpleDelegatorRewardPeriod(types.BondDenom, cs(c("hard", 1e6)))

	suite.SetupWithGenState(authBulder, incentBuilder)

	// create a delegation (need to create a validator first, which will have a self delegation)
	suite.NoError(
		suite.DeliverMsgCreateValidator(sdk.ValAddress(userAddr), c("ukava", 1e9)),
	)
	// new block required to bond validator
	suite.NextBlockAfter(7 * time.Second)
	// Now the delegation is bonded, accumulate some delegator rewards
	suite.NextBlockAfter(7 * time.Second)

	preClaimBal := suite.GetBalance(userAddr)

	// Check rewards cannot be claimed by vvesting claim msgs
	err := suite.DeliverIncentiveMsg(
		types.NewMsgClaimDelegatorRewardVVesting(userAddr, receiverAddr, "large", nil),
	)
	suite.ErrorIs(err, types.ErrInvalidAccountType)

	// Claim a single denom
	err = suite.DeliverIncentiveMsg(
		types.NewMsgClaimDelegatorReward(userAddr, "large", nil),
	)
	suite.NoError(err)

	// Check rewards were paid out
	expectedRewards := cs(c("hard", 2*7*1e6))
	suite.BalanceEquals(userAddr, preClaimBal.Add(expectedRewards...))

	suite.VestingPeriodsEqual(userAddr, vesting.Periods{
		{Length: 33004786, Amount: expectedRewards},
	})
	// Check that claimed coins have been removed from a claim's reward
	suite.DelegatorRewardEquals(userAddr, nil)
}

func (suite *HandlerTestSuite) TestPayoutDelegatorClaimSingleDenom() {
	userAddr := suite.addrs[0]

	authBulder := suite.authBuilder().
		WithSimpleAccount(userAddr, cs(c("ukava", 1e12)))

	incentBuilder := suite.incentiveBuilder().
		WithSimpleDelegatorRewardPeriod(types.BondDenom, cs(c("hard", 1e6), c("swap", 1e6)))

	suite.SetupWithGenState(authBulder, incentBuilder)

	// create a delegation (need to create a validator first, which will have a self delegation)
	suite.NoError(
		suite.DeliverMsgCreateValidator(sdk.ValAddress(userAddr), c("ukava", 1e9)),
	)
	// new block required to bond validator
	suite.NextBlockAfter(7 * time.Second)
	// Now the delegation is bonded, accumulate some delegator rewards
	suite.NextBlockAfter(7 * time.Second)

	preClaimBal := suite.GetBalance(userAddr)

	// Check rewards cannot be claimed by vvesting claim msgs
	err := suite.DeliverIncentiveMsg(
		types.NewMsgClaimDelegatorRewardVVesting(userAddr, suite.addrs[1], "large", nil),
	)
	suite.ErrorIs(err, types.ErrInvalidAccountType)

	// Claim rewards
	err = suite.DeliverIncentiveMsg(
		types.NewMsgClaimDelegatorReward(userAddr, "large", []string{"swap"}),
	)
	suite.NoError(err)

	// Check rewards were paid out
	expectedRewards := c("swap", 2*7*1e6)
	suite.BalanceEquals(userAddr, preClaimBal.Add(expectedRewards))

	suite.VestingPeriodsEqual(userAddr, vesting.Periods{
		{Length: 33004786, Amount: cs(expectedRewards)},
	})

	// Check that claimed coins have been removed from a claim's reward
	suite.DelegatorRewardEquals(userAddr, cs(c("hard", 2*7*1e6)))
}

func (suite *HandlerTestSuite) TestPayoutDelegatorClaimVVesting() {
	valAddr, receiverAddr := suite.addrs[0], suite.addrs[1]

	vva := suite.NewValidatorVestingAccountWithBalance(valAddr, cs(c("ukava", 1e12)))

	authBulder := suite.authBuilder().
		WithAccounts(vva).
		WithSimpleAccount(receiverAddr, nil)

	incentBuilder := suite.incentiveBuilder().
		WithSimpleDelegatorRewardPeriod(types.BondDenom, cs(c("hard", 1e6)))

	suite.SetupWithGenState(authBulder, incentBuilder)

	// create a delegation (need to create a validator first, which will have a self delegation)
	suite.NoError(
		suite.DeliverMsgCreateValidator(sdk.ValAddress(valAddr), c("ukava", 1e9)),
	)
	// new block required to bond validator
	suite.NextBlockAfter(7 * time.Second)
	// Now the delegation is bonded, accumulate some delegator rewards
	suite.NextBlockAfter(7 * time.Second)

	preClaimBal := suite.GetBalance(receiverAddr)

	// Check rewards cannot be claimed by normal claim msgs
	err := suite.DeliverIncentiveMsg(
		types.NewMsgClaimDelegatorReward(valAddr, "large", nil),
	)
	suite.ErrorIs(err, types.ErrInvalidAccountType)

	// Claim rewards
	err = suite.DeliverIncentiveMsg(
		types.NewMsgClaimDelegatorRewardVVesting(valAddr, receiverAddr, "large", nil),
	)
	suite.NoError(err)

	// Check rewards were paid out
	expectedRewards := c("hard", 2*7*1e6)
	suite.BalanceEquals(receiverAddr, preClaimBal.Add(expectedRewards))

	suite.VestingPeriodsEqual(receiverAddr, vesting.Periods{
		{Length: 33004786, Amount: cs(expectedRewards)},
	})

	// Check that each claim reward coin's amount has been reset to 0
	suite.DelegatorRewardEquals(valAddr, nil)
}

func (suite *HandlerTestSuite) TestPayoutDelegatorClaimVVestingSingleDenom() {
	valAddr, receiverAddr := suite.addrs[0], suite.addrs[1]

	vva := suite.NewValidatorVestingAccountWithBalance(valAddr, cs(c("ukava", 1e12)))

	authBulder := suite.authBuilder().
		WithAccounts(vva).
		WithSimpleAccount(receiverAddr, nil)

	incentBuilder := suite.incentiveBuilder().
		WithSimpleDelegatorRewardPeriod(types.BondDenom, cs(c("hard", 1e6), c("swap", 1e6)))

	suite.SetupWithGenState(authBulder, incentBuilder)

	// create a delegation (need to create a validator first, which will have a self delegation)
	suite.NoError(
		suite.DeliverMsgCreateValidator(sdk.ValAddress(valAddr), c("ukava", 1e9)),
	)
	// new block required to bond validator
	suite.NextBlockAfter(7 * time.Second)
	// Now the delegation is bonded, accumulate some delegator rewards
	suite.NextBlockAfter(7 * time.Second)

	preClaimBal := suite.GetBalance(receiverAddr)

	// Check rewards cannot be claimed by normal claim msgs
	err := suite.DeliverIncentiveMsg(
		types.NewMsgClaimDelegatorReward(valAddr, "large", []string{"swap"}),
	)
	suite.ErrorIs(err, types.ErrInvalidAccountType)

	// Claim rewards
	err = suite.DeliverIncentiveMsg(
		types.NewMsgClaimDelegatorRewardVVesting(valAddr, receiverAddr, "large", []string{"swap"}),
	)
	suite.NoError(err)

	// Check rewards were paid out
	expectedRewards := c("swap", 2*7*1e6)
	suite.BalanceEquals(receiverAddr, preClaimBal.Add(expectedRewards))

	suite.VestingPeriodsEqual(receiverAddr, vesting.Periods{
		{Length: 33004786, Amount: cs(expectedRewards)},
	})

	// Check that claimed coins have been removed from a claim's reward
	suite.DelegatorRewardEquals(valAddr, cs(c("hard", 2*7*1e6)))
}
