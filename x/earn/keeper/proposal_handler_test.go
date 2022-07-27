package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	distKeeper := suite.App.GetDistrKeeper()
	ctx := suite.Ctx
	macc := distKeeper.GetDistributionAccount(ctx)
	fundAmount := sdk.NewCoins(sdk.NewInt64Coin("ukava", 100000000))
	depositAmount := sdk.NewCoin("ukava", sdk.NewInt(10000000))
	suite.Require().NoError(suite.App.FundModuleAccount(ctx, macc.GetName(), fundAmount))
	feePool := distKeeper.GetFeePool(ctx)
	feePool.CommunityPool = sdk.NewDecCoinsFromCoins(fundAmount...)
	distKeeper.SetFeePool(ctx, feePool)
	// TODO update to STRATEGY_TYPE_SAVINGS once implemented
	suite.CreateVault("ukava", types.STRATEGY_TYPE_HARD)
	prop := types.NewCommunityPoolDepositProposal("test title",
		"desc", depositAmount)
	err := keeper.HandleCommunityPoolDepositProposal(ctx, suite.Keeper, prop)
	suite.Require().NoError(err)

	balance := suite.BankKeeper.GetAllBalances(ctx, macc.GetAddress())
	suite.Require().Equal(fundAmount.Sub(sdk.NewCoins(depositAmount)), balance)
}

func (suite *proposalTestSuite) TestCommunityWithdrawProposal() {
	distKeeper := suite.App.GetDistrKeeper()
	ctx := suite.Ctx
	macc := distKeeper.GetDistributionAccount(ctx)
	fundAmount := sdk.NewCoins(sdk.NewInt64Coin("ukava", 100000000))
	depositAmount := sdk.NewCoin("ukava", sdk.NewInt(10000000))
	suite.Require().NoError(suite.App.FundModuleAccount(ctx, macc.GetName(), fundAmount))
	feePool := distKeeper.GetFeePool(ctx)
	feePool.CommunityPool = sdk.NewDecCoinsFromCoins(fundAmount...)
	distKeeper.SetFeePool(ctx, feePool)
	// TODO update to STRATEGY_TYPE_SAVINGS once implemented
	suite.CreateVault("ukava", types.STRATEGY_TYPE_HARD)
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
}
