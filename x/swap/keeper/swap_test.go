package keeper_test

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/kava-labs/kava/x/swap/types"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

func (suite *keeperTestSuite) TestSwapExactForTokens() {
	suite.Keeper.SetParams(suite.Ctx, types.Params{
		SwapFee: sdk.MustNewDecFromStr("0.0025"),
	})
	owner := suite.CreateAccount(sdk.Coins{})
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(1000e6)),
		sdk.NewCoin("usdx", sdk.NewInt(5000e6)),
	)
	totalShares := sdk.NewInt(30e6)
	poolID := suite.setupPool(reserves, totalShares, owner.GetAddress())

	balance := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
	)
	requester := suite.NewAccountFromAddr(sdk.AccAddress("requester"), balance)
	coinA := sdk.NewCoin("ukava", sdk.NewInt(1e6))
	coinB := sdk.NewCoin("usdx", sdk.NewInt(5e6))

	err := suite.Keeper.SwapExactForTokens(suite.Ctx, requester.GetAddress(), coinA, coinB, sdk.MustNewDecFromStr("0.01"))
	suite.Require().NoError(err)

	expectedOutput := sdk.NewCoin("usdx", sdk.NewInt(4982529))

	suite.AccountBalanceEqual(requester, balance.Sub(sdk.NewCoins(coinA)).Add(expectedOutput))
	suite.ModuleAccountBalanceEqual(reserves.Add(coinA).Sub(sdk.NewCoins(expectedOutput)))
	suite.PoolLiquidityEqual(reserves.Add(coinA).Sub(sdk.NewCoins(expectedOutput)))

	suite.EventsContains(suite.Ctx.EventManager().Events(), sdk.NewEvent(
		types.EventTypeSwapTrade,
		sdk.NewAttribute(types.AttributeKeyPoolID, poolID),
		sdk.NewAttribute(types.AttributeKeyRequester, requester.GetAddress().String()),
		sdk.NewAttribute(types.AttributeKeySwapInput, coinA.String()),
		sdk.NewAttribute(types.AttributeKeySwapOutput, expectedOutput.String()),
		sdk.NewAttribute(types.AttributeKeyFeePaid, "2500ukava"),
		sdk.NewAttribute(types.AttributeKeyExactDirection, "input"),
	))
}

func (suite *keeperTestSuite) TestSwapExactForTokens_OutputGreaterThanZero() {
	owner := suite.CreateAccount(sdk.Coins{})
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
		sdk.NewCoin("usdx", sdk.NewInt(50e6)),
	)
	totalShares := sdk.NewInt(30e6)
	suite.setupPool(reserves, totalShares, owner.GetAddress())

	balance := sdk.NewCoins(
		sdk.NewCoin("usdx", sdk.NewInt(10e6)),
	)
	requester := suite.NewAccountFromAddr(sdk.AccAddress("requester"), balance)
	coinA := sdk.NewCoin("usdx", sdk.NewInt(5))
	coinB := sdk.NewCoin("ukava", sdk.NewInt(1))

	err := suite.Keeper.SwapExactForTokens(suite.Ctx, requester.GetAddress(), coinA, coinB, sdk.MustNewDecFromStr("1"))
	suite.EqualError(err, "insufficient liquidity: increase input amount")
}

func (suite *keeperTestSuite) TestSwapExactForTokens_Slippage() {
	owner := suite.CreateAccount(sdk.Coins{})
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(100e6)),
		sdk.NewCoin("usdx", sdk.NewInt(500e6)),
	)
	totalShares := sdk.NewInt(30e6)
	suite.setupPool(reserves, totalShares, owner.GetAddress())

	testCases := []struct {
		coinA      sdk.Coin
		coinB      sdk.Coin
		slippage   sdk.Dec
		fee        sdk.Dec
		shouldFail bool
	}{
		// positive slippage OK
		{sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.NewCoin("usdx", sdk.NewInt(2e6)), sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0.0025"), false},
		{sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.NewCoin("usdx", sdk.NewInt(4e6)), sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0.0025"), false},
		{sdk.NewCoin("usdx", sdk.NewInt(50e6)), sdk.NewCoin("ukava", sdk.NewInt(5e6)), sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0.0025"), false},
		{sdk.NewCoin("usdx", sdk.NewInt(50e6)), sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0.0025"), false},
		// positive slippage with zero slippage OK
		{sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.NewCoin("usdx", sdk.NewInt(2e6)), sdk.ZeroDec(), sdk.MustNewDecFromStr("0.0025"), false},
		{sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.NewCoin("usdx", sdk.NewInt(4e6)), sdk.ZeroDec(), sdk.MustNewDecFromStr("0.0025"), false},
		{sdk.NewCoin("usdx", sdk.NewInt(50e6)), sdk.NewCoin("ukava", sdk.NewInt(5e6)), sdk.ZeroDec(), sdk.MustNewDecFromStr("0.0025"), false},
		{sdk.NewCoin("usdx", sdk.NewInt(50e6)), sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.ZeroDec(), sdk.MustNewDecFromStr("0.0025"), false},
		// exact zero slippage OK
		{sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.NewCoin("usdx", sdk.NewInt(4950495)), sdk.ZeroDec(), sdk.MustNewDecFromStr("0"), false},
		{sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.NewCoin("usdx", sdk.NewInt(4935790)), sdk.ZeroDec(), sdk.MustNewDecFromStr("0.003"), false},
		{sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.NewCoin("usdx", sdk.NewInt(4705299)), sdk.ZeroDec(), sdk.MustNewDecFromStr("0.05"), false},
		{sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.NewCoin("ukava", sdk.NewInt(990099)), sdk.ZeroDec(), sdk.MustNewDecFromStr("0"), false},
		{sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.NewCoin("ukava", sdk.NewInt(987158)), sdk.ZeroDec(), sdk.MustNewDecFromStr("0.003"), false},
		{sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.NewCoin("ukava", sdk.NewInt(941059)), sdk.ZeroDec(), sdk.MustNewDecFromStr("0.05"), false},
		// slippage failure, zero slippage tolerance
		{sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.NewCoin("usdx", sdk.NewInt(4950496)), sdk.ZeroDec(), sdk.MustNewDecFromStr("0"), true},
		{sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.NewCoin("usdx", sdk.NewInt(4935793)), sdk.ZeroDec(), sdk.MustNewDecFromStr("0.003"), true},
		{sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.NewCoin("usdx", sdk.NewInt(4705300)), sdk.ZeroDec(), sdk.MustNewDecFromStr("0.05"), true},
		{sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.NewCoin("ukava", sdk.NewInt(990100)), sdk.ZeroDec(), sdk.MustNewDecFromStr("0"), true},
		{sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.NewCoin("ukava", sdk.NewInt(987159)), sdk.ZeroDec(), sdk.MustNewDecFromStr("0.003"), true},
		{sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.NewCoin("ukava", sdk.NewInt(941060)), sdk.ZeroDec(), sdk.MustNewDecFromStr("0.05"), true},
		// slippage failure, 1 percent slippage
		{sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.NewCoin("usdx", sdk.NewInt(5000501)), sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0"), true},
		{sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.NewCoin("usdx", sdk.NewInt(4985647)), sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0.003"), true},
		{sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.NewCoin("usdx", sdk.NewInt(4752828)), sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0.05"), true},
		{sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.NewCoin("ukava", sdk.NewInt(1000101)), sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0"), true},
		{sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.NewCoin("ukava", sdk.NewInt(997130)), sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0.003"), true},
		{sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.NewCoin("ukava", sdk.NewInt(950565)), sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0.05"), true},
		// slippage OK, 1 percent slippage
		{sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.NewCoin("usdx", sdk.NewInt(5000500)), sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0"), false},
		{sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.NewCoin("usdx", sdk.NewInt(4985646)), sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0.003"), false},
		{sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.NewCoin("usdx", sdk.NewInt(4752827)), sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0.05"), false},
		{sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.NewCoin("ukava", sdk.NewInt(1000100)), sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0"), false},
		{sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.NewCoin("ukava", sdk.NewInt(997129)), sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0.003"), false},
		{sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.NewCoin("ukava", sdk.NewInt(950564)), sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0.05"), false},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("coinA=%s coinB=%s slippage=%s fee=%s", tc.coinA, tc.coinB, tc.slippage, tc.fee), func() {
			suite.SetupTest()
			suite.Keeper.SetParams(suite.Ctx, types.Params{
				SwapFee: tc.fee,
			})
			owner := suite.CreateAccount(sdk.Coins{})
			reserves := sdk.NewCoins(
				sdk.NewCoin("ukava", sdk.NewInt(100e6)),
				sdk.NewCoin("usdx", sdk.NewInt(500e6)),
			)
			totalShares := sdk.NewInt(30e6)
			suite.setupPool(reserves, totalShares, owner.GetAddress())
			balance := sdk.NewCoins(
				sdk.NewCoin("ukava", sdk.NewInt(100e6)),
				sdk.NewCoin("usdx", sdk.NewInt(100e6)),
			)
			requester := suite.NewAccountFromAddr(sdk.AccAddress("requester"), balance)

			ctx := suite.App.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
			err := suite.Keeper.SwapExactForTokens(ctx, requester.GetAddress(), tc.coinA, tc.coinB, tc.slippage)

			if tc.shouldFail {
				suite.Require().Error(err)
				suite.Contains(err.Error(), "slippage exceeded")
			} else {
				suite.NoError(err)
			}
		})
	}
}

func (suite *keeperTestSuite) TestSwapExactForTokens_InsufficientFunds() {
	testCases := []struct {
		name     string
		balanceA sdk.Coin
		coinA    sdk.Coin
		coinB    sdk.Coin
	}{
		{"no ukava balance", sdk.NewCoin("ukava", sdk.ZeroInt()), sdk.NewCoin("ukava", sdk.NewInt(100)), sdk.NewCoin("usdx", sdk.NewInt(500))},
		{"no usdx balance", sdk.NewCoin("usdx", sdk.ZeroInt()), sdk.NewCoin("usdx", sdk.NewInt(500)), sdk.NewCoin("ukava", sdk.NewInt(100))},
		{"low ukava balance", sdk.NewCoin("ukava", sdk.NewInt(1000000)), sdk.NewCoin("ukava", sdk.NewInt(1000001)), sdk.NewCoin("usdx", sdk.NewInt(5000000))},
		{"low ukava balance", sdk.NewCoin("usdx", sdk.NewInt(5000000)), sdk.NewCoin("usdx", sdk.NewInt(5000001)), sdk.NewCoin("ukava", sdk.NewInt(1000000))},
		{"large ukava balance difference", sdk.NewCoin("ukava", sdk.NewInt(100e6)), sdk.NewCoin("ukava", sdk.NewInt(1000e6)), sdk.NewCoin("usdx", sdk.NewInt(5000e6))},
		{"large usdx balance difference", sdk.NewCoin("usdx", sdk.NewInt(500e6)), sdk.NewCoin("usdx", sdk.NewInt(5000e6)), sdk.NewCoin("ukava", sdk.NewInt(1000e6))},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			owner := suite.CreateAccount(sdk.Coins{})
			reserves := sdk.NewCoins(
				sdk.NewCoin("ukava", sdk.NewInt(100000e6)),
				sdk.NewCoin("usdx", sdk.NewInt(500000e6)),
			)
			totalShares := sdk.NewInt(30000e6)
			suite.setupPool(reserves, totalShares, owner.GetAddress())
			balance := sdk.NewCoins(tc.balanceA)
			requester := suite.NewAccountFromAddr(sdk.AccAddress("requester"), balance)

			ctx := suite.App.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
			err := suite.Keeper.SwapExactForTokens(ctx, requester.GetAddress(), tc.coinA, tc.coinB, sdk.MustNewDecFromStr("0.1"))
			suite.Require().True(errors.Is(err, sdkerrors.ErrInsufficientFunds), fmt.Sprintf("got err %s", err))
		})
	}
}

func (suite *keeperTestSuite) TestSwapExactForTokens_InsufficientFunds_Vesting() {
	testCases := []struct {
		name     string
		balanceA sdk.Coin
		vestingA sdk.Coin
		coinA    sdk.Coin
		coinB    sdk.Coin
	}{
		{"no ukava balance, vesting only", sdk.NewCoin("ukava", sdk.ZeroInt()), sdk.NewCoin("ukava", sdk.NewInt(100)), sdk.NewCoin("ukava", sdk.NewInt(100)), sdk.NewCoin("usdx", sdk.NewInt(500))},
		{"no usdx balance, vesting only", sdk.NewCoin("usdx", sdk.ZeroInt()), sdk.NewCoin("usdx", sdk.NewInt(500)), sdk.NewCoin("usdx", sdk.NewInt(500)), sdk.NewCoin("ukava", sdk.NewInt(100))},
		{"low ukava balance, vesting matches exact", sdk.NewCoin("ukava", sdk.NewInt(1000000)), sdk.NewCoin("ukava", sdk.NewInt(1)), sdk.NewCoin("ukava", sdk.NewInt(1000001)), sdk.NewCoin("usdx", sdk.NewInt(5000000))},
		{"low ukava balance, vesting matches exact", sdk.NewCoin("usdx", sdk.NewInt(5000000)), sdk.NewCoin("usdx", sdk.NewInt(1)), sdk.NewCoin("usdx", sdk.NewInt(5000001)), sdk.NewCoin("ukava", sdk.NewInt(1000000))},
		{"large ukava balance difference, vesting covers difference", sdk.NewCoin("ukava", sdk.NewInt(100e6)), sdk.NewCoin("ukava", sdk.NewInt(1000e6)), sdk.NewCoin("ukava", sdk.NewInt(1000e6)), sdk.NewCoin("usdx", sdk.NewInt(5000e6))},
		{"large usdx balance difference, vesting covers difference", sdk.NewCoin("usdx", sdk.NewInt(500e6)), sdk.NewCoin("usdx", sdk.NewInt(5000e6)), sdk.NewCoin("usdx", sdk.NewInt(5000e6)), sdk.NewCoin("ukava", sdk.NewInt(1000e6))},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			owner := suite.CreateAccount(sdk.Coins{})
			reserves := sdk.NewCoins(
				sdk.NewCoin("ukava", sdk.NewInt(100000e6)),
				sdk.NewCoin("usdx", sdk.NewInt(500000e6)),
			)
			totalShares := sdk.NewInt(30000e6)
			suite.setupPool(reserves, totalShares, owner.GetAddress())
			balance := sdk.NewCoins(tc.balanceA)
			vesting := sdk.NewCoins(tc.vestingA)
			requester := suite.CreateVestingAccount(balance, vesting)

			ctx := suite.App.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
			err := suite.Keeper.SwapExactForTokens(ctx, requester.GetAddress(), tc.coinA, tc.coinB, sdk.MustNewDecFromStr("0.1"))
			suite.Require().True(errors.Is(err, sdkerrors.ErrInsufficientFunds), fmt.Sprintf("got err %s", err))
		})
	}
}

func (suite *keeperTestSuite) TestSwapExactForTokens_PoolNotFound() {
	owner := suite.CreateAccount(sdk.Coins{})
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(1000e6)),
		sdk.NewCoin("usdx", sdk.NewInt(5000e6)),
	)
	totalShares := sdk.NewInt(3000e6)
	poolID := suite.setupPool(reserves, totalShares, owner.GetAddress())
	suite.Keeper.DeletePool(suite.Ctx, poolID)

	balance := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
		sdk.NewCoin("usdx", sdk.NewInt(10e6)),
	)
	requester := suite.NewAccountFromAddr(sdk.AccAddress("requester"), balance)
	coinA := sdk.NewCoin("ukava", sdk.NewInt(1e6))
	coinB := sdk.NewCoin("usdx", sdk.NewInt(5e6))

	err := suite.Keeper.SwapExactForTokens(suite.Ctx, requester.GetAddress(), coinA, coinB, sdk.MustNewDecFromStr("0.01"))
	suite.EqualError(err, "invalid pool: pool ukava/usdx not found")

	err = suite.Keeper.SwapExactForTokens(suite.Ctx, requester.GetAddress(), coinB, coinA, sdk.MustNewDecFromStr("0.01"))
	suite.EqualError(err, "invalid pool: pool ukava/usdx not found")
}

func (suite *keeperTestSuite) TestSwapExactForTokens_PanicOnInvalidPool() {
	owner := suite.CreateAccount(sdk.Coins{})
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(1000e6)),
		sdk.NewCoin("usdx", sdk.NewInt(5000e6)),
	)
	totalShares := sdk.NewInt(3000e6)
	poolID := suite.setupPool(reserves, totalShares, owner.GetAddress())

	poolRecord, found := suite.Keeper.GetPool(suite.Ctx, poolID)
	suite.Require().True(found, "expected pool record to exist")

	poolRecord.TotalShares = sdk.ZeroInt()
	suite.Keeper.SetPool(suite.Ctx, poolRecord)

	balance := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
		sdk.NewCoin("usdx", sdk.NewInt(10e6)),
	)
	requester := suite.NewAccountFromAddr(sdk.AccAddress("requester"), balance)
	coinA := sdk.NewCoin("ukava", sdk.NewInt(1e6))
	coinB := sdk.NewCoin("usdx", sdk.NewInt(5e6))

	suite.PanicsWithValue("invalid pool ukava/usdx: invalid pool: total shares must be greater than zero", func() {
		_ = suite.Keeper.SwapExactForTokens(suite.Ctx, requester.GetAddress(), coinA, coinB, sdk.MustNewDecFromStr("0.01"))
	}, "expected invalid pool record to panic")

	suite.PanicsWithValue("invalid pool ukava/usdx: invalid pool: total shares must be greater than zero", func() {
		_ = suite.Keeper.SwapExactForTokens(suite.Ctx, requester.GetAddress(), coinB, coinA, sdk.MustNewDecFromStr("0.01"))
	}, "expected invalid pool record to panic")
}

func (suite *keeperTestSuite) TestSwapExactForTokens_PanicOnInsuffientModuleAccFunds() {
	owner := suite.CreateAccount(sdk.Coins{})
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(1000e6)),
		sdk.NewCoin("usdx", sdk.NewInt(5000e6)),
	)
	totalShares := sdk.NewInt(3000e6)
	suite.setupPool(reserves, totalShares, owner.GetAddress())

	suite.RemoveCoinsFromModule(sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(1000e6)),
		sdk.NewCoin("usdx", sdk.NewInt(5000e6)),
	))

	balance := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
		sdk.NewCoin("usdx", sdk.NewInt(10e6)),
	)
	requester := suite.NewAccountFromAddr(sdk.AccAddress("requester"), balance)
	coinA := sdk.NewCoin("ukava", sdk.NewInt(1e6))
	coinB := sdk.NewCoin("usdx", sdk.NewInt(5e6))

	suite.Panics(func() {
		_ = suite.Keeper.SwapExactForTokens(suite.Ctx, requester.GetAddress(), coinA, coinB, sdk.MustNewDecFromStr("0.01"))
	}, "expected panic when module account does not have enough funds")

	suite.Panics(func() {
		_ = suite.Keeper.SwapExactForTokens(suite.Ctx, requester.GetAddress(), coinA, coinB, sdk.MustNewDecFromStr("0.01"))
	}, "expected panic when module account does not have enough funds")
}

func (suite *keeperTestSuite) TestSwapForExactTokens() {
	suite.Keeper.SetParams(suite.Ctx, types.Params{
		SwapFee: sdk.MustNewDecFromStr("0.0025"),
	})
	owner := suite.CreateAccount(sdk.Coins{})
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(1000e6)),
		sdk.NewCoin("usdx", sdk.NewInt(5000e6)),
	)
	totalShares := sdk.NewInt(30e6)
	poolID := suite.setupPool(reserves, totalShares, owner.GetAddress())

	balance := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
	)
	requester := suite.NewAccountFromAddr(sdk.AccAddress("requester"), balance)
	coinA := sdk.NewCoin("ukava", sdk.NewInt(1e6))
	coinB := sdk.NewCoin("usdx", sdk.NewInt(5e6))

	err := suite.Keeper.SwapForExactTokens(suite.Ctx, requester.GetAddress(), coinA, coinB, sdk.MustNewDecFromStr("0.01"))
	suite.Require().NoError(err)

	expectedInput := sdk.NewCoin("ukava", sdk.NewInt(1003511))

	suite.AccountBalanceEqual(requester, balance.Sub(sdk.NewCoins(expectedInput)).Add(coinB))
	suite.ModuleAccountBalanceEqual(reserves.Add(expectedInput).Sub(sdk.NewCoins(coinB)))
	suite.PoolLiquidityEqual(reserves.Add(expectedInput).Sub(sdk.NewCoins(coinB)))

	suite.EventsContains(suite.Ctx.EventManager().Events(), sdk.NewEvent(
		types.EventTypeSwapTrade,
		sdk.NewAttribute(types.AttributeKeyPoolID, poolID),
		sdk.NewAttribute(types.AttributeKeyRequester, requester.GetAddress().String()),
		sdk.NewAttribute(types.AttributeKeySwapInput, expectedInput.String()),
		sdk.NewAttribute(types.AttributeKeySwapOutput, coinB.String()),
		sdk.NewAttribute(types.AttributeKeyFeePaid, "2509ukava"),
		sdk.NewAttribute(types.AttributeKeyExactDirection, "output"),
	))
}

func (suite *keeperTestSuite) TestSwapForExactTokens_OutputLessThanPoolReserves() {
	owner := suite.CreateAccount(sdk.Coins{})
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(100e6)),
		sdk.NewCoin("usdx", sdk.NewInt(500e6)),
	)
	totalShares := sdk.NewInt(300e6)
	suite.setupPool(reserves, totalShares, owner.GetAddress())

	balance := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(1000e6)),
	)
	requester := suite.NewAccountFromAddr(sdk.AccAddress("requester"), balance)
	coinA := sdk.NewCoin("ukava", sdk.NewInt(1e6))

	coinB := sdk.NewCoin("usdx", sdk.NewInt(500e6).Add(sdk.OneInt()))
	err := suite.Keeper.SwapForExactTokens(suite.Ctx, requester.GetAddress(), coinA, coinB, sdk.MustNewDecFromStr("0.01"))
	suite.EqualError(err, "insufficient liquidity: output 500000001 >= pool reserves 500000000")

	coinB = sdk.NewCoin("usdx", sdk.NewInt(500e6))
	err = suite.Keeper.SwapForExactTokens(suite.Ctx, requester.GetAddress(), coinA, coinB, sdk.MustNewDecFromStr("0.01"))
	suite.EqualError(err, "insufficient liquidity: output 500000000 >= pool reserves 500000000")
}

func (suite *keeperTestSuite) TestSwapForExactTokens_Slippage() {
	owner := suite.CreateAccount(sdk.Coins{})
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(100e6)),
		sdk.NewCoin("usdx", sdk.NewInt(500e6)),
	)
	totalShares := sdk.NewInt(30e6)
	suite.setupPool(reserves, totalShares, owner.GetAddress())

	testCases := []struct {
		coinA      sdk.Coin
		coinB      sdk.Coin
		slippage   sdk.Dec
		fee        sdk.Dec
		shouldFail bool
	}{
		// positive slippage OK
		{sdk.NewCoin("ukava", sdk.NewInt(5e6)), sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0.0025"), false},
		{sdk.NewCoin("ukava", sdk.NewInt(5e6)), sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0.0025"), false},
		{sdk.NewCoin("usdx", sdk.NewInt(100e6)), sdk.NewCoin("ukava", sdk.NewInt(10e6)), sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0.0025"), false},
		{sdk.NewCoin("usdx", sdk.NewInt(100e6)), sdk.NewCoin("ukava", sdk.NewInt(10e6)), sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0.0025"), false},
		// positive slippage with zero slippage OK
		{sdk.NewCoin("ukava", sdk.NewInt(5e6)), sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.ZeroDec(), sdk.MustNewDecFromStr("0.0025"), false},
		{sdk.NewCoin("ukava", sdk.NewInt(5e6)), sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.ZeroDec(), sdk.MustNewDecFromStr("0.0025"), false},
		{sdk.NewCoin("usdx", sdk.NewInt(100e6)), sdk.NewCoin("ukava", sdk.NewInt(10e6)), sdk.ZeroDec(), sdk.MustNewDecFromStr("0.0025"), false},
		{sdk.NewCoin("usdx", sdk.NewInt(100e6)), sdk.NewCoin("ukava", sdk.NewInt(10e6)), sdk.ZeroDec(), sdk.MustNewDecFromStr("0.0025"), false},
		// exact zero slippage OK
		{sdk.NewCoin("ukava", sdk.NewInt(1010102)), sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.ZeroDec(), sdk.MustNewDecFromStr("0"), false},
		{sdk.NewCoin("ukava", sdk.NewInt(1010102)), sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.ZeroDec(), sdk.MustNewDecFromStr("0.003"), false},
		{sdk.NewCoin("ukava", sdk.NewInt(1010102)), sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.ZeroDec(), sdk.MustNewDecFromStr("0.05"), false},
		{sdk.NewCoin("usdx", sdk.NewInt(5050506)), sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.ZeroDec(), sdk.MustNewDecFromStr("0"), false},
		{sdk.NewCoin("usdx", sdk.NewInt(5050506)), sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.ZeroDec(), sdk.MustNewDecFromStr("0.003"), false},
		{sdk.NewCoin("usdx", sdk.NewInt(5050506)), sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.ZeroDec(), sdk.MustNewDecFromStr("0.05"), false},
		// slippage failure, zero slippage tolerance
		{sdk.NewCoin("ukava", sdk.NewInt(1010101)), sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.ZeroDec(), sdk.MustNewDecFromStr("0"), true},
		{sdk.NewCoin("ukava", sdk.NewInt(1010101)), sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.ZeroDec(), sdk.MustNewDecFromStr("0.003"), true},
		{sdk.NewCoin("ukava", sdk.NewInt(1010101)), sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.ZeroDec(), sdk.MustNewDecFromStr("0.05"), true},
		{sdk.NewCoin("usdx", sdk.NewInt(5050505)), sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.ZeroDec(), sdk.MustNewDecFromStr("0"), true},
		{sdk.NewCoin("usdx", sdk.NewInt(5050505)), sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.ZeroDec(), sdk.MustNewDecFromStr("0.003"), true},
		{sdk.NewCoin("usdx", sdk.NewInt(5050505)), sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.ZeroDec(), sdk.MustNewDecFromStr("0.05"), true},
		// slippage failure, 1 percent slippage
		{sdk.NewCoin("ukava", sdk.NewInt(1000000)), sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0"), true},
		{sdk.NewCoin("ukava", sdk.NewInt(1000000)), sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0.003"), true},
		{sdk.NewCoin("ukava", sdk.NewInt(1000000)), sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0.05"), true},
		{sdk.NewCoin("usdx", sdk.NewInt(5000000)), sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0"), true},
		{sdk.NewCoin("usdx", sdk.NewInt(5000000)), sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0.003"), true},
		{sdk.NewCoin("usdx", sdk.NewInt(5000000)), sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0.05"), true},
		// slippage OK, 1 percent slippage
		{sdk.NewCoin("ukava", sdk.NewInt(1000001)), sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0"), false},
		{sdk.NewCoin("ukava", sdk.NewInt(1000001)), sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0.003"), false},
		{sdk.NewCoin("ukava", sdk.NewInt(1000001)), sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0.05"), false},
		{sdk.NewCoin("usdx", sdk.NewInt(5000001)), sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0"), false},
		{sdk.NewCoin("usdx", sdk.NewInt(5000001)), sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0.003"), false},
		{sdk.NewCoin("usdx", sdk.NewInt(5000001)), sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.MustNewDecFromStr("0.01"), sdk.MustNewDecFromStr("0.05"), false},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("coinA=%s coinB=%s slippage=%s fee=%s", tc.coinA, tc.coinB, tc.slippage, tc.fee), func() {
			suite.SetupTest()
			suite.Keeper.SetParams(suite.Ctx, types.Params{
				SwapFee: tc.fee,
			})
			owner := suite.CreateAccount(sdk.Coins{})
			reserves := sdk.NewCoins(
				sdk.NewCoin("ukava", sdk.NewInt(100e6)),
				sdk.NewCoin("usdx", sdk.NewInt(500e6)),
			)
			totalShares := sdk.NewInt(30e6)
			suite.setupPool(reserves, totalShares, owner.GetAddress())
			balance := sdk.NewCoins(
				sdk.NewCoin("ukava", sdk.NewInt(100e6)),
				sdk.NewCoin("usdx", sdk.NewInt(100e6)),
			)
			requester := suite.NewAccountFromAddr(sdk.AccAddress("requester"), balance)

			ctx := suite.App.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
			err := suite.Keeper.SwapForExactTokens(ctx, requester.GetAddress(), tc.coinA, tc.coinB, tc.slippage)

			if tc.shouldFail {
				suite.Require().Error(err)
				suite.Contains(err.Error(), "slippage exceeded")
			} else {
				suite.NoError(err)
			}
		})
	}
}

func (suite *keeperTestSuite) TestSwapForExactTokens_InsufficientFunds() {
	testCases := []struct {
		name     string
		balanceA sdk.Coin
		coinA    sdk.Coin
		coinB    sdk.Coin
	}{
		{"no ukava balance", sdk.NewCoin("ukava", sdk.ZeroInt()), sdk.NewCoin("ukava", sdk.NewInt(100)), sdk.NewCoin("usdx", sdk.NewInt(500))},
		{"no usdx balance", sdk.NewCoin("usdx", sdk.ZeroInt()), sdk.NewCoin("usdx", sdk.NewInt(500)), sdk.NewCoin("ukava", sdk.NewInt(100))},
		{"low ukava balance", sdk.NewCoin("ukava", sdk.NewInt(1000000)), sdk.NewCoin("ukava", sdk.NewInt(1000000)), sdk.NewCoin("usdx", sdk.NewInt(5000000))},
		{"low ukava balance", sdk.NewCoin("usdx", sdk.NewInt(5000000)), sdk.NewCoin("usdx", sdk.NewInt(5000000)), sdk.NewCoin("ukava", sdk.NewInt(1000000))},
		{"large ukava balance difference", sdk.NewCoin("ukava", sdk.NewInt(100e6)), sdk.NewCoin("ukava", sdk.NewInt(1000e6)), sdk.NewCoin("usdx", sdk.NewInt(5000e6))},
		{"large usdx balance difference", sdk.NewCoin("usdx", sdk.NewInt(500e6)), sdk.NewCoin("usdx", sdk.NewInt(5000e6)), sdk.NewCoin("ukava", sdk.NewInt(1000e6))},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			owner := suite.CreateAccount(sdk.Coins{})
			reserves := sdk.NewCoins(
				sdk.NewCoin("ukava", sdk.NewInt(100000e6)),
				sdk.NewCoin("usdx", sdk.NewInt(500000e6)),
			)
			totalShares := sdk.NewInt(30000e6)
			suite.setupPool(reserves, totalShares, owner.GetAddress())
			balance := sdk.NewCoins(tc.balanceA)
			requester := suite.NewAccountFromAddr(sdk.AccAddress("requester"), balance)

			ctx := suite.App.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
			err := suite.Keeper.SwapForExactTokens(ctx, requester.GetAddress(), tc.coinA, tc.coinB, sdk.MustNewDecFromStr("0.1"))
			suite.Require().True(errors.Is(err, sdkerrors.ErrInsufficientFunds), fmt.Sprintf("got err %s", err))
		})
	}
}

func (suite *keeperTestSuite) TestSwapForExactTokens_InsufficientFunds_Vesting() {
	testCases := []struct {
		name     string
		balanceA sdk.Coin
		vestingA sdk.Coin
		coinA    sdk.Coin
		coinB    sdk.Coin
	}{
		{"no ukava balance, vesting only", sdk.NewCoin("ukava", sdk.ZeroInt()), sdk.NewCoin("ukava", sdk.NewInt(100)), sdk.NewCoin("ukava", sdk.NewInt(1000)), sdk.NewCoin("usdx", sdk.NewInt(500))},
		{"no usdx balance, vesting only", sdk.NewCoin("usdx", sdk.ZeroInt()), sdk.NewCoin("usdx", sdk.NewInt(500)), sdk.NewCoin("usdx", sdk.NewInt(5000)), sdk.NewCoin("ukava", sdk.NewInt(100))},
		{"low ukava balance, vesting matches exact", sdk.NewCoin("ukava", sdk.NewInt(1000000)), sdk.NewCoin("ukava", sdk.NewInt(100000)), sdk.NewCoin("ukava", sdk.NewInt(1000000)), sdk.NewCoin("usdx", sdk.NewInt(5000000))},
		{"low ukava balance, vesting matches exact", sdk.NewCoin("usdx", sdk.NewInt(5000000)), sdk.NewCoin("usdx", sdk.NewInt(500000)), sdk.NewCoin("usdx", sdk.NewInt(5000000)), sdk.NewCoin("ukava", sdk.NewInt(1000000))},
		{"large ukava balance difference, vesting covers difference", sdk.NewCoin("ukava", sdk.NewInt(100e6)), sdk.NewCoin("ukava", sdk.NewInt(10000e6)), sdk.NewCoin("ukava", sdk.NewInt(1000e6)), sdk.NewCoin("usdx", sdk.NewInt(5000e6))},
		{"large usdx balance difference, vesting covers difference", sdk.NewCoin("usdx", sdk.NewInt(500e6)), sdk.NewCoin("usdx", sdk.NewInt(500000e6)), sdk.NewCoin("usdx", sdk.NewInt(5000e6)), sdk.NewCoin("ukava", sdk.NewInt(1000e6))},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			owner := suite.CreateAccount(sdk.Coins{})
			reserves := sdk.NewCoins(
				sdk.NewCoin("ukava", sdk.NewInt(100000e6)),
				sdk.NewCoin("usdx", sdk.NewInt(500000e6)),
			)
			totalShares := sdk.NewInt(30000e6)
			suite.setupPool(reserves, totalShares, owner.GetAddress())
			balance := sdk.NewCoins(tc.balanceA)
			vesting := sdk.NewCoins(tc.vestingA)
			requester := suite.CreateVestingAccount(balance, vesting)

			ctx := suite.App.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
			err := suite.Keeper.SwapForExactTokens(ctx, requester.GetAddress(), tc.coinA, tc.coinB, sdk.MustNewDecFromStr("0.1"))
			suite.Require().True(errors.Is(err, sdkerrors.ErrInsufficientFunds), fmt.Sprintf("got err %s", err))
		})
	}
}

func (suite *keeperTestSuite) TestSwapForExactTokens_PoolNotFound() {
	owner := suite.CreateAccount(sdk.Coins{})
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(1000e6)),
		sdk.NewCoin("usdx", sdk.NewInt(5000e6)),
	)
	totalShares := sdk.NewInt(3000e6)
	poolID := suite.setupPool(reserves, totalShares, owner.GetAddress())
	suite.Keeper.DeletePool(suite.Ctx, poolID)

	balance := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
		sdk.NewCoin("usdx", sdk.NewInt(10e6)),
	)
	requester := suite.NewAccountFromAddr(sdk.AccAddress("requester"), balance)
	coinA := sdk.NewCoin("ukava", sdk.NewInt(1e6))
	coinB := sdk.NewCoin("usdx", sdk.NewInt(5e6))

	err := suite.Keeper.SwapForExactTokens(suite.Ctx, requester.GetAddress(), coinA, coinB, sdk.MustNewDecFromStr("0.01"))
	suite.EqualError(err, "invalid pool: pool ukava/usdx not found")

	err = suite.Keeper.SwapForExactTokens(suite.Ctx, requester.GetAddress(), coinB, coinA, sdk.MustNewDecFromStr("0.01"))
	suite.EqualError(err, "invalid pool: pool ukava/usdx not found")
}

func (suite *keeperTestSuite) TestSwapForExactTokens_PanicOnInvalidPool() {
	owner := suite.CreateAccount(sdk.Coins{})
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(1000e6)),
		sdk.NewCoin("usdx", sdk.NewInt(5000e6)),
	)
	totalShares := sdk.NewInt(3000e6)
	poolID := suite.setupPool(reserves, totalShares, owner.GetAddress())

	poolRecord, found := suite.Keeper.GetPool(suite.Ctx, poolID)
	suite.Require().True(found, "expected pool record to exist")

	poolRecord.TotalShares = sdk.ZeroInt()
	suite.Keeper.SetPool(suite.Ctx, poolRecord)

	balance := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
		sdk.NewCoin("usdx", sdk.NewInt(10e6)),
	)
	requester := suite.NewAccountFromAddr(sdk.AccAddress("requester"), balance)
	coinA := sdk.NewCoin("ukava", sdk.NewInt(1e6))
	coinB := sdk.NewCoin("usdx", sdk.NewInt(5e6))

	suite.PanicsWithValue("invalid pool ukava/usdx: invalid pool: total shares must be greater than zero", func() {
		_ = suite.Keeper.SwapForExactTokens(suite.Ctx, requester.GetAddress(), coinA, coinB, sdk.MustNewDecFromStr("0.01"))
	}, "expected invalid pool record to panic")

	suite.PanicsWithValue("invalid pool ukava/usdx: invalid pool: total shares must be greater than zero", func() {
		_ = suite.Keeper.SwapForExactTokens(suite.Ctx, requester.GetAddress(), coinB, coinA, sdk.MustNewDecFromStr("0.01"))
	}, "expected invalid pool record to panic")
}

func (suite *keeperTestSuite) TestSwapForExactTokens_PanicOnInsuffientModuleAccFunds() {
	owner := suite.CreateAccount(sdk.Coins{})
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(1000e6)),
		sdk.NewCoin("usdx", sdk.NewInt(5000e6)),
	)
	totalShares := sdk.NewInt(3000e6)
	suite.setupPool(reserves, totalShares, owner.GetAddress())

	suite.RemoveCoinsFromModule(sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(1000e6)),
		sdk.NewCoin("usdx", sdk.NewInt(5000e6)),
	))

	balance := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
		sdk.NewCoin("usdx", sdk.NewInt(10e6)),
	)
	requester := suite.NewAccountFromAddr(sdk.AccAddress("requester"), balance)
	coinA := sdk.NewCoin("ukava", sdk.NewInt(1e6))
	coinB := sdk.NewCoin("usdx", sdk.NewInt(5e6))

	suite.Panics(func() {
		_ = suite.Keeper.SwapForExactTokens(suite.Ctx, requester.GetAddress(), coinA, coinB, sdk.MustNewDecFromStr("0.01"))
	}, "expected panic when module account does not have enough funds")

	suite.Panics(func() {
		_ = suite.Keeper.SwapForExactTokens(suite.Ctx, requester.GetAddress(), coinA, coinB, sdk.MustNewDecFromStr("0.01"))
	}, "expected panic when module account does not have enough funds")
}
