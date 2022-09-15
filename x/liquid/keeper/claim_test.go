package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/liquid/types"
)

func (suite *KeeperTestSuite) TestCollectStakingRewards() {
	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	valAccAddr1, delegator := addrs[0], addrs[1]
	valAddr1 := sdk.ValAddress(valAccAddr1)

	initialBalance := i(1e9)
	vestedBalance := i(500e6)
	delegateAmount := i(100e6)

	suite.CreateAccountWithAddress(valAccAddr1, suite.NewBondCoins(initialBalance))
	suite.CreateVestingAccountWithAddress(delegator, suite.NewBondCoins(initialBalance), suite.NewBondCoins(vestedBalance))

	suite.CreateNewUnbondedValidator(valAddr1, initialBalance)
	suite.CreateDelegation(valAddr1, delegator, delegateAmount)

	_, err := suite.Keeper.MintDerivative(suite.Ctx, delegator, valAddr1, suite.NewBondCoin(delegateAmount))
	suite.Require().NoError(err)

	validator, found := suite.StakingKeeper.GetValidator(suite.Ctx, valAddr1)
	suite.Require().True(found)

	rewardCoins := sdk.NewDecCoins(sdk.NewDecCoin("ukava", sdk.NewInt(100e6)))

	distrKeeper := suite.App.GetDistrKeeper()
	distrKeeper.AllocateTokensToValidator(suite.Ctx, validator, rewardCoins)

	derivativeDenom := suite.Keeper.GetLiquidStakingTokenDenom(valAddr1)
	suite.T().Logf("derivativeDenom: %s", derivativeDenom)

	rewards, err := suite.Keeper.CollectStakingRewardsByDenom(suite.Ctx, derivativeDenom, types.ModuleName)
	suite.Require().NoError(err)

	suite.Require().Equal(rewardCoins, (rewards))
}
