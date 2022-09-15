package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
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

	err := suite.App.FundModuleAccount(
		suite.Ctx,
		distrtypes.ModuleName,
		sdk.NewCoins(
			sdk.NewCoin("ukava", initialBalance),
		),
	)
	suite.NoError(err)

	suite.CreateAccountWithAddress(valAccAddr1, suite.NewBondCoins(initialBalance))
	suite.CreateAccountWithAddress(delegator, suite.NewBondCoins(initialBalance))

	suite.CreateNewUnbondedValidator(valAddr1, initialBalance)
	suite.CreateDelegation(valAddr1, delegator, delegateAmount)
	staking.EndBlocker(suite.Ctx, suite.StakingKeeper)

	// Transfers delegation to module account
	_, err = suite.Keeper.MintDerivative(suite.Ctx, delegator, valAddr1, suite.NewBondCoin(delegateAmount))
	suite.Require().NoError(err)

	validator, found := suite.StakingKeeper.GetValidator(suite.Ctx, valAddr1)
	suite.Require().True(found)

	suite.Ctx = suite.Ctx.WithBlockHeight(2)

	rewardCoins := sdk.NewDecCoins(sdk.NewDecCoin("ukava", sdk.NewInt(500e6)))

	distrKeeper := suite.App.GetDistrKeeper()
	stakingKeeper := suite.App.GetStakingKeeper()
	accKeeper := suite.App.GetAccountKeeper()
	liquidMacc := accKeeper.GetModuleAccount(suite.Ctx, types.ModuleAccountName)

	distrKeeper.AllocateTokensToValidator(suite.Ctx, validator, rewardCoins)

	delegation, found := stakingKeeper.GetDelegation(suite.Ctx, liquidMacc.GetAddress(), valAddr1)
	suite.Require().True(found)

	endingPeriod := distrKeeper.IncrementValidatorPeriod(suite.Ctx, validator)
	delegationRewards := distrKeeper.CalculateDelegationRewards(suite.Ctx, validator, delegation, endingPeriod)
	truncatedRewards, _ := delegationRewards.TruncateDecimal()

	derivativeDenom := suite.Keeper.GetLiquidStakingTokenDenom(valAddr1)
	rewards, err := suite.Keeper.CollectStakingRewardsByDenom(suite.Ctx, derivativeDenom, types.ModuleName)
	suite.Require().NoError(err)

	suite.Require().Equal(truncatedRewards, rewards)
}
