package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	communitytypes "github.com/kava-labs/kava/x/community/types"
	"github.com/kava-labs/kava/x/earn/keeper"
	"github.com/kava-labs/kava/x/earn/testutil"
	"github.com/kava-labs/kava/x/earn/types"
	"github.com/stretchr/testify/suite"
)

type proposalTestSuite struct {
	testutil.Suite
}

func (suite *proposalTestSuite) SetupTest() {
	suite.Suite.SetupTest()
	suite.Keeper.SetParams(suite.Ctx, types.DefaultParams())
}

func TestProposalTestSuite(t *testing.T) {
	suite.Run(t, new(proposalTestSuite))
}

func (suite *proposalTestSuite) TestCommunityDepositProposal() {
	ctx := suite.Ctx
	macc := suite.App.GetAccountKeeper().GetModuleAccount(ctx, communitytypes.ModuleAccountName)
	fundAmount := sdk.NewCoins(sdk.NewInt64Coin("ukava", 100000000))
	depositAmount := sdk.NewCoin("ukava", sdk.NewInt(10000000))
	suite.Require().NoError(suite.App.FundModuleAccount(ctx, macc.GetName(), fundAmount))

	suite.CreateVault("ukava", types.StrategyTypes{types.STRATEGY_TYPE_SAVINGS}, false, nil)
	prop := types.NewCommunityPoolDepositProposal("test title",
		"desc", depositAmount)
	err := keeper.HandleCommunityPoolDepositProposal(ctx, suite.Keeper, prop)
	suite.Require().NoError(err)

	balance := suite.BankKeeper.GetAllBalances(ctx, macc.GetAddress())
	suite.Require().Equal(fundAmount.Sub(sdk.NewCoins(depositAmount)), balance)

	communityPoolBalance := suite.App.GetCommunityKeeper().GetModuleAccountBalance(ctx)
	suite.Require().Equal(fundAmount.Sub(sdk.NewCoins(depositAmount)), communityPoolBalance)
}

func (suite *proposalTestSuite) TestCommunityWithdrawProposal() {
	ctx := suite.Ctx
	macc := suite.App.GetAccountKeeper().GetModuleAccount(ctx, communitytypes.ModuleAccountName)
	fundAmount := sdk.NewCoins(sdk.NewInt64Coin("ukava", 100000000))
	depositAmount := sdk.NewCoin("ukava", sdk.NewInt(10000000))
	suite.Require().NoError(suite.App.FundModuleAccount(ctx, macc.GetName(), fundAmount))

	// TODO update to STRATEGY_TYPE_SAVINGS once implemented
	suite.CreateVault("ukava", types.StrategyTypes{types.STRATEGY_TYPE_SAVINGS}, false, nil)
	deposit := types.NewCommunityPoolDepositProposal("test title",
		"desc", depositAmount)
	err := keeper.HandleCommunityPoolDepositProposal(ctx, suite.Keeper, deposit)
	suite.Require().NoError(err)

	balance := suite.BankKeeper.GetAllBalances(ctx, macc.GetAddress())
	suite.Require().Equal(fundAmount.Sub(sdk.NewCoins(depositAmount)), balance)

	withdraw := types.NewCommunityPoolWithdrawProposal("test title",
		"desc", depositAmount)
	err = keeper.HandleCommunityPoolWithdrawProposal(ctx, suite.Keeper, withdraw)
	suite.Require().NoError(err)
	balance = suite.BankKeeper.GetAllBalances(ctx, macc.GetAddress())
	suite.Require().Equal(fundAmount, balance)

	communityPoolBalance := suite.App.GetCommunityKeeper().GetModuleAccountBalance(ctx)
	suite.Require().Equal(fundAmount, communityPoolBalance)
}
