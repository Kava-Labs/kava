package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	communitytypes "github.com/kava-labs/kava/x/community/types"
	"github.com/kava-labs/kava/x/kavadist/keeper"
	"github.com/kava-labs/kava/x/kavadist/types"
)

func (suite *keeperTestSuite) TestHandleCommunityPoolMultiSpendProposal() {
	addr, communityKeeper, ctx := suite.Addrs[0], suite.App.GetCommunityKeeper(), suite.Ctx
	initBalances := suite.BankKeeper.GetAllBalances(ctx, addr)

	// add coins to the module account and fund community pool
	initialFunds := int64(1000000)
	fundAmount := sdk.NewCoins(sdk.NewInt64Coin("ukava", initialFunds))
	suite.Require().NoError(suite.App.FundModuleAccount(ctx, communitytypes.ModuleAccountName, fundAmount))
	// expect funds to start in community pool
	commPoolFunds := communityKeeper.GetModuleAccountBalance(ctx)
	suite.Require().True(fundAmount.IsEqual(commPoolFunds))

	proposalAmount1 := int64(1100)
	proposalAmount2 := int64(1200)
	proposal := types.NewCommunityPoolMultiSpendProposal("test title", "description", []types.MultiSpendRecipient{
		{
			Address: addr.String(),
			Amount:  sdk.NewCoins(sdk.NewInt64Coin("ukava", proposalAmount1)),
		},
		{
			Address: addr.String(),
			Amount:  sdk.NewCoins(sdk.NewInt64Coin("ukava", proposalAmount2)),
		},
	})
	err := keeper.HandleCommunityPoolMultiSpendProposal(ctx, suite.Keeper, proposal)
	suite.Require().Nil(err)

	balances := suite.BankKeeper.GetAllBalances(ctx, addr)

	// expect funds to be transferred to recipient
	expected := initBalances.AmountOf("ukava").Add(sdk.NewInt(proposalAmount1 + proposalAmount2))
	suite.Require().Equal(expected, balances.AmountOf("ukava"))

	// expect funds to be deducted from community pool
	expectedCommPool := commPoolFunds.AmountOf("ukava").SubRaw(proposalAmount1 + proposalAmount2)
	suite.Require().Equal(expectedCommPool, communityKeeper.GetModuleAccountBalance(ctx).AmountOf("ukava"))
}
