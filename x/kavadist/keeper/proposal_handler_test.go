package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/kavadist/keeper"
	"github.com/kava-labs/kava/x/kavadist/types"
)

func (suite *keeperTestSuite) TestHandleCommunityPoolMultiSpendProposal() {
	addr1, addr2, distrKeeper, ctx := suite.Addrs[0], suite.Addrs[1], suite.App.GetDistrKeeper(), suite.Ctx

	upgradeTime := ctx.BlockTime().Add(24 * time.Hour)

	// add coins to the module account and fund fee pool
	macc := distrKeeper.GetDistributionAccount(ctx)
	fundAmount := sdk.NewCoins(sdk.NewInt64Coin("ukava", 1000000))
	suite.Require().NoError(suite.App.FundModuleAccount(ctx, macc.GetName(), fundAmount))
	feePool := distrKeeper.GetFeePool(ctx)
	feePool.CommunityPool = sdk.NewDecCoinsFromCoins(fundAmount...)
	distrKeeper.SetFeePool(ctx, feePool)

	proposalAmount1 := int64(1100)
	proposalAmount2 := int64(1200)
	proposal := types.NewCommunityPoolMultiSpendProposal("test title", "description", []types.MultiSpendRecipient{
		{
			Address: addr1.String(),
			Amount:  sdk.NewCoins(sdk.NewInt64Coin("ukava", proposalAmount1)),
		},
		{
			Address: addr1.String(),
			Amount:  sdk.NewCoins(sdk.NewInt64Coin("ukava", proposalAmount2)),
		},
		{
			Address: addr2.String(),
			Amount:  sdk.NewCoins(sdk.NewInt64Coin("ukava", proposalAmount1)),
		},
		{
			Address: addr2.String(),
			Amount:  sdk.NewCoins(sdk.NewInt64Coin("ukava", proposalAmount2)),
		},
	})

	suite.Suite.Run("it panics before upgrade time", func() {
		handler := func() { keeper.HandleCommunityPoolMultiSpendProposal(ctx, suite.Keeper, proposal, upgradeTime) }
		suite.Require().PanicsWithValue(fmt.Sprintf("cannot submit multi-spend proposal before %s", upgradeTime.UTC().String()), handler)
	})

	suite.Suite.Run("it sends funds to all recipients at or after upgrade time", func() {
		ctx = ctx.WithBlockTime(upgradeTime)
		initBalances1 := suite.BankKeeper.GetAllBalances(ctx, addr1)
		initBalances2 := suite.BankKeeper.GetAllBalances(ctx, addr2)

		err := keeper.HandleCommunityPoolMultiSpendProposal(ctx, suite.Keeper, proposal, upgradeTime)
		suite.Require().Nil(err)

		balances := suite.BankKeeper.GetAllBalances(ctx, addr1)
		expected := initBalances1.AmountOf("ukava").Add(sdk.NewInt(proposalAmount1 + proposalAmount2))
		suite.Require().Equal(expected, balances.AmountOf("ukava"))

		balances = suite.BankKeeper.GetAllBalances(ctx, addr2)
		expected = initBalances2.AmountOf("ukava").Add(sdk.NewInt(proposalAmount1 + proposalAmount2))
		suite.Require().Equal(expected, balances.AmountOf("ukava"))

		ctx = ctx.WithBlockTime(upgradeTime.Add(24 * time.Hour))
		err = keeper.HandleCommunityPoolMultiSpendProposal(ctx, suite.Keeper, proposal, upgradeTime)
		suite.Require().Nil(err)

		balances = suite.BankKeeper.GetAllBalances(ctx, addr1)
		expected = initBalances1.AmountOf("ukava").Add(sdk.NewInt(proposalAmount1 + proposalAmount2).Mul(sdk.NewInt(2)))
		suite.Require().Equal(expected, balances.AmountOf("ukava"))

		balances = suite.BankKeeper.GetAllBalances(ctx, addr2)
		expected = initBalances2.AmountOf("ukava").Add(sdk.NewInt(proposalAmount1 + proposalAmount2).Mul(sdk.NewInt(2)))
		suite.Require().Equal(expected, balances.AmountOf("ukava"))
	})

	suite.Suite.Run("it errors if pool doesn't have enough funds", func() {
		feePool := distrKeeper.GetFeePool(ctx)
		feePool.CommunityPool = sdk.DecCoins{}
		distrKeeper.SetFeePool(ctx, feePool)

		ctx = ctx.WithBlockTime(upgradeTime)
		err := keeper.HandleCommunityPoolMultiSpendProposal(ctx, suite.Keeper, proposal, upgradeTime)

		suite.Require().Error(err)
	})
}
