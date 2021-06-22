package keeper_test

import (
	"time"

	"github.com/kava-labs/kava/x/swap/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *keeperTestSuite) setupPoolDeposit() (string, sdk.AccAddress) {
	pool := types.NewAllowedPool("ukava", "usdx")
	suite.Require().NoError(pool.Validate())
	suite.Keeper.SetParams(suite.Ctx, types.NewParams(types.NewAllowedPools(pool), types.DefaultSwapFee))

	amountA := sdk.NewCoin(pool.TokenA, sdk.NewInt(11e6))
	amountB := sdk.NewCoin(pool.TokenB, sdk.NewInt(51e6))
	balance := sdk.NewCoins(amountA, amountB)
	depositor := suite.CreateAccount(balance)

	depositA := sdk.NewCoin(pool.TokenA, sdk.NewInt(10e6))
	depositB := sdk.NewCoin(pool.TokenB, sdk.NewInt(50e6))
	deposit := sdk.NewCoins(depositA, depositB)

	err := suite.Keeper.Deposit(suite.Ctx, depositor.GetAddress(), depositA, depositB)
	suite.Require().NoError(err)
	suite.AccountBalanceEqual(depositor, sdk.NewCoins(amountA.Sub(depositA), amountB.Sub(depositB)))
	suite.ModuleAccountBalanceEqual(sdk.NewCoins(depositA, depositB))
	suite.PoolLiquidityEqual(deposit)
	suite.PoolShareValueEqual(depositor, pool, deposit)

	return pool.Name(), depositor.GetAddress()
}

func (suite *keeperTestSuite) TestWithdraw_Full() {
	poolID, depositorAddr := suite.setupPoolDeposit()

	// Confirm module account holds pool's reserves
	initialPoolRecord, found := suite.Keeper.GetPool(suite.Ctx, poolID)
	suite.Require().True(found)
	depositedCoins := sdk.NewCoins(initialPoolRecord.ReservesA, initialPoolRecord.ReservesB)
	suite.ModuleAccountBalanceEqual(depositedCoins)

	// Fetch initial depositor balances and share record
	depositor := suite.GetAccount(depositorAddr)
	initialDepositorCoins := depositor.GetCoins()
	initialShareRecord, found := suite.Keeper.GetDepositorShares(suite.Ctx, depositor.GetAddress(), poolID)
	suite.Require().True(found)

	// Depositor withdraws all shares, expecting all coins to be withdrawn with a slippage of 1%
	err := suite.Keeper.Withdraw(suite.Ctx, depositor.GetAddress(), poolID,
		initialShareRecord.SharesOwned, sdk.MustNewDecFromStr("0.01"),
		initialPoolRecord.ReservesA, initialPoolRecord.ReservesB)
	suite.Require().NoError(err)

	// Move forward block time one minute
	suite.Ctx = suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(time.Minute))

	// Check that full withdraw deleted the pool
	_, found = suite.Keeper.GetPool(suite.Ctx, poolID)
	suite.Require().False(found)

	// Confirm that depositor received withdrawn coins and module account balance is empty
	suite.AccountBalanceEqual(depositor, initialDepositorCoins.Add(initialPoolRecord.ReservesA, initialPoolRecord.ReservesB))
	suite.ModuleAccountBalanceEqual(nil)

	// Check withdraw event attributes
	suite.EventsContains(suite.Ctx.EventManager().Events(), sdk.NewEvent(
		types.EventTypeSwapWithdraw,
		sdk.NewAttribute(types.AttributeKeyPoolID, types.PoolID(initialPoolRecord.ReservesA.Denom, initialPoolRecord.ReservesB.Denom)),
		sdk.NewAttribute(types.AttributeKeyOwner, depositor.GetAddress().String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, depositedCoins.String()),
		sdk.NewAttribute(types.AttributeKeyShares, initialShareRecord.SharesOwned.String()),
	))
}

func (suite *keeperTestSuite) TestWithdraw_Partial() {

	testCases := []struct {
		name                    string
		percentageExpectedCoinA sdk.Dec
		percentageExpectedCoinB sdk.Dec
		percentageShares        sdk.Dec
		slippage                sdk.Dec
		expectErr               bool
		expectedErr             string
	}{
		{
			name:                    "normal",
			percentageExpectedCoinA: sdk.MustNewDecFromStr("0.99"),
			percentageExpectedCoinB: sdk.MustNewDecFromStr("0.99"),
			percentageShares:        sdk.MustNewDecFromStr("0.99"),
			slippage:                sdk.NewDec(0),
			expectErr:               false,
			expectedErr:             "",
		},
		// TODO: add test cases for each error
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			poolID, depositorAddr := suite.setupPoolDeposit()

			// Fetch initial pool record
			initialPoolRecord, found := suite.Keeper.GetPool(suite.Ctx, poolID)
			suite.Require().True(found)

			// Fetch initial depositor balances and share record
			depositor := suite.GetAccount(depositorAddr)
			initialDepositorCoins := depositor.GetCoins()
			initialShareRecord, found := suite.Keeper.GetDepositorShares(suite.Ctx, depositor.GetAddress(), poolID)
			suite.Require().True(found)

			withdrawShares := initialShareRecord.SharesOwned.ToDec().Mul(tc.percentageShares).RoundInt()
			expectedCoinAmountA := initialPoolRecord.ReservesA.Amount.ToDec().Mul(tc.percentageExpectedCoinA)
			expectedCoinA := sdk.NewCoin(initialPoolRecord.ReservesA.Denom, expectedCoinAmountA.RoundInt())
			expectedCoinAmountB := initialPoolRecord.ReservesB.Amount.ToDec().Mul(tc.percentageExpectedCoinB)
			expectedCoinB := sdk.NewCoin(initialPoolRecord.ReservesB.Denom, expectedCoinAmountB.RoundInt())

			// Depositor withdraws shares
			err := suite.Keeper.Withdraw(suite.Ctx, depositor.GetAddress(), poolID,
				withdrawShares, tc.slippage, expectedCoinA, expectedCoinB)
			if tc.expectErr {
				suite.Require().Error(err)
				suite.Contains(err, tc.expectedErr)

				// TODO: confirm pool/depositor balances are unchanged
			}

			suite.Require().NoError(err)
			suite.AccountBalanceEqual(depositor, initialDepositorCoins.Add(expectedCoinA, expectedCoinB))

			// // TODO: Fetch pool record after withdrawal and check shares/reserves
			// suite.Ctx = suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(time.Minute))
			// finalPoolRecord, found = suite.Keeper.GetPool(suite.Ctx, poolID)
		})
	}
}
