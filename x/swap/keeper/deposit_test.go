package keeper_test

import (
	"errors"
	"fmt"

	"github.com/kava-labs/kava/x/swap/types"

	sdkmath "cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (suite *keeperTestSuite) TestDeposit_CreatePool_PoolNotAllowed() {
	depositor := suite.CreateAccount(sdk.Coins{})
	amountA := sdk.NewCoin("ukava", sdkmath.NewInt(10e6))
	amountB := sdk.NewCoin("usdx", sdkmath.NewInt(50e6))

	err := suite.Keeper.Deposit(suite.Ctx, depositor.GetAddress(), amountA, amountB, sdk.MustNewDecFromStr("0.01"))
	suite.Require().EqualError(err, "can not create pool 'ukava:usdx': not allowed")
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
			balanceA: sdk.NewCoin("unuseddenom", sdk.ZeroInt()),
			balanceB: sdk.NewCoin("unuseddenom", sdk.ZeroInt()),
			depositA: sdk.NewCoin("ukava", sdkmath.NewInt(100)),
			depositB: sdk.NewCoin("usdx", sdkmath.NewInt(100)),
		},
		{
			name:     "low balance",
			balanceA: sdk.NewCoin("ukava", sdkmath.NewInt(1000000)),
			balanceB: sdk.NewCoin("usdx", sdkmath.NewInt(1000000)),
			depositA: sdk.NewCoin("ukava", sdkmath.NewInt(1000001)),
			depositB: sdk.NewCoin("usdx", sdkmath.NewInt(10000001)),
		},
		{
			name:     "large balance difference",
			balanceA: sdk.NewCoin("ukava", sdkmath.NewInt(100e6)),
			balanceB: sdk.NewCoin("usdx", sdkmath.NewInt(500e6)),
			depositA: sdk.NewCoin("ukava", sdkmath.NewInt(1000e6)),
			depositB: sdk.NewCoin("usdx", sdkmath.NewInt(5000e6)),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			pool := types.NewAllowedPool(tc.depositA.Denom, tc.depositB.Denom)
			suite.Require().NoError(pool.Validate())
			suite.Keeper.SetParams(suite.Ctx, types.NewParams(types.NewAllowedPools(pool), types.DefaultSwapFee))

			balance := sdk.NewCoins(tc.balanceA, tc.balanceB)
			depositor := suite.CreateAccount(balance)

			err := suite.Keeper.Deposit(suite.Ctx, depositor.GetAddress(), tc.depositA, tc.depositB, sdk.MustNewDecFromStr("0"))
			// TODO: wrap in module specific error?
			suite.Require().True(errors.Is(err, sdkerrors.ErrInsufficientFunds), fmt.Sprintf("got err %s", err))

			suite.SetupTest()
			// test deposit to existing pool insuffient funds
			err = suite.CreatePool(sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(10e6)), sdk.NewCoin("usdx", sdkmath.NewInt(50e6))))
			suite.Require().NoError(err)
			err = suite.Keeper.Deposit(suite.Ctx, depositor.GetAddress(), tc.depositA, tc.depositB, sdk.MustNewDecFromStr("10"))
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
			balanceA: sdk.NewCoin("ukava", sdk.ZeroInt()),
			balanceB: sdk.NewCoin("usdx", sdk.ZeroInt()),
			vestingA: sdk.NewCoin("ukava", sdkmath.NewInt(100)),
			vestingB: sdk.NewCoin("usdx", sdkmath.NewInt(100)),
			depositA: sdk.NewCoin("ukava", sdkmath.NewInt(100)),
			depositB: sdk.NewCoin("usdx", sdkmath.NewInt(100)),
		},
		{
			name:     "vesting matches balance exactly",
			balanceA: sdk.NewCoin("ukava", sdkmath.NewInt(1000000)),
			balanceB: sdk.NewCoin("usdx", sdkmath.NewInt(1000000)),
			vestingA: sdk.NewCoin("ukava", sdkmath.NewInt(1)),
			vestingB: sdk.NewCoin("usdx", sdkmath.NewInt(1)),
			depositA: sdk.NewCoin("ukava", sdkmath.NewInt(1000001)),
			depositB: sdk.NewCoin("usdx", sdkmath.NewInt(10000001)),
		},
		{
			name:     "large balance difference, vesting covers difference",
			balanceA: sdk.NewCoin("ukava", sdkmath.NewInt(100e6)),
			balanceB: sdk.NewCoin("usdx", sdkmath.NewInt(500e6)),
			vestingA: sdk.NewCoin("ukava", sdkmath.NewInt(1000e6)),
			vestingB: sdk.NewCoin("usdx", sdkmath.NewInt(5000e6)),
			depositA: sdk.NewCoin("ukava", sdkmath.NewInt(1000e6)),
			depositB: sdk.NewCoin("usdx", sdkmath.NewInt(5000e6)),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			pool := types.NewAllowedPool(tc.depositA.Denom, tc.depositB.Denom)
			suite.Require().NoError(pool.Validate())
			suite.Keeper.SetParams(suite.Ctx, types.NewParams(types.NewAllowedPools(pool), types.DefaultSwapFee))

			balance := sdk.NewCoins(tc.balanceA, tc.balanceB)
			vesting := sdk.NewCoins(tc.vestingA, tc.vestingB)
			depositor := suite.CreateVestingAccount(balance, vesting)

			// test create pool insuffient funds
			err := suite.Keeper.Deposit(suite.Ctx, depositor.GetAddress(), tc.depositA, tc.depositB, sdk.MustNewDecFromStr("0"))
			// TODO: wrap in module specific error?
			suite.Require().True(errors.Is(err, sdkerrors.ErrInsufficientFunds))

			suite.SetupTest()
			// test deposit to existing pool insuffient funds
			err = suite.CreatePool(sdk.NewCoins(sdk.NewCoin("ukava", sdkmath.NewInt(10e6)), sdk.NewCoin("usdx", sdkmath.NewInt(50e6))))
			suite.Require().NoError(err)
			err = suite.Keeper.Deposit(suite.Ctx, depositor.GetAddress(), tc.depositA, tc.depositB, sdk.MustNewDecFromStr("4"))
			suite.Require().True(errors.Is(err, sdkerrors.ErrInsufficientFunds))
		})
	}
}

func (suite *keeperTestSuite) TestDeposit_CreatePool() {
	pool := types.NewAllowedPool("ukava", "usdx")
	suite.Require().NoError(pool.Validate())
	suite.Keeper.SetParams(suite.Ctx, types.NewParams(types.NewAllowedPools(pool), types.DefaultSwapFee))

	amountA := sdk.NewCoin(pool.TokenA, sdkmath.NewInt(11e6))
	amountB := sdk.NewCoin(pool.TokenB, sdkmath.NewInt(51e6))
	balance := sdk.NewCoins(amountA, amountB)
	depositor := suite.CreateAccount(balance)

	depositA := sdk.NewCoin(pool.TokenA, sdkmath.NewInt(10e6))
	depositB := sdk.NewCoin(pool.TokenB, sdkmath.NewInt(50e6))
	deposit := sdk.NewCoins(depositA, depositB)

	err := suite.Keeper.Deposit(suite.Ctx, depositor.GetAddress(), depositA, depositB, sdk.MustNewDecFromStr("0"))
	suite.Require().NoError(err)
	suite.AccountBalanceEqual(depositor.GetAddress(), sdk.NewCoins(amountA.Sub(depositA), amountB.Sub(depositB)))
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
		sdk.NewCoin("ukava", sdkmath.NewInt(10e6)),
		sdk.NewCoin("usdx", sdkmath.NewInt(50e6)),
	)
	err := suite.CreatePool(reserves)
	suite.Require().NoError(err)

	balance := sdk.NewCoins(
		sdk.NewCoin("ukava", sdkmath.NewInt(5e6)),
		sdk.NewCoin("usdx", sdkmath.NewInt(5e6)),
	)
	depositor := suite.NewAccountFromAddr(sdk.AccAddress("new depositor-------"), balance) // TODO this is padded to the correct length, find a nicer way of creating test addresses

	depositA := sdk.NewCoin("usdx", balance.AmountOf("usdx"))
	depositB := sdk.NewCoin("ukava", balance.AmountOf("ukava"))

	ctx := suite.App.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})

	err = suite.Keeper.Deposit(ctx, depositor.GetAddress(), depositA, depositB, sdk.MustNewDecFromStr("4"))
	suite.Require().NoError(err)

	expectedDeposit := sdk.NewCoins(
		sdk.NewCoin("ukava", sdkmath.NewInt(1e6)),
		sdk.NewCoin("usdx", sdkmath.NewInt(5e6)),
	)

	expectedShareValue := sdk.NewCoins(
		sdk.NewCoin("ukava", sdkmath.NewInt(999999)),
		sdk.NewCoin("usdx", sdkmath.NewInt(4999998)),
	)

	suite.AccountBalanceEqual(depositor.GetAddress(), balance.Sub(expectedDeposit...))
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

func (suite *keeperTestSuite) TestDeposit_MultipleDeposit() {
	fundsToDeposit := sdk.NewCoins(
		sdk.NewCoin("ukava", sdkmath.NewInt(5e6)),
		sdk.NewCoin("usdx", sdkmath.NewInt(25e6)),
	)
	owner := suite.CreateAccount(fundsToDeposit)
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdkmath.NewInt(10e6)),
		sdk.NewCoin("usdx", sdkmath.NewInt(50e6)),
	)
	initialShares := sdkmath.NewInt(30e6)
	poolID := suite.setupPool(reserves, initialShares, owner.GetAddress())

	depositA := sdk.NewCoin("usdx", fundsToDeposit.AmountOf("usdx"))
	depositB := sdk.NewCoin("ukava", fundsToDeposit.AmountOf("ukava"))

	err := suite.Keeper.Deposit(suite.Ctx, owner.GetAddress(), depositA, depositB, sdk.MustNewDecFromStr("4"))
	suite.Require().NoError(err)

	totalDeposit := reserves.Add(fundsToDeposit...)
	totalShares := initialShares.Add(sdkmath.NewInt(15e6))

	suite.AccountBalanceEqual(owner.GetAddress(), sdk.Coins{})
	suite.ModuleAccountBalanceEqual(totalDeposit)
	suite.PoolLiquidityEqual(totalDeposit)
	suite.PoolDepositorSharesEqual(owner.GetAddress(), poolID, totalShares)

	suite.EventsContains(suite.Ctx.EventManager().Events(), sdk.NewEvent(
		types.EventTypeSwapDeposit,
		sdk.NewAttribute(types.AttributeKeyPoolID, poolID),
		sdk.NewAttribute(types.AttributeKeyDepositor, owner.GetAddress().String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, fundsToDeposit.String()),
		sdk.NewAttribute(types.AttributeKeyShares, "15000000"),
	))
}

func (suite *keeperTestSuite) TestDeposit_Slippage() {
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdkmath.NewInt(10e6)),
		sdk.NewCoin("usdx", sdkmath.NewInt(50e6)),
	)

	testCases := []struct {
		depositA   sdk.Coin
		depositB   sdk.Coin
		slippage   sdk.Dec
		shouldFail bool
	}{
		{sdk.NewCoin("usdx", sdkmath.NewInt(5e6)), sdk.NewCoin("ukava", sdkmath.NewInt(5e6)), sdk.MustNewDecFromStr("0.7"), true},
		{sdk.NewCoin("usdx", sdkmath.NewInt(5e6)), sdk.NewCoin("ukava", sdkmath.NewInt(5e6)), sdk.MustNewDecFromStr("0.8"), true},
		{sdk.NewCoin("ukava", sdkmath.NewInt(5e6)), sdk.NewCoin("usdx", sdkmath.NewInt(5e6)), sdk.MustNewDecFromStr("3"), true},
		{sdk.NewCoin("ukava", sdkmath.NewInt(5e6)), sdk.NewCoin("usdx", sdkmath.NewInt(5e6)), sdk.MustNewDecFromStr("4"), false},
		{sdk.NewCoin("ukava", sdkmath.NewInt(1e6)), sdk.NewCoin("usdx", sdkmath.NewInt(5e6)), sdk.MustNewDecFromStr("0"), false},
		{sdk.NewCoin("ukava", sdkmath.NewInt(1e6)), sdk.NewCoin("usdx", sdkmath.NewInt(4e6)), sdk.MustNewDecFromStr("0.25"), false},
		{sdk.NewCoin("ukava", sdkmath.NewInt(1e6)), sdk.NewCoin("usdx", sdkmath.NewInt(4e6)), sdk.MustNewDecFromStr("0.2"), true},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("depositA=%s depositB=%s slippage=%s", tc.depositA, tc.depositB, tc.slippage), func() {
			suite.SetupTest()

			err := suite.CreatePool(reserves)
			suite.Require().NoError(err)

			balance := sdk.NewCoins(
				sdk.NewCoin("ukava", sdkmath.NewInt(100e6)),
				sdk.NewCoin("usdx", sdkmath.NewInt(100e6)),
			)
			depositor := suite.CreateAccount(balance)

			ctx := suite.App.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})

			err = suite.Keeper.Deposit(ctx, depositor.GetAddress(), tc.depositA, tc.depositB, tc.slippage)
			if tc.shouldFail {
				suite.Require().Error(err)
				suite.Contains(err.Error(), "slippage exceeded")
			} else {
				suite.NoError(err)
			}
		})
	}
}

func (suite *keeperTestSuite) TestDeposit_InsufficientLiquidity() {
	testCases := []struct {
		poolA      sdk.Coin
		poolB      sdk.Coin
		poolShares sdkmath.Int
		depositA   sdk.Coin
		depositB   sdk.Coin
	}{
		// test deposit amount truncating to zero
		{sdk.NewCoin("ukava", sdkmath.NewInt(10e6)), sdk.NewCoin("usdx", sdkmath.NewInt(50e6)), sdkmath.NewInt(40e6), sdk.NewCoin("ukava", sdkmath.NewInt(1)), sdk.NewCoin("usdx", sdkmath.NewInt(1))},
		// test share value rounding to zero
		{sdk.NewCoin("ukava", sdkmath.NewInt(10e6)), sdk.NewCoin("usdx", sdkmath.NewInt(10e6)), sdkmath.NewInt(100), sdk.NewCoin("ukava", sdkmath.NewInt(1000)), sdk.NewCoin("usdx", sdkmath.NewInt(1000))},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("depositA=%s depositB=%s", tc.depositA, tc.depositB), func() {
			suite.SetupTest()

			record := types.PoolRecord{
				PoolID:      types.PoolID("ukava", "usdx"),
				ReservesA:   tc.poolA,
				ReservesB:   tc.poolB,
				TotalShares: tc.poolShares,
			}

			suite.Keeper.SetPool(suite.Ctx, record)

			balance := sdk.NewCoins(tc.depositA, tc.depositB)
			depositor := suite.CreateAccount(balance)

			err := suite.Keeper.Deposit(suite.Ctx, depositor.GetAddress(), tc.depositA, tc.depositB, sdk.MustNewDecFromStr("10"))
			suite.EqualError(err, "deposit must be increased: insufficient liquidity")
		})
	}
}
