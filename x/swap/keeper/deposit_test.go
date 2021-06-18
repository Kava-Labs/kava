package keeper_test

import (
	"errors"
	"fmt"

	"github.com/kava-labs/kava/x/swap/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

func (suite *keeperTestSuite) TestDeposit_CreatePool_PoolNotAllowed() {
	depositor := suite.CreateAccount(sdk.Coins{})
	amountA := sdk.NewCoin("ukava", sdk.NewInt(10e6))
	amountB := sdk.NewCoin("usdx", sdk.NewInt(50e6))

	err := suite.Keeper.Deposit(suite.Ctx, depositor.GetAddress(), amountA, amountB)
	suite.Require().EqualError(err, "not allowed: can not create pool 'ukava/usdx'")
}

func (suite *keeperTestSuite) TestDeposit_InsufficientFunds() {
	testCases := []struct {
		name     string
		balanceA sdk.Coin
		balanceB sdk.Coin
		depositA sdk.Coin
		depositB sdk.Coin
	}{
		{
			name:     "no balance",
			balanceA: sdk.Coin{},
			balanceB: sdk.Coin{},
			depositA: sdk.NewCoin("ukava", sdk.NewInt(1)),
			depositB: sdk.NewCoin("usdx", sdk.NewInt(1)),
		},
		{
			name:     "low balance",
			balanceA: sdk.NewCoin("ukava", sdk.NewInt(1000000)),
			balanceB: sdk.NewCoin("usdx", sdk.NewInt(1000000)),
			depositA: sdk.NewCoin("ukava", sdk.NewInt(1000001)),
			depositB: sdk.NewCoin("usdx", sdk.NewInt(10000001)),
		},
		{
			name:     "large balance difference",
			balanceA: sdk.NewCoin("ukava", sdk.NewInt(100e6)),
			balanceB: sdk.NewCoin("usdx", sdk.NewInt(500e6)),
			depositA: sdk.NewCoin("ukava", sdk.NewInt(1000e6)),
			depositB: sdk.NewCoin("usdx", sdk.NewInt(5000e6)),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			pool := types.NewAllowedPool(tc.depositA.Denom, tc.depositB.Denom)
			suite.Require().NoError(pool.Validate())
			suite.Keeper.SetParams(suite.Ctx, types.NewParams(types.NewAllowedPools(pool), types.DefaultSwapFee))

			balance := sdk.Coins{tc.balanceA, tc.balanceB}
			balance.Sort()
			depositor := suite.CreateAccount(balance)

			err := suite.Keeper.Deposit(suite.Ctx, depositor.GetAddress(), tc.depositA, tc.depositB)
			// TODO: wrap in module specific error?
			suite.Require().True(errors.Is(err, sdkerrors.ErrInsufficientFunds), fmt.Sprintf("got err %s", err))

			// test deposit to existing pool insuffient funds
			suite.CreatePool(sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(10e6)), sdk.NewCoin("usdx", sdk.NewInt(50e6))))
			err = suite.Keeper.Deposit(suite.Ctx, depositor.GetAddress(), tc.depositA, tc.depositB)
			suite.Require().True(errors.Is(err, sdkerrors.ErrInsufficientFunds))
		})
	}
}

func (suite *keeperTestSuite) TestDeposit_InsufficientFunds_Vesting() {
	testCases := []struct {
		name     string
		balanceA sdk.Coin
		balanceB sdk.Coin
		vestingA sdk.Coin
		vestingB sdk.Coin
		depositA sdk.Coin
		depositB sdk.Coin
	}{
		{
			name:     "no balance, vesting only",
			balanceA: sdk.Coin{},
			balanceB: sdk.Coin{},
			vestingA: sdk.NewCoin("ukava", sdk.NewInt(1)),
			vestingB: sdk.NewCoin("ukava", sdk.NewInt(1)),
			depositA: sdk.NewCoin("ukava", sdk.NewInt(1)),
			depositB: sdk.NewCoin("usdx", sdk.NewInt(1)),
		},
		{
			name:     "vesting matches balance exactly",
			balanceA: sdk.NewCoin("ukava", sdk.NewInt(1000000)),
			balanceB: sdk.NewCoin("usdx", sdk.NewInt(1000000)),
			vestingA: sdk.NewCoin("ukava", sdk.NewInt(1)),
			vestingB: sdk.NewCoin("usdx", sdk.NewInt(1)),
			depositA: sdk.NewCoin("ukava", sdk.NewInt(1000001)),
			depositB: sdk.NewCoin("usdx", sdk.NewInt(10000001)),
		},
		{
			name:     "large balance difference, vesting covers difference",
			balanceA: sdk.NewCoin("ukava", sdk.NewInt(100e6)),
			balanceB: sdk.NewCoin("usdx", sdk.NewInt(500e6)),
			vestingA: sdk.NewCoin("ukava", sdk.NewInt(1000e6)),
			vestingB: sdk.NewCoin("usdx", sdk.NewInt(5000e6)),
			depositA: sdk.NewCoin("ukava", sdk.NewInt(1000e6)),
			depositB: sdk.NewCoin("usdx", sdk.NewInt(5000e6)),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			pool := types.NewAllowedPool(tc.depositA.Denom, tc.depositB.Denom)
			suite.Require().NoError(pool.Validate())
			suite.Keeper.SetParams(suite.Ctx, types.NewParams(types.NewAllowedPools(pool), types.DefaultSwapFee))

			balance := sdk.Coins{tc.balanceA, tc.balanceB}
			balance.Sort()
			vesting := sdk.Coins{tc.vestingA, tc.vestingB}
			vesting.Sort()
			depositor := suite.CreateVestingAccount(balance, vesting)

			// test create pool insuffient funds
			err := suite.Keeper.Deposit(suite.Ctx, depositor.GetAddress(), tc.depositA, tc.depositB)
			// TODO: wrap in module specific error?
			suite.Require().True(errors.Is(err, sdkerrors.ErrInsufficientFunds))

			// test deposit to existing pool insuffient funds
			suite.CreatePool(sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(10e6)), sdk.NewCoin("usdx", sdk.NewInt(50e6))))
			err = suite.Keeper.Deposit(suite.Ctx, depositor.GetAddress(), tc.depositA, tc.depositB)
			suite.Require().True(errors.Is(err, sdkerrors.ErrInsufficientFunds))
		})
	}
}

func (suite *keeperTestSuite) TestDeposit_CreatePool() {
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

	suite.EventsContains(suite.Ctx.EventManager().Events(), sdk.NewEvent(
		types.EventTypeSwapDeposit,
		sdk.NewAttribute(types.AttributeKeyPoolID, pool.Name()),
		sdk.NewAttribute(types.AttributeKeyDepositor, depositor.GetAddress().String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, deposit.String()),
		sdk.NewAttribute(types.AttributeKeyShares, "22360679"),
	))
}

func (suite *keeperTestSuite) TestDeposit_PoolExists() {
	pool := types.NewAllowedPool("ukava", "usdx")
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
		sdk.NewCoin("usdx", sdk.NewInt(50e6)),
	)
	err := suite.CreatePool(reserves)
	suite.Require().NoError(err)

	balance := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(5e6)),
		sdk.NewCoin("usdx", sdk.NewInt(5e6)),
	)
	depositor := suite.CreateAccount(balance)

	depositA := sdk.NewCoin("usdx", depositor.GetCoins().AmountOf("usdx"))
	depositB := sdk.NewCoin("ukava", depositor.GetCoins().AmountOf("ukava"))

	ctx := suite.App.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	err = suite.Keeper.Deposit(ctx, depositor.GetAddress(), depositA, depositB)
	suite.Require().NoError(err)

	expectedDeposit := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(1e6)),
		sdk.NewCoin("usdx", sdk.NewInt(5e6)),
	)

	expectedShareValue := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(999999)),
		sdk.NewCoin("usdx", sdk.NewInt(4999998)),
	)

	suite.AccountBalanceEqual(depositor, balance.Sub(expectedDeposit))
	suite.ModuleAccountBalanceEqual(reserves.Add(expectedDeposit...))
	suite.PoolLiquidityEqual(reserves.Add(expectedDeposit...))
	suite.PoolShareValueEqual(depositor, pool, expectedShareValue)

	suite.EventsContains(ctx.EventManager().Events(), sdk.NewEvent(
		types.EventTypeSwapDeposit,
		sdk.NewAttribute(types.AttributeKeyPoolID, types.PoolID(pool.TokenA, pool.TokenB)),
		sdk.NewAttribute(types.AttributeKeyDepositor, depositor.GetAddress().String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, expectedDeposit.String()),
		sdk.NewAttribute(types.AttributeKeyShares, "2236067"),
	))
}
