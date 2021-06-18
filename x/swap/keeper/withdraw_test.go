package keeper_test

import (
	"fmt"
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

func (suite *keeperTestSuite) TestWithdraw() {
	poolID, depositorAddr := suite.setupPoolDeposit()

	// Fetch pools pre-withdraw balances
	initialPoolRecord, found := suite.Keeper.GetPool(suite.Ctx, poolID)
	suite.Require().True(found)
	initialReservesA := initialPoolRecord.ReservesA
	initialReservesB := initialPoolRecord.ReservesB
	initialTotalShares := initialPoolRecord.TotalShares

	// Fetch updated account and initial share record
	depositor := suite.GetAccount(depositorAddr)
	// initialCoins := depositor.GetCoins()
	initialShareRecord, found := suite.Keeper.GetDepositorShares(suite.Ctx, depositor.GetAddress(), poolID)
	suite.Require().True(found)

	withdrawSharesAmt := initialShareRecord.SharesOwned
	err := suite.Keeper.Withdraw(suite.Ctx, depositor.GetAddress(), poolID, withdrawSharesAmt)
	suite.Require().NoError(err)

	// Move forward block time one minute
	suite.Ctx = suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(time.Minute))

	// Check pool's post-withdraw balances
	finalPoolRecord, found := suite.Keeper.GetPool(suite.Ctx, poolID)
	suite.Require().True(found)

	// TODO: these print statements are just here to satisfy the compiler
	fmt.Println(initialReservesA)
	fmt.Println(initialReservesB)
	fmt.Println(initialTotalShares)
	fmt.Println(finalPoolRecord)

	// TODO: check initial vs. final pool record
	// TODO: check depositor balances
	// suite.AccountBalanceEqual(depositor, balance.Sub(expectedDeposit))
	// suite.ModuleAccountBalanceEqual(reserves.Add(expectedDeposit...))
	// suite.PoolLiquidityEqual(reserves.Add(expectedDeposit...))
	// suite.PoolShareValueEqual(depositor, pool, expectedShareValue)

	suite.EventsContains(suite.Ctx.EventManager().Events(), sdk.NewEvent(
		types.EventTypeSwapWithdraw,
		sdk.NewAttribute(types.AttributeKeyPoolID, types.PoolID(initialPoolRecord.ReservesA.Denom, initialPoolRecord.ReservesB.Denom)),
		sdk.NewAttribute(types.AttributeKeyDepositor, depositor.GetAddress().String()),
		sdk.NewAttribute(types.AttributeKeyShares, withdrawSharesAmt.String()),
	))
}
