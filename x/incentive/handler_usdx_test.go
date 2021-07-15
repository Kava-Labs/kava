package incentive_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/x/auth/vesting"

	"github.com/kava-labs/kava/x/incentive/types"
)

func (suite *HandlerTestSuite) TestPayoutUSDXClaim() {
	userAddr, receiverAddr := suite.addrs[0], suite.addrs[1]

	authBulder := suite.authBuilder().
		WithSimpleAccount(userAddr, cs(c("bnb", 1e12))).
		WithSimpleAccount(receiverAddr, nil)

	incentBuilder := suite.incentiveBuilder().
		WithSimpleUSDXRewardPeriod("bnb-a", c(types.USDXMintingRewardDenom, 1e6))

	suite.SetupWithGenState(authBulder, incentBuilder)

	// mint some usdx
	suite.DeliverMsgCreateCDP(userAddr, c("bnb", 1e9), c("usdx", 1e7), "bnb-a")
	// accumulate some rewards
	suite.NextBlockAfter(7 * time.Second)

	preClaimBal := suite.GetBalance(userAddr)

	// Check rewards cannot be claimed by vvesting claim msgs
	err := suite.DeliverIncentiveMsg(
		types.NewMsgClaimUSDXMintingRewardVVesting(userAddr, receiverAddr, "large"),
	)
	suite.ErrorIs(err, types.ErrInvalidAccountType)

	// Claim a single denom
	err = suite.DeliverIncentiveMsg(
		types.NewMsgClaimUSDXMintingReward(userAddr, "large"),
	)
	suite.NoError(err)

	// Check rewards were paid out
	expectedRewards := cs(c(types.USDXMintingRewardDenom, 7*1e6))
	suite.BalanceEquals(userAddr, preClaimBal.Add(expectedRewards...))

	suite.VestingPeriodsEqual(userAddr, vesting.Periods{
		{Length: 33004793, Amount: expectedRewards},
	})
	// Check that claimed coins have been removed from a claim's reward
	suite.USDXRewardEquals(userAddr, c(types.USDXMintingRewardDenom, 0))
}

func (suite *HandlerTestSuite) TestPayoutUSDXClaimVVesting() {
	valAddr, receiverAddr := suite.addrs[0], suite.addrs[1]

	vva := suite.NewValidatorVestingAccountWithBalance(valAddr, cs(c("bnb", 1e12)))

	authBulder := suite.authBuilder().
		WithAccounts(vva).
		WithSimpleAccount(receiverAddr, nil)

	incentBuilder := suite.incentiveBuilder().
		WithSimpleUSDXRewardPeriod("bnb-a", c(types.USDXMintingRewardDenom, 1e6))

	suite.SetupWithGenState(authBulder, incentBuilder)

	// mint some usdx
	suite.DeliverMsgCreateCDP(valAddr, c("bnb", 1e9), c("usdx", 1e7), "bnb-a")
	// accumulate some rewards
	suite.NextBlockAfter(7 * time.Second)

	preClaimBal := suite.GetBalance(receiverAddr)

	// Check rewards cannot be claimed by normal claim msgs
	err := suite.DeliverIncentiveMsg(
		types.NewMsgClaimUSDXMintingReward(valAddr, "large"),
	)
	suite.ErrorIs(err, types.ErrInvalidAccountType)

	// Claim rewards
	err = suite.DeliverIncentiveMsg(
		types.NewMsgClaimUSDXMintingRewardVVesting(valAddr, receiverAddr, "large"),
	)
	suite.NoError(err)

	// Check rewards were paid out
	expectedRewards := c(types.USDXMintingRewardDenom, 7*1e6)
	suite.BalanceEquals(receiverAddr, preClaimBal.Add(expectedRewards))

	suite.VestingPeriodsEqual(receiverAddr, vesting.Periods{
		{Length: 33004793, Amount: cs(expectedRewards)},
	})

	// Check that each claim reward coin's amount has been reset to 0
	suite.USDXRewardEquals(valAddr, c(types.USDXMintingRewardDenom, 0))
}
