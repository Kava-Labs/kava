package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	earntypes "github.com/kava-labs/kava/x/earn/types"
	"github.com/kava-labs/kava/x/incentive/testutil"
	"github.com/kava-labs/kava/x/incentive/types"
	liquidtypes "github.com/kava-labs/kava/x/liquid/types"
)

func (suite *HandlerTestSuite) TestEarnLiquidClaim() {
	userAddr1, userAddr2, validatorAddr1, validatorAddr2 := suite.addrs[0], suite.addrs[1], suite.addrs[2], suite.addrs[3]
	valAddr1 := sdk.ValAddress(validatorAddr1)
	valAddr2 := sdk.ValAddress(validatorAddr2)

	bkavaDenom1 := fmt.Sprintf("bkava-%s", valAddr1.String())
	bkavaDenom2 := fmt.Sprintf("bkava-%s", valAddr2.String())

	authBuilder := suite.authBuilder().
		WithSimpleAccount(userAddr1, cs(c("ukava", 1e12))).
		WithSimpleAccount(userAddr2, cs(c("ukava", 1e12))).
		WithSimpleAccount(validatorAddr1, cs(c("ukava", 1e12))).
		WithSimpleAccount(validatorAddr2, cs(c("ukava", 1e12)))

	incentBuilder := suite.incentiveBuilder()

	savingsBuilder := testutil.NewSavingsGenesisBuilder().
		WithSupportedDenoms("bkava")

	earnBuilder := suite.earnBuilder().
		WithVault(earntypes.AllowedVault{
			Denom:             "bkava",
			Strategies:        earntypes.StrategyTypes{earntypes.STRATEGY_TYPE_SAVINGS},
			IsPrivateVault:    false,
			AllowedDepositors: nil,
		})

	suite.SetupWithGenState(authBuilder, incentBuilder, earnBuilder, savingsBuilder)

	err := suite.App.FundModuleAccount(suite.Ctx, distrtypes.ModuleName, cs(c("ukava", 1e12)))
	suite.NoError(err)

	// Create validators
	err = suite.DeliverMsgCreateValidator(valAddr1, c("ukava", 1e9))
	suite.Require().NoError(err)

	err = suite.DeliverMsgCreateValidator(valAddr2, c("ukava", 1e9))
	suite.Require().NoError(err)

	// new block required to bond validator
	suite.NextBlockAfter(7 * time.Second)
	// Now the delegation is bonded, accumulate some delegator rewards
	suite.NextBlockAfter(7 * time.Second)

	// Create delegations from users
	// User 1: 1e9 ukava to validator 1
	// User 2: 99e9 ukava to validator 1 AND 2
	err = suite.DeliverMsgDelegate(userAddr1, valAddr1, c("ukava", 1e9))
	suite.Require().NoError(err)

	err = suite.DeliverMsgDelegate(userAddr2, valAddr1, c("ukava", 99e9))
	suite.Require().NoError(err)

	err = suite.DeliverMsgDelegate(userAddr2, valAddr2, c("ukava", 99e9))
	suite.Require().NoError(err)

	// Mint liquid tokens
	err = suite.DeliverMsgMintDerivative(userAddr1, valAddr1, c("ukava", 1e9))
	suite.Require().NoError(err)

	err = suite.DeliverMsgMintDerivative(userAddr2, valAddr1, c("ukava", 99e9))
	suite.Require().NoError(err)

	err = suite.DeliverMsgMintDerivative(userAddr2, valAddr2, c("ukava", 99e9))
	suite.Require().NoError(err)

	// Deposit liquid tokens to earn
	err = suite.DeliverEarnMsgDeposit(userAddr1, c(bkavaDenom1, 1e9), earntypes.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	err = suite.DeliverEarnMsgDeposit(userAddr2, c(bkavaDenom1, 99e9), earntypes.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)
	err = suite.DeliverEarnMsgDeposit(userAddr2, c(bkavaDenom2, 99e9), earntypes.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	// Accumulate some staking rewards
	// _ = suite.App.EndBlocker(suite.Ctx, abci.RequestEndBlock{})
	// suite.Ctx = suite.Ctx.WithBlockTime(time.Now().Add(1 * time.Hour))
	// suite.App.BeginBlocker(suite.Ctx, abci.RequestBeginBlock{})
	sk := suite.App.GetStakingKeeper()
	dk := suite.App.GetDistrKeeper()
	ik := suite.App.GetIncentiveKeeper()

	rewardCoins := sdk.NewDecCoins(sdk.NewDecCoin("ukava", sdk.NewInt(500e6)))
	validator1, found := sk.GetValidator(suite.Ctx, valAddr1)
	suite.Require().True(found)

	dk.AllocateTokensToValidator(suite.Ctx, validator1, rewardCoins)

	liquidMacc := suite.App.GetAccountKeeper().GetModuleAccount(suite.Ctx, liquidtypes.ModuleAccountName)
	delegation, found := sk.GetDelegation(suite.Ctx, liquidMacc.GetAddress(), valAddr1)
	suite.Require().True(found)

	suite.Ctx = suite.Ctx.WithBlockHeight(100).
		WithBlockTime(suite.Ctx.BlockTime().Add(1 * time.Hour))

	// Get amount of rewards
	endingPeriod := dk.IncrementValidatorPeriod(suite.Ctx, validator1)
	delegationRewards := dk.CalculateDelegationRewards(suite.Ctx, validator1, delegation, endingPeriod)

	// Accumulate rewards - claim rewards
	rewardPeriod := types.NewMultiRewardPeriod(
		true,
		"bkava",         // reward period is set for "bkava" to apply to all vaults
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(), // no incentives, so only the staking rewards are distributed
	)
	err = ik.AccumulateEarnRewards(suite.Ctx, rewardPeriod)
	suite.Require().NoError(err)

	preClaimBal1 := suite.GetBalance(userAddr1)
	preClaimBal2 := suite.GetBalance(userAddr2)

	// Claim ukava staking rewards
	denomsToClaim := map[string]string{"ukava": "large"}
	selections := types.NewSelectionsFromMap(denomsToClaim)

	msg1 := types.NewMsgClaimEarnReward(userAddr1.String(), selections)
	err = suite.DeliverIncentiveMsg(&msg1)
	suite.Require().NoError(err)

	msg2 := types.NewMsgClaimEarnReward(userAddr2.String(), selections)
	err = suite.DeliverIncentiveMsg(&msg2)
	suite.Require().NoError(err)

	// Check rewards were paid out
	// User 1 gets 1% of rewards
	// User 2 gets 99% of rewards
	stakingRewards1 := delegationRewards.
		AmountOf("ukava").
		QuoInt64(100).
		TruncateInt()
	suite.BalanceEquals(userAddr1, preClaimBal1.Add(sdk.NewCoin("ukava", stakingRewards1)))

	// Total * 99 / 100
	stakingRewards2 := delegationRewards.
		AmountOf("ukava").
		MulInt64(99).
		QuoInt64(100).
		TruncateInt()
	suite.BalanceEquals(userAddr2, preClaimBal2.Add(sdk.NewCoin("ukava", stakingRewards2)))

	suite.Equal(delegationRewards.AmountOf("ukava").TruncateInt(), stakingRewards1.Add(stakingRewards2))

	// Check that claimed coins have been removed from a claim's reward
	suite.EarnRewardEquals(userAddr1, cs())
	suite.EarnRewardEquals(userAddr2, cs())
}

// earnBuilder returns a new earn genesis builder with a genesis time and multipliers set
func (suite *HandlerTestSuite) earnBuilder() testutil.EarnGenesisBuilder {
	return testutil.NewEarnGenesisBuilder().
		WithGenesisTime(suite.genesisTime)
}
