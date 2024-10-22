package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/liquid/types"

	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func (suite *KeeperTestSuite) TestCollectStakingRewards() {
	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	valAccAddr1, delegator := addrs[0], addrs[1]
	valAddr1 := sdk.ValAddress(valAccAddr1)

	initialBalance := i(1e9)
	delegateAmount := i(100e6)

	suite.NoError(suite.App.FundModuleAccount(
		suite.Ctx,
		distrtypes.ModuleName,
		sdk.NewCoins(
			sdk.NewCoin("ukava", initialBalance),
		),
	))

	suite.CreateAccountWithAddress(valAccAddr1, suite.NewBondCoins(initialBalance))
	suite.CreateAccountWithAddress(delegator, suite.NewBondCoins(initialBalance))

	suite.CreateNewUnbondedValidator(valAddr1, initialBalance)
	suite.CreateDelegation(valAddr1, delegator, delegateAmount)
	err := suite.StakingKeeper.BeginBlocker(suite.Ctx)
	suite.Require().NoError(err)

	// Transfers delegation to module account
	_, err = suite.Keeper.MintDerivative(suite.Ctx, delegator, valAddr1, suite.NewBondCoin(delegateAmount))
	suite.Require().NoError(err)

	validator, err := suite.StakingKeeper.GetValidator(suite.Ctx, valAddr1)
	suite.Require().NoError(err)

	suite.Ctx = suite.Ctx.WithBlockHeight(2)

	distrKeeper := suite.App.GetDistrKeeper()
	stakingKeeper := suite.App.GetStakingKeeper()
	accKeeper := suite.App.GetAccountKeeper()
	liquidMacc := accKeeper.GetModuleAccount(suite.Ctx, types.ModuleAccountName)

	// Add rewards
	rewardCoins := sdk.NewDecCoins(sdk.NewDecCoin("ukava", sdkmath.NewInt(500e6)))
	distrKeeper.AllocateTokensToValidator(suite.Ctx, validator, rewardCoins)

	delegation, err := stakingKeeper.GetDelegation(suite.Ctx, liquidMacc.GetAddress(), valAddr1)
	suite.Require().NoError(err)

	// Get amount of rewards
	endingPeriod, err := distrKeeper.IncrementValidatorPeriod(suite.Ctx, validator)
	suite.Require().NoError(err)
	delegationRewards, err := distrKeeper.CalculateDelegationRewards(suite.Ctx, validator, delegation, endingPeriod)
	suite.Require().NoError(err)
	truncatedRewards, _ := delegationRewards.TruncateDecimal()

	suite.Run("collect staking rewards", func() {
		// Collect rewards
		derivativeDenom := suite.Keeper.GetLiquidStakingTokenDenom(valAddr1)
		rewards, err := suite.Keeper.CollectStakingRewardsByDenom(suite.Ctx, derivativeDenom, types.ModuleName)
		suite.Require().NoError(err)
		suite.Require().Equal(truncatedRewards, rewards)

		suite.True(rewards.AmountOf("ukava").IsPositive())

		// Check balances
		suite.AccountBalanceEqual(liquidMacc.GetAddress(), rewards)
	})

	suite.Run("collect staking rewards with non-validator", func() {
		// acc2 not a validator
		derivativeDenom := suite.Keeper.GetLiquidStakingTokenDenom(sdk.ValAddress(addrs[2]))
		_, err := suite.Keeper.CollectStakingRewardsByDenom(suite.Ctx, derivativeDenom, types.ModuleName)
		suite.Require().Error(err)
		suite.Require().Equal("no validator distribution info", err.Error())
	})

	suite.Run("collect staking rewards with invalid denom", func() {
		derivativeDenom := "bkava"
		_, err := suite.Keeper.CollectStakingRewardsByDenom(suite.Ctx, derivativeDenom, types.ModuleName)
		suite.Require().Error(err)
		suite.Require().Equal("cannot parse denom bkava", err.Error())
	})
}
