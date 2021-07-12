package keeper_test

import (
	//"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/swap/types"
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
