package incentive_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/incentive/testutil"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/kava-labs/kava/x/kavadist"
)

func (suite HandlerTestSuite) DelegatorRewardEquals(owner sdk.AccAddress, expected sdk.Coins) {
	claim, found := suite.keeper.GetDelegatorClaim(suite.ctx, owner)
	suite.Require().Truef(found, "expected delegator claim to be found for %s", owner)
	suite.Equalf(expected, claim.Reward, "expected delegator claim reward to be %s, but got %s", expected, claim.Reward)
}

func (suite *HandlerTestSuite) TestPayoutDelegatorClaim() {
	userAddr := suite.addrs[0]
	receiverAddr := suite.addrs[1]

	authBulder := app.NewAuthGenesisBuilder().
		WithSimpleAccount(userAddr, cs(c("ukava", 1e12))).
		WithSimpleAccount(receiverAddr, cs(c("ukava", 1e12))).
		WithSimpleModuleAccount(kavadist.ModuleName, cs(c("bnb", 1e18), c("hard", 1e18), c("swap", 1e18)))

	incentBuilder := testutil.NewIncentiveGenesisBuilder().
		WithGenesisTime(suite.genesisTime).
		WithMultipliers(types.Multipliers{
			types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0")),
		}).
		WithSimpleDelegatorRewardPeriod(types.BondDenom, cs(c("bnb", 1e6), c("hard", 1e6), c("swap", 1e6)))

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
	failMsg := types.NewMsgClaimDelegatorRewardVVesting(userAddr, receiverAddr, "large", nil)
	_, err := suite.handler(suite.ctx, failMsg)
	suite.ErrorIs(err, types.ErrInvalidAccountType)

	// Claim a single denom
	msg := types.NewMsgClaimDelegatorReward(userAddr, "large", []string{"swap"})
	_, err = suite.handler(suite.ctx, msg)
	suite.NoError(err)

	// Check rewards were paid out
	expectedRewards := cs(c("swap", 2*7*1e6))
	suite.BalanceEquals(userAddr, preClaimBal.Add(expectedRewards...))

	suite.VestingPeriodsEqual(userAddr, vesting.Periods{
		{Length: 33004786, Amount: expectedRewards},
	})
	// Check that claimed coins have been removed from a claim's reward
	suite.DelegatorRewardEquals(userAddr, cs(c("hard", 2*7*1e6), c("bnb", 2*7*1e6)))

	// Claim all remaining denoms
	preClaimBal = suite.GetBalance(userAddr)

	msg = types.NewMsgClaimDelegatorReward(userAddr, "large", nil)
	_, err = suite.handler(suite.ctx, msg)
	suite.NoError(err)

	// Check rewards were paid out
	otherExpectedRewards := cs(c("hard", 2*7*1e6), c("bnb", 2*7*1e6))
	suite.BalanceEquals(userAddr, preClaimBal.Add(otherExpectedRewards...))

	suite.VestingPeriodsEqual(userAddr, vesting.Periods{
		{Length: 33004786, Amount: expectedRewards.Add(otherExpectedRewards...)},
	})

	// Check that each claim reward coin's amount has been reset to 0
	suite.DelegatorRewardEquals(userAddr, nil)
}

func (suite *HandlerTestSuite) TestPayoutDelegatorClaimSingleDenom() {
	userAddr := suite.addrs[0]

	suite.SetupWithGenState(

		app.NewAuthGenesisBuilder().
			WithSimpleAccount(userAddr, cs(c("ukava", 1e12))).
			WithSimpleModuleAccount(kavadist.ModuleName, cs(c("hard", 1e18), c("swap", 1e18))),

		testutil.NewIncentiveGenesisBuilder().
			WithGenesisTime(suite.genesisTime).
			WithMultipliers(types.Multipliers{
				types.NewMultiplier("large", 12, d("1.0")),
			}).
			WithSimpleDelegatorRewardPeriod(types.BondDenom, cs(c("hard", 1e6), c("swap", 1e6))),
	)

	// create a delegation (need to create a validator first, which will have a self delegation)
	suite.NoError(
		suite.DeliverMsgCreateValidator(sdk.ValAddress(userAddr), c("ukava", 1e9)),
	)
	// new block required to bond validator
	suite.NextBlockAfter(7 * time.Second)
	// Now the delegation is bonded, accumulate some delegator rewards
	suite.NextBlockAfter(7 * time.Second)

	preClaimBal := suite.GetBalance(userAddr)

	// Claim rewards
	msg := types.NewMsgClaimDelegatorReward(userAddr, "large", []string{"swap"})
	_, err := suite.handler(suite.ctx, msg)
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
	valAddr := suite.addrs[0]
	receiverAddr := suite.addrs[1]

	vva := suite.NewValidatorVestingAccountWithBalance(valAddr, cs(c("ukava", 1e12)))

	authBulder := app.NewAuthGenesisBuilder().
		WithAccounts(vva).
		WithSimpleAccount(receiverAddr, nil).
		WithSimpleModuleAccount(kavadist.ModuleName, cs(c("hard", 1e18)))

	incentBuilder := testutil.NewIncentiveGenesisBuilder().
		WithGenesisTime(suite.genesisTime).
		WithMultipliers(types.Multipliers{
			types.NewMultiplier(types.MultiplierName("large"), 12, d("1.0")),
		}).
		WithSimpleDelegatorRewardPeriod(types.BondDenom, cs(c("hard", 1e6)))

	suite.SetupWithGenState(authBulder, incentBuilder)

	// create a delegation (need to create a validator first, which will have a self delegation)
	suite.NoError(
		suite.DeliverMsgCreateValidator(sdk.ValAddress(valAddr), c("ukava", 1e9)),
	)
	suite.NextBlockAfter(7 * time.Second) // new block required to bond validator

	// Now the delegation is bonded, accumulate some delegator rewards
	suite.NextBlockAfter(7 * time.Second)

	preClaimBal := suite.GetBalance(receiverAddr)

	// Check rewards cannot be claimed by normal claim msgs
	failMsg := types.NewMsgClaimDelegatorReward(valAddr, "large", nil)
	_, err := suite.handler(suite.ctx, failMsg)
	suite.ErrorIs(err, types.ErrInvalidAccountType)

	// Claim the delegation rewards
	msg := types.NewMsgClaimDelegatorRewardVVesting(valAddr, receiverAddr, "large", nil)
	_, err = suite.handler(suite.ctx, msg)
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
