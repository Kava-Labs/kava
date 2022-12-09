package keeper_test

import (
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/hard/types"
	"github.com/kava-labs/kava/x/hard/types/mocks"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/mock"
)

func (suite *KeeperTestSuite) TestHooks_DepositAndWithdraw() {
	suite.keeper.ClearHooks()
	hardHooks := &mocks.HARDHooks{}
	suite.keeper.SetHooks(hardHooks)

	tokenA := "ukava"
	tokenB := "bnb"

	suite.keeper.SetParams(suite.ctx, types.NewParams(
		types.MoneyMarkets{
			types.NewMoneyMarket("ukava",
				types.NewBorrowLimit(false, sdk.NewDec(100000000*KAVA_CF), sdk.MustNewDecFromStr("0.8")), // Borrow Limit
				"kava:usd",          // Market ID
				sdk.NewInt(KAVA_CF), // Conversion Factor
				types.NewInterestRateModel(
					sdk.MustNewDecFromStr("0.05"),
					sdk.MustNewDecFromStr("2"),
					sdk.MustNewDecFromStr("0.8"),
					sdk.MustNewDecFromStr("10"),
				),
				sdk.MustNewDecFromStr("0.05"),
				sdk.ZeroDec(), // Keeper Reward Percentage
			),
			types.NewMoneyMarket("bnb",
				types.NewBorrowLimit(false, sdk.NewDec(100000000*BNB_CF), sdk.MustNewDecFromStr("0.8")), // Borrow Limit
				"bnb:usd",          // Market ID
				sdk.NewInt(BNB_CF), // Conversion Factor
				types.NewInterestRateModel(
					sdk.MustNewDecFromStr("0.05"),
					sdk.MustNewDecFromStr("2"),
					sdk.MustNewDecFromStr("0.8"),
					sdk.MustNewDecFromStr("10"),
				),
				sdk.MustNewDecFromStr("0.05"),
				sdk.ZeroDec()), // Keeper Reward Percentage
		},
		sdk.NewDec(10),
	))

	balance := sdk.NewCoins(
		sdk.NewCoin(tokenA, sdk.NewInt(1000e6)),
		sdk.NewCoin(tokenB, sdk.NewInt(1000e6)),
	)

	_, addrs := app.GeneratePrivKeyAddressPairs(2)
	suite.Require().NoError(suite.app.FundAccount(suite.ctx, addrs[0], balance))
	suite.Require().NoError(suite.app.FundAccount(suite.ctx, addrs[1], balance))

	depositor_1 := addrs[0]
	depositor_2 := addrs[1]

	depositA := sdk.NewCoin(tokenA, sdk.NewInt(10e6))
	depositB := sdk.NewCoin(tokenB, sdk.NewInt(50e6))

	suite.Run("deposit 1", func() {
		// first deposit creates deposit - calls AfterDepositModified with initial shares
		hardHooks.On("AfterDepositModified", suite.ctx, types.NewDeposit(depositor_1, cs(depositA), nil)).Once()
		err := suite.keeper.Deposit(suite.ctx, depositor_1, cs(depositA))
		suite.Require().NoError(err)

		// second deposit adds to deposit - calls AfterDepositModified
		// shares given are the initial shares, along with a slice that includes new deposit denoms
		hardHooks.On("AfterDepositModified", suite.ctx,
			types.NewDeposit(depositor_1, cs(depositA), nil), // old deposit
			[]string{depositB.Denom},                         // new deposit denoms
		).Once()
		err = suite.keeper.Deposit(suite.ctx, depositor_1, cs(depositB))
		suite.Require().NoError(err)

		// get the shares from the store from the last deposit
		deposit, found := suite.keeper.GetDeposit(suite.ctx, depositor_1)
		suite.Require().True(found)

		// third deposit adds to deposit - calls AfterDepositModified
		// shares given are the shares added in previous deposit, not the shares added to the deposit now
		hardHooks.On("AfterDepositModified", suite.ctx,
			deposit,    // previous deposit
			[]string{}, // no new denoms
		).Once()
		err = suite.keeper.Deposit(suite.ctx, depositor_1, cs(depositB))
		suite.Require().NoError(err)
	})

	suite.Run("deposit 2", func() {
		// first deposit creates deposit - calls AfterDepositModified with initial shares
		hardHooks.On("AfterDepositModified", suite.ctx, types.NewDeposit(depositor_2, cs(depositA), nil)).Once()
		err := suite.keeper.Deposit(suite.ctx, depositor_2, cs(depositA))
		suite.Require().NoError(err)

		// second deposit adds to deposit - calls AfterDepositModified
		// shares given are the initial shares, along with a slice that includes new deposit denoms
		hardHooks.On("AfterDepositModified", suite.ctx,
			types.NewDeposit(depositor_2, cs(depositA), nil), // old deposit
			[]string{depositB.Denom},                         // new deposit denoms
		).Once()
		err = suite.keeper.Deposit(suite.ctx, depositor_2, cs(depositB))
		suite.Require().NoError(err)

		// get the shares from the store from the last deposit
		deposit, found := suite.keeper.GetDeposit(suite.ctx, depositor_2)
		suite.Require().True(found)

		// third deposit adds to deposit - calls AfterDepositModified
		// shares given are the shares added in previous deposit, not the shares added to the deposit now
		hardHooks.On("AfterDepositModified", suite.ctx,
			deposit,    // previous deposit
			[]string{}, // no new denoms
		).Once()
		err = suite.keeper.Deposit(suite.ctx, depositor_2, cs(depositB))
		suite.Require().NoError(err)
	})

	suite.Run("borrow", func() {

	})

	// test hooks with a full withdraw of all shares
	shareRecord, found = suite.keeper.GetDepositorShares(suite.ctx, depositor_1, types.PoolIDFromCoins(deposit))
	suite.Require().True(found)
	// all shares given to AfterDepositModified
	hardHooks.On("AfterDepositModified", suite.ctx, types.PoolIDFromCoins(deposit), depositor_1, shareRecord.SharesOwned).Once()
	err = suite.keeper.Withdraw(suite.ctx, depositor_1, shareRecord.SharesOwned, sdk.NewCoin("ukava", sdk.NewInt(1)), sdk.NewCoin("usdx", sdk.NewInt(1)))
	suite.Require().NoError(err)

	// test hooks on partial withdraw
	shareRecord, found = suite.keeper.GetDepositorShares(suite.ctx, depositor_2, types.PoolIDFromCoins(deposit))
	suite.Require().True(found)
	partialShares := shareRecord.SharesOwned.Quo(sdk.NewInt(3))
	// all shares given to before deposit modified even with partial withdraw
	hardHooks.On("AfterDepositModified", suite.ctx, types.PoolIDFromCoins(deposit), depositor_2, shareRecord.SharesOwned).Once()
	err = suite.keeper.Withdraw(suite.ctx, depositor_2, partialShares, sdk.NewCoin("ukava", sdk.NewInt(1)), sdk.NewCoin("usdx", sdk.NewInt(1)))
	suite.Require().NoError(err)

	// test hooks on second partial withdraw
	shareRecord, found = suite.keeper.GetDepositorShares(suite.ctx, depositor_2, types.PoolIDFromCoins(deposit))
	suite.Require().True(found)
	partialShares = shareRecord.SharesOwned.Quo(sdk.NewInt(2))
	// all shares given to before deposit modified even with partial withdraw
	hardHooks.On("AfterDepositModified", suite.ctx, types.PoolIDFromCoins(deposit), depositor_2, shareRecord.SharesOwned).Once()
	err = suite.keeper.Withdraw(suite.ctx, depositor_2, partialShares, sdk.NewCoin("ukava", sdk.NewInt(1)), sdk.NewCoin("usdx", sdk.NewInt(1)))
	suite.Require().NoError(err)

	// test hooks withdraw all shares with second depositor
	shareRecord, found = suite.keeper.GetDepositorShares(suite.ctx, depositor_2, types.PoolIDFromCoins(deposit))
	suite.Require().True(found)
	// all shares given to before deposit modified even with partial withdraw
	hardHooks.On("AfterDepositModified", suite.ctx, types.PoolIDFromCoins(deposit), depositor_2, shareRecord.SharesOwned).Once()
	err = suite.keeper.Withdraw(suite.ctx, depositor_2, shareRecord.SharesOwned, sdk.NewCoin("ukava", sdk.NewInt(1)), sdk.NewCoin("usdx", sdk.NewInt(1)))
	suite.Require().NoError(err)

	hardHooks.AssertExpectations(suite.T())
}

func (suite *KeeperTestSuite) TestHooks_NoPanicsOnNilHooks() {
	suite.keeper.ClearHooks()

	pool := types.NewAllowedPool("ukava", "usdx")
	suite.Require().NoError(pool.Validate())
	suite.keeper.SetParams(suite.ctx, types.NewParams(types.NewAllowedPools(pool), types.DefaultSwapFee))

	balance := sdk.NewCoins(
		sdk.NewCoin(tokenA, sdk.NewInt(1000e6)),
		sdk.NewCoin(tokenB, sdk.NewInt(1000e6)),
	)
	depositor := suite.CreateAccount(balance)

	depositA := sdk.NewCoin(tokenA, sdk.NewInt(10e6))
	depositB := sdk.NewCoin(tokenB, sdk.NewInt(50e6))
	deposit := sdk.NewCoins(depositA, depositB)

	// deposit create pool should not panic when hooks are not set
	err := suite.keeper.Deposit(suite.ctx, depositor, depositA, depositB, sdk.MustNewDecFromStr("0.0015"))
	suite.Require().NoError(err)

	// existing deposit should not panic with hooks are not set
	err = suite.keeper.Deposit(suite.ctx, depositor, sdk.NewCoin("ukava", sdk.NewInt(5e6)), sdk.NewCoin("usdx", sdk.NewInt(25e6)), sdk.MustNewDecFromStr("0.0015"))
	suite.Require().NoError(err)

	// withdraw of shares should not panic when hooks are not set
	shareRecord, found := suite.keeper.GetDepositorShares(suite.ctx, depositor, types.PoolIDFromCoins(deposit))
	suite.Require().True(found)
	err = suite.keeper.Withdraw(suite.ctx, depositor, shareRecord.SharesOwned, sdk.NewCoin("ukava", sdk.NewInt(1)), sdk.NewCoin("usdx", sdk.NewInt(1)))
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestHooks_HookOrdering() {
	suite.keeper.ClearHooks()
	hardHooks := &mocks.HARDHooks{}
	suite.keeper.SetHooks(hardHooks)

	pool := types.NewAllowedPool("ukava", "usdx")
	suite.Require().NoError(pool.Validate())
	suite.keeper.SetParams(suite.ctx, types.NewParams(types.NewAllowedPools(pool), types.DefaultSwapFee))

	balance := sdk.NewCoins(
		sdk.NewCoin(tokenA, sdk.NewInt(1000e6)),
		sdk.NewCoin(tokenB, sdk.NewInt(1000e6)),
	)
	depositor := suite.CreateAccount(balance)

	depositA := sdk.NewCoin(tokenA, sdk.NewInt(10e6))
	depositB := sdk.NewCoin(tokenB, sdk.NewInt(50e6))
	deposit := sdk.NewCoins(depositA, depositB)

	poolID := types.PoolIDFromCoins(deposit)
	expectedShares := sdk.NewInt(22360679)

	hardHooks.On("AfterDepositModified", suite.ctx, poolID, depositor, expectedShares).Run(func(args mock.Arguments) {
		_, found := suite.keeper.GetDepositorShares(suite.ctx, depositor, poolID)
		suite.Require().True(found, "expected after hook to be called after shares are updated")
	})
	err := suite.keeper.Deposit(suite.ctx, depositor, depositA, depositB, sdk.MustNewDecFromStr("0.0015"))
	suite.Require().NoError(err)

	hardHooks.On("AfterDepositModified", suite.ctx, poolID, depositor, expectedShares).Run(func(args mock.Arguments) {
		shareRecord, found := suite.keeper.GetDepositorShares(suite.ctx, depositor, poolID)
		suite.Require().True(found, "expected share record to exist")
		suite.Equal(expectedShares, shareRecord.SharesOwned, "expected hook to be called before shares are updated")
	})
	err = suite.keeper.Deposit(suite.ctx, depositor, depositA, depositB, sdk.MustNewDecFromStr("0.0015"))
	suite.Require().NoError(err)

	existingShareRecord, found := suite.keeper.GetDepositorShares(suite.ctx, depositor, types.PoolIDFromCoins(deposit))
	suite.Require().True(found)
	hardHooks.On("AfterDepositModified", suite.ctx, poolID, depositor, existingShareRecord.SharesOwned).Run(func(args mock.Arguments) {
		shareRecord, found := suite.keeper.GetDepositorShares(suite.ctx, depositor, poolID)
		suite.Require().True(found, "expected share record to exist")
		suite.Equal(existingShareRecord.SharesOwned, shareRecord.SharesOwned, "expected hook to be called before shares are updated")
	})
	err = suite.keeper.Withdraw(suite.ctx, depositor, existingShareRecord.SharesOwned.Quo(sdk.NewInt(2)), sdk.NewCoin("ukava", sdk.NewInt(1)), sdk.NewCoin("usdx", sdk.NewInt(1)))
	suite.Require().NoError(err)
}
