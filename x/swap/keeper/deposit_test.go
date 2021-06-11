package keeper_test

import (
	"errors"

	"github.com/kava-labs/kava/x/swap/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (suite *keeperTestSuite) TestDeposit_CreatePool_PoolExists() {
	depositor := suite.GetAccount(sdk.Coins{})

	amountA := sdk.NewCoin("ukava", sdk.NewInt(10e6))
	amountB := sdk.NewCoin("usdx", sdk.NewInt(50e6))
	pool := types.NewPool(amountA, amountB)

	suite.Keeper.SetPool(suite.Ctx, pool)

	err := suite.Keeper.Deposit(suite.Ctx, depositor.GetAddress(), amountA, amountB)
	suite.Require().EqualError(err, "not implemented: can not deposit into existing pool 'ukava/usdx'")
}

func (suite *keeperTestSuite) TestDeposit_CreatePool_PoolNotAllowed() {
	depositor := suite.GetAccount(sdk.Coins{})
	amountA := sdk.NewCoin("ukava", sdk.NewInt(10e6))
	amountB := sdk.NewCoin("usdx", sdk.NewInt(50e6))

	err := suite.Keeper.Deposit(suite.Ctx, depositor.GetAddress(), amountA, amountB)
	suite.Require().EqualError(err, "not allowed: can not create pool 'ukava/usdx'")
}

func (suite *keeperTestSuite) TestDeposit_CreatePool_InsufficientFunds() {
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
		pool := types.NewAllowedPool(tc.depositA.Denom, tc.depositB.Denom)
		suite.Require().NoError(pool.Validate())
		suite.Keeper.SetParams(suite.Ctx, types.NewParams(types.NewAllowedPools(pool), types.DefaultSwapFee))

		balance := sdk.Coins{tc.balanceA, tc.balanceB}
		balance.Sort()
		depositor := suite.GetAccount(balance)

		err := suite.Keeper.Deposit(suite.Ctx, depositor.GetAddress(), tc.depositA, tc.depositB)
		// TODO: wrap in module specific error?
		suite.Require().True(errors.Is(err, sdkerrors.ErrInsufficientFunds))
	}
}

func (suite *keeperTestSuite) TestDeposit_CreatePool_InsufficientFunds_Vesting() {
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
		pool := types.NewAllowedPool(tc.depositA.Denom, tc.depositB.Denom)
		suite.Require().NoError(pool.Validate())
		suite.Keeper.SetParams(suite.Ctx, types.NewParams(types.NewAllowedPools(pool), types.DefaultSwapFee))

		balance := sdk.Coins{tc.balanceA, tc.balanceB}
		balance.Sort()
		vesting := sdk.Coins{tc.vestingA, tc.vestingB}
		vesting.Sort()
		depositor := suite.GetVestingAccount(balance, vesting)

		err := suite.Keeper.Deposit(suite.Ctx, depositor.GetAddress(), tc.depositA, tc.depositB)
		// TODO: wrap in module specific error?
		suite.Require().True(errors.Is(err, sdkerrors.ErrInsufficientFunds))
	}
}

func (suite *keeperTestSuite) TestDeposit_CreatePool() {
	pool := types.NewAllowedPool("ukava", "usdx")
	suite.Require().NoError(pool.Validate())
	suite.Keeper.SetParams(suite.Ctx, types.NewParams(types.NewAllowedPools(pool), types.DefaultSwapFee))

	amountA := sdk.NewCoin(pool.TokenA, sdk.NewInt(11e6))
	amountB := sdk.NewCoin(pool.TokenB, sdk.NewInt(51e6))
	balance := sdk.NewCoins(amountA, amountB)
	depositor := suite.GetAccount(balance)

	depositA := sdk.NewCoin(pool.TokenA, sdk.NewInt(10e6))
	depositB := sdk.NewCoin(pool.TokenB, sdk.NewInt(50e6))
	deposit := sdk.NewCoins(depositA, depositB)

	err := suite.Keeper.Deposit(suite.Ctx, depositor.GetAddress(), depositA, depositB)
	suite.Require().NoError(err)
	suite.AccountBalanceEqual(depositor, sdk.NewCoins(amountA.Sub(depositA), amountB.Sub(depositB)))
	suite.ModuleAccountBalanceEqual(sdk.NewCoins(depositA, depositB))
	suite.PoolLiquidityEqual(pool, deposit)
	suite.PoolShareValueEqual(depositor, pool, deposit)

	suite.EventsContains(suite.Ctx.EventManager().Events(), sdk.NewEvent(
		types.EventTypeSwapDeposit,
		sdk.NewAttribute(types.AttributeKeyPoolName, pool.Name()),
		sdk.NewAttribute(types.AttributeKeyDepositor, depositor.GetAddress().String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, deposit.String()),
	))
}
