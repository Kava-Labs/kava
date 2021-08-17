package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/swap/types"
)

func (suite *keeperTestSuite) TestWithdraw_AllShares() {
	owner := suite.CreateAccount(sdk.Coins{})
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
		sdk.NewCoin("usdx", sdk.NewInt(50e6)),
	)
	totalShares := sdk.NewInt(30e6)
	poolID := suite.setupPool(reserves, totalShares, owner.GetAddress())

	err := suite.Keeper.Withdraw(suite.Ctx, owner.GetAddress(), totalShares, reserves[0], reserves[1])
	suite.Require().NoError(err)

	suite.PoolDeleted(reserves[0].Denom, reserves[1].Denom)
	suite.PoolSharesDeleted(owner.GetAddress(), reserves[0].Denom, reserves[1].Denom)
	suite.AccountBalanceEqual(owner, reserves)
	suite.ModuleAccountBalanceEqual(sdk.Coins(nil))

	suite.EventsContains(suite.Ctx.EventManager().Events(), sdk.NewEvent(
		types.EventTypeSwapWithdraw,
		sdk.NewAttribute(types.AttributeKeyPoolID, poolID),
		sdk.NewAttribute(types.AttributeKeyOwner, owner.GetAddress().String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, reserves.String()),
		sdk.NewAttribute(types.AttributeKeyShares, totalShares.String()),
	))
}

func (suite *keeperTestSuite) TestWithdraw_PartialShares() {
	owner := suite.CreateAccount(sdk.Coins{})
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
		sdk.NewCoin("usdx", sdk.NewInt(50e6)),
	)
	totalShares := sdk.NewInt(30e6)
	poolID := suite.setupPool(reserves, totalShares, owner.GetAddress())

	sharesToWithdraw := sdk.NewInt(15e6)
	minCoinA := sdk.NewCoin("usdx", sdk.NewInt(25e6))
	minCoinB := sdk.NewCoin("ukava", sdk.NewInt(5e6))

	err := suite.Keeper.Withdraw(suite.Ctx, owner.GetAddress(), sharesToWithdraw, minCoinA, minCoinB)
	suite.Require().NoError(err)

	sharesLeft := totalShares.Sub(sharesToWithdraw)
	reservesLeft := sdk.NewCoins(reserves[0].Sub(minCoinB), reserves[1].Sub(minCoinA))

	suite.PoolShareTotalEqual(poolID, sharesLeft)
	suite.PoolDepositorSharesEqual(owner.GetAddress(), poolID, sharesLeft)
	suite.PoolReservesEqual(poolID, reservesLeft)
	suite.AccountBalanceEqual(owner, sdk.NewCoins(minCoinA, minCoinB))
	suite.ModuleAccountBalanceEqual(reservesLeft)

	suite.EventsContains(suite.Ctx.EventManager().Events(), sdk.NewEvent(
		types.EventTypeSwapWithdraw,
		sdk.NewAttribute(types.AttributeKeyPoolID, poolID),
		sdk.NewAttribute(types.AttributeKeyOwner, owner.GetAddress().String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, sdk.NewCoins(minCoinA, minCoinB).String()),
		sdk.NewAttribute(types.AttributeKeyShares, sharesToWithdraw.String()),
	))
}

func (suite *keeperTestSuite) TestWithdraw_NoSharesOwned() {
	owner := suite.CreateAccount(sdk.Coins{})
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
		sdk.NewCoin("usdx", sdk.NewInt(50e6)),
	)
	totalShares := sdk.NewInt(30e6)
	poolID := suite.setupPool(reserves, totalShares, owner.GetAddress())

	accWithNoDeposit := sdk.AccAddress("some account")

	err := suite.Keeper.Withdraw(suite.Ctx, accWithNoDeposit, totalShares, reserves[0], reserves[1])
	suite.EqualError(err, fmt.Sprintf("deposit not found: no deposit for account %s and pool %s", accWithNoDeposit.String(), poolID))
}

func (suite *keeperTestSuite) TestWithdraw_GreaterThanSharesOwned() {
	owner := suite.CreateAccount(sdk.Coins{})
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
		sdk.NewCoin("usdx", sdk.NewInt(50e6)),
	)
	totalShares := sdk.NewInt(30e6)
	suite.setupPool(reserves, totalShares, owner.GetAddress())

	sharesToWithdraw := totalShares.Add(sdk.OneInt())
	err := suite.Keeper.Withdraw(suite.Ctx, owner.GetAddress(), sharesToWithdraw, reserves[0], reserves[1])
	suite.EqualError(err, fmt.Sprintf("invalid shares: withdraw of %s shares greater than %s shares owned", sharesToWithdraw, totalShares))
}

func (suite *keeperTestSuite) TestWithdraw_MinWithdraw() {
	owner := suite.CreateAccount(sdk.Coins{})
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
		sdk.NewCoin("usdx", sdk.NewInt(50e6)),
	)
	totalShares := sdk.NewInt(30e6)

	testCases := []struct {
		shares     sdk.Int
		minCoinA   sdk.Coin
		minCoinB   sdk.Coin
		shouldFail bool
	}{
		{sdk.NewInt(1), sdk.NewCoin("ukava", sdk.NewInt(1)), sdk.NewCoin("usdx", sdk.NewInt(1)), true},
		{sdk.NewInt(1), sdk.NewCoin("usdx", sdk.NewInt(5)), sdk.NewCoin("ukava", sdk.NewInt(1)), true},

		{sdk.NewInt(2), sdk.NewCoin("ukava", sdk.NewInt(1)), sdk.NewCoin("usdx", sdk.NewInt(1)), true},
		{sdk.NewInt(2), sdk.NewCoin("usdx", sdk.NewInt(5)), sdk.NewCoin("ukava", sdk.NewInt(1)), true},

		{sdk.NewInt(3), sdk.NewCoin("ukava", sdk.NewInt(1)), sdk.NewCoin("usdx", sdk.NewInt(5)), false},
		{sdk.NewInt(3), sdk.NewCoin("usdx", sdk.NewInt(5)), sdk.NewCoin("ukava", sdk.NewInt(1)), false},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("shares=%s minCoinA=%s minCoinB=%s", tc.shares, tc.minCoinA, tc.minCoinB), func() {
			suite.SetupTest()
			suite.setupPool(reserves, totalShares, owner.GetAddress())

			err := suite.Keeper.Withdraw(suite.Ctx, owner.GetAddress(), tc.shares, tc.minCoinA, tc.minCoinB)
			if tc.shouldFail {
				suite.EqualError(err, "insufficient liquidity: shares must be increased")
			} else {
				suite.NoError(err, "expected no liquidity error")
			}
		})
	}
}

func (suite *keeperTestSuite) TestWithdraw_BelowMinimum() {
	owner := suite.CreateAccount(sdk.Coins{})
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
		sdk.NewCoin("usdx", sdk.NewInt(50e6)),
	)
	totalShares := sdk.NewInt(30e6)

	testCases := []struct {
		shares     sdk.Int
		minCoinA   sdk.Coin
		minCoinB   sdk.Coin
		shouldFail bool
	}{
		{sdk.NewInt(15e6), sdk.NewCoin("ukava", sdk.NewInt(5000001)), sdk.NewCoin("usdx", sdk.NewInt(25e6)), true},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("shares=%s minCoinA=%s minCoinB=%s", tc.shares, tc.minCoinA, tc.minCoinB), func() {
			suite.SetupTest()
			suite.setupPool(reserves, totalShares, owner.GetAddress())

			err := suite.Keeper.Withdraw(suite.Ctx, owner.GetAddress(), tc.shares, tc.minCoinA, tc.minCoinB)
			if tc.shouldFail {
				suite.EqualError(err, "slippage exceeded: minimum withdraw not met")
			} else {
				suite.NoError(err, "expected no slippage error")
			}
		})
	}
}

func (suite *keeperTestSuite) TestWithdraw_ErrOnMissingPool() {
	owner := suite.CreateAccount(sdk.Coins{})
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
		sdk.NewCoin("usdx", sdk.NewInt(50e6)),
	)
	totalShares := sdk.NewInt(30e6)
	poolID := suite.setupPool(reserves, totalShares, owner.GetAddress())

	suite.Keeper.DeletePool(suite.Ctx, poolID)

	err := suite.Keeper.Withdraw(suite.Ctx, owner.GetAddress(), totalShares, reserves[0], reserves[1])
	suite.Error(err)
	suite.Equal("pool not found: ukava:usdx", err.Error())
}

func (suite *keeperTestSuite) TestWithdraw_ErrOnInvalidPool() {
	owner := suite.CreateAccount(sdk.Coins{})
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
		sdk.NewCoin("usdx", sdk.NewInt(50e6)),
	)
	totalShares := sdk.NewInt(30e6)
	poolID := suite.setupPool(reserves, totalShares, owner.GetAddress())

	poolRecord, found := suite.Keeper.GetPool(suite.Ctx, poolID)
	suite.Require().True(found, "expected pool record to exist")

	poolRecord.TotalShares = sdk.ZeroInt()
	suite.Keeper.SetPool_Raw(suite.Ctx, poolRecord)

	err := suite.Keeper.Withdraw(suite.Ctx, owner.GetAddress(), totalShares, reserves[0], reserves[1])
	suite.Error(err)
	suite.Equal("invalid pool: ukava:usdx: invalid pool: total shares must be greater than zero", err.Error())
}

func (suite *keeperTestSuite) TestWithdraw_ErrOnModuleInsufficientFunds() {
	owner := suite.CreateAccount(sdk.Coins{})
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
		sdk.NewCoin("usdx", sdk.NewInt(50e6)),
	)
	totalShares := sdk.NewInt(30e6)
	suite.setupPool(reserves, totalShares, owner.GetAddress())

	suite.RemoveCoinsFromModule(sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(1e6)),
		sdk.NewCoin("usdx", sdk.NewInt(5e6)),
	))

	err := suite.Keeper.Withdraw(suite.Ctx, owner.GetAddress(), totalShares, reserves[0], reserves[1])
	suite.Error(err)
}
