package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/kava-labs/kava/x/swap/keeper"
	"github.com/kava-labs/kava/x/swap/testutil"
	"github.com/kava-labs/kava/x/swap/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/tendermint/tendermint/crypto"
)

var swapModuleAccountAddress = sdk.AccAddress(crypto.AddressHash([]byte(types.ModuleAccountName)))

type msgServerTestSuite struct {
	testutil.Suite
	msgServer types.MsgServer
}

func (suite *msgServerTestSuite) SetupTest() {
	suite.Suite.SetupTest()
	suite.msgServer = keeper.NewMsgServerImpl(suite.Keeper)
}

func (suite *msgServerTestSuite) TestDeposit_CreatePool() {
	pool := types.NewAllowedPool("ukava", "usdx")
	suite.Require().NoError(pool.Validate())
	suite.Keeper.SetParams(suite.Ctx, types.NewParams(types.AllowedPools{pool}, types.DefaultSwapFee))

	balance := sdk.NewCoins(
		sdk.NewCoin(pool.TokenA, sdk.NewInt(10e6)),
		sdk.NewCoin(pool.TokenB, sdk.NewInt(50e6)),
	)
	depositor := suite.NewAccountFromAddr(sdk.AccAddress("new depositor-------"), balance)

	deposit := types.NewMsgDeposit(
		depositor.GetAddress().String(),
		suite.BankKeeper.GetBalance(suite.Ctx, depositor.GetAddress(), pool.TokenA),
		suite.BankKeeper.GetBalance(suite.Ctx, depositor.GetAddress(), pool.TokenB),
		sdk.MustNewDecFromStr("0.01"),
		time.Now().Add(10*time.Minute).Unix(),
	)

	res, err := suite.msgServer.Deposit(sdk.WrapSDKContext(suite.Ctx), deposit)
	suite.Require().Equal(&types.MsgDepositResponse{}, res)
	suite.Require().NoError(err)

	suite.AccountBalanceEqual(depositor.GetAddress(), sdk.Coins{})
	suite.ModuleAccountBalanceEqual(balance)
	suite.PoolLiquidityEqual(balance)
	suite.PoolShareValueEqual(depositor, pool, balance)

	suite.EventsContains(suite.GetEvents(), sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		sdk.NewAttribute(sdk.AttributeKeySender, depositor.GetAddress().String()),
	))

	suite.EventsContains(suite.GetEvents(), sdk.NewEvent(
		bank.EventTypeTransfer,
		sdk.NewAttribute(bank.AttributeKeyRecipient, swapModuleAccountAddress.String()),
		sdk.NewAttribute(bank.AttributeKeySender, depositor.GetAddress().String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, balance.String()),
	))

	suite.EventsContains(suite.GetEvents(), sdk.NewEvent(
		types.EventTypeSwapDeposit,
		sdk.NewAttribute(types.AttributeKeyPoolID, types.PoolID(pool.TokenA, pool.TokenB)),
		sdk.NewAttribute(types.AttributeKeyDepositor, depositor.GetAddress().String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, balance.String()),
		sdk.NewAttribute(types.AttributeKeyShares, "22360679"),
	))
}

func (suite *msgServerTestSuite) TestDeposit_DeadlineExceeded() {
	pool := types.NewAllowedPool("ukava", "usdx")
	suite.Require().NoError(pool.Validate())
	suite.Keeper.SetParams(suite.Ctx, types.NewParams(types.AllowedPools{pool}, types.DefaultSwapFee))

	balance := sdk.NewCoins(
		sdk.NewCoin(pool.TokenA, sdk.NewInt(10e6)),
		sdk.NewCoin(pool.TokenB, sdk.NewInt(50e6)),
	)
	depositor := suite.NewAccountFromAddr(sdk.AccAddress("new depositor-------"), balance)

	deposit := types.NewMsgDeposit(
		depositor.GetAddress().String(),
		suite.BankKeeper.GetBalance(suite.Ctx, depositor.GetAddress(), pool.TokenA),
		suite.BankKeeper.GetBalance(suite.Ctx, depositor.GetAddress(), pool.TokenB),
		sdk.MustNewDecFromStr("0.01"),
		suite.Ctx.BlockTime().Add(-1*time.Second).Unix(),
	)

	res, err := suite.msgServer.Deposit(sdk.WrapSDKContext(suite.Ctx), deposit)
	suite.Require().Nil(res)
	suite.EqualError(err, fmt.Sprintf("block time %d >= deadline %d: deadline exceeded", suite.Ctx.BlockTime().Unix(), deposit.GetDeadline().Unix()))
	suite.Nil(res)
}

func (suite *msgServerTestSuite) TestDeposit_ExistingPool() {
	pool := types.NewAllowedPool("ukava", "usdx")
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
		sdk.NewCoin("usdx", sdk.NewInt(50e6)),
	)
	err := suite.CreatePool(reserves)
	suite.Require().NoError(err)

	balance := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(1e6)),
		sdk.NewCoin("usdx", sdk.NewInt(5e6)),
	)
	depositor := suite.NewAccountFromAddr(sdk.AccAddress("new depositor-------"), balance)

	deposit := types.NewMsgDeposit(
		depositor.GetAddress().String(),
		suite.BankKeeper.GetBalance(suite.Ctx, depositor.GetAddress(), "usdx"),
		suite.BankKeeper.GetBalance(suite.Ctx, depositor.GetAddress(), "ukava"),
		sdk.MustNewDecFromStr("0.01"),
		time.Now().Add(10*time.Minute).Unix(),
	)

	res, err := suite.msgServer.Deposit(sdk.WrapSDKContext(suite.Ctx), deposit)
	suite.Require().Equal(&types.MsgDepositResponse{}, res)
	suite.Require().NoError(err)

	expectedDeposit := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(1e6)),
		sdk.NewCoin("usdx", sdk.NewInt(5e6)),
	)

	expectedShareValue := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(999999)),
		sdk.NewCoin("usdx", sdk.NewInt(4999998)),
	)

	// Use sdk.NewCoins to remove zero coins, otherwise it will compare sdk.Coins(nil) with sdk.Coins{}
	suite.AccountBalanceEqual(depositor.GetAddress(), sdk.NewCoins(balance.Sub(expectedDeposit)...))
	suite.ModuleAccountBalanceEqual(reserves.Add(expectedDeposit...))
	suite.PoolLiquidityEqual(reserves.Add(expectedDeposit...))
	suite.PoolShareValueEqual(depositor, pool, expectedShareValue)

	suite.EventsContains(suite.GetEvents(), sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		sdk.NewAttribute(sdk.AttributeKeySender, depositor.GetAddress().String()),
	))

	suite.EventsContains(suite.GetEvents(), sdk.NewEvent(
		bank.EventTypeTransfer,
		sdk.NewAttribute(bank.AttributeKeyRecipient, swapModuleAccountAddress.String()),
		sdk.NewAttribute(bank.AttributeKeySender, depositor.GetAddress().String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, expectedDeposit.String()),
	))

	suite.EventsContains(suite.GetEvents(), sdk.NewEvent(
		types.EventTypeSwapDeposit,
		sdk.NewAttribute(types.AttributeKeyPoolID, types.PoolID(pool.TokenA, pool.TokenB)),
		sdk.NewAttribute(types.AttributeKeyDepositor, depositor.GetAddress().String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, expectedDeposit.String()),
		sdk.NewAttribute(types.AttributeKeyShares, "2236067"),
	))
}

func (suite *msgServerTestSuite) TestDeposit_ExistingPool_SlippageFailure() {
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
	depositor := suite.NewAccountFromAddr(sdk.AccAddress("new depositor-------"), balance)

	deposit := types.NewMsgDeposit(
		depositor.GetAddress().String(),
		suite.BankKeeper.GetBalance(suite.Ctx, depositor.GetAddress(), "usdx"),
		suite.BankKeeper.GetBalance(suite.Ctx, depositor.GetAddress(), "ukava"),
		sdk.MustNewDecFromStr("0.01"),
		time.Now().Add(10*time.Minute).Unix(),
	)

	res, err := suite.msgServer.Deposit(sdk.WrapSDKContext(suite.Ctx), deposit)
	suite.Require().Nil(res)
	suite.EqualError(err, "slippage 4.000000000000000000 > limit 0.010000000000000000: slippage exceeded")
	suite.Nil(res)
}

func (suite *msgServerTestSuite) TestWithdraw_AllShares() {
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
		sdk.NewCoin("usdx", sdk.NewInt(50e6)),
	)
	depositor := suite.NewAccountFromAddr(sdk.AccAddress("new depositor-------"), reserves)
	pool := types.NewAllowedPool(reserves[0].Denom, reserves[1].Denom)
	suite.Require().NoError(pool.Validate())
	suite.Keeper.SetParams(suite.Ctx, types.NewParams(types.AllowedPools{pool}, types.DefaultSwapFee))

	err := suite.Keeper.Deposit(suite.Ctx, depositor.GetAddress(), reserves[0], reserves[1], sdk.MustNewDecFromStr("1"))
	suite.Require().NoError(err)

	withdraw := types.NewMsgWithdraw(
		depositor.GetAddress().String(),
		sdk.NewInt(22360679),
		reserves[0],
		reserves[1],
		time.Now().Add(10*time.Minute).Unix(),
	)

	suite.Ctx = suite.App.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})
	res, err := suite.msgServer.Withdraw(sdk.WrapSDKContext(suite.Ctx), withdraw)
	suite.Require().Equal(&types.MsgWithdrawResponse{}, res)
	suite.Require().NoError(err)

	suite.AccountBalanceEqual(depositor.GetAddress(), reserves)
	suite.ModuleAccountBalanceEqual(sdk.Coins{})
	suite.PoolDeleted("ukava", "usdx")
	suite.PoolSharesDeleted(depositor.GetAddress(), "ukava", "usdx")

	suite.EventsContains(suite.GetEvents(), sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		sdk.NewAttribute(sdk.AttributeKeySender, depositor.GetAddress().String()),
	))

	suite.EventsContains(suite.GetEvents(), sdk.NewEvent(
		bank.EventTypeTransfer,
		sdk.NewAttribute(bank.AttributeKeyRecipient, depositor.GetAddress().String()),
		sdk.NewAttribute(bank.AttributeKeySender, swapModuleAccountAddress.String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, reserves.String()),
	))

	suite.EventsContains(suite.GetEvents(), sdk.NewEvent(
		types.EventTypeSwapWithdraw,
		sdk.NewAttribute(types.AttributeKeyPoolID, types.PoolID(pool.TokenA, pool.TokenB)),
		sdk.NewAttribute(types.AttributeKeyOwner, depositor.GetAddress().String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, reserves.String()),
		sdk.NewAttribute(types.AttributeKeyShares, "22360679"),
	))
}

func (suite *msgServerTestSuite) TestWithdraw_PartialShares() {
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
		sdk.NewCoin("usdx", sdk.NewInt(50e6)),
	)
	depositor := suite.NewAccountFromAddr(sdk.AccAddress("new depositor-------"), reserves)
	pool := types.NewAllowedPool(reserves[0].Denom, reserves[1].Denom)
	suite.Require().NoError(pool.Validate())
	suite.Keeper.SetParams(suite.Ctx, types.NewParams(types.AllowedPools{pool}, types.DefaultSwapFee))

	err := suite.Keeper.Deposit(suite.Ctx, depositor.GetAddress(), reserves[0], reserves[1], sdk.MustNewDecFromStr("1"))
	suite.Require().NoError(err)

	minTokenA := sdk.NewCoin("ukava", sdk.NewInt(4999999))
	minTokenB := sdk.NewCoin("usdx", sdk.NewInt(24999998))

	withdraw := types.NewMsgWithdraw(
		depositor.GetAddress().String(),
		sdk.NewInt(11180339),
		minTokenA,
		minTokenB,
		time.Now().Add(10*time.Minute).Unix(),
	)

	suite.Ctx = suite.App.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})
	res, err := suite.msgServer.Withdraw(sdk.WrapSDKContext(suite.Ctx), withdraw)
	suite.Require().Equal(&types.MsgWithdrawResponse{}, res)
	suite.Require().NoError(err)

	expectedCoinsReceived := sdk.NewCoins(minTokenA, minTokenB)

	suite.AccountBalanceEqual(depositor.GetAddress(), expectedCoinsReceived)
	suite.ModuleAccountBalanceEqual(reserves.Sub(expectedCoinsReceived))
	suite.PoolLiquidityEqual(reserves.Sub(expectedCoinsReceived))
	suite.PoolShareValueEqual(depositor, types.NewAllowedPool("ukava", "usdx"), reserves.Sub(expectedCoinsReceived))

	suite.EventsContains(suite.GetEvents(), sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		sdk.NewAttribute(sdk.AttributeKeySender, depositor.GetAddress().String()),
	))

	suite.EventsContains(suite.GetEvents(), sdk.NewEvent(
		bank.EventTypeTransfer,
		sdk.NewAttribute(bank.AttributeKeyRecipient, depositor.GetAddress().String()),
		sdk.NewAttribute(bank.AttributeKeySender, swapModuleAccountAddress.String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, expectedCoinsReceived.String()),
	))

	suite.EventsContains(suite.GetEvents(), sdk.NewEvent(
		types.EventTypeSwapWithdraw,
		sdk.NewAttribute(types.AttributeKeyPoolID, types.PoolID(pool.TokenA, pool.TokenB)),
		sdk.NewAttribute(types.AttributeKeyOwner, depositor.GetAddress().String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, expectedCoinsReceived.String()),
		sdk.NewAttribute(types.AttributeKeyShares, "11180339"),
	))
}

func (suite *msgServerTestSuite) TestWithdraw_SlippageFailure() {
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
		sdk.NewCoin("usdx", sdk.NewInt(50e6)),
	)
	depositor := suite.NewAccountFromAddr(sdk.AccAddress("new depositor-------"), reserves)
	pool := types.NewAllowedPool(reserves[0].Denom, reserves[1].Denom)
	suite.Require().NoError(pool.Validate())
	suite.Keeper.SetParams(suite.Ctx, types.NewParams(types.AllowedPools{pool}, types.DefaultSwapFee))

	err := suite.Keeper.Deposit(suite.Ctx, depositor.GetAddress(), reserves[0], reserves[1], sdk.MustNewDecFromStr("1"))
	suite.Require().NoError(err)

	minTokenA := sdk.NewCoin("ukava", sdk.NewInt(5e6))
	minTokenB := sdk.NewCoin("usdx", sdk.NewInt(25e6))

	withdraw := types.NewMsgWithdraw(
		depositor.GetAddress().String(),
		sdk.NewInt(11180339),
		minTokenA,
		minTokenB,
		time.Now().Add(10*time.Minute).Unix(),
	)

	res, err := suite.msgServer.Withdraw(sdk.WrapSDKContext(suite.Ctx), withdraw)
	suite.Require().Nil(res)
	suite.EqualError(err, "minimum withdraw not met: slippage exceeded")
	suite.Nil(res)
}

func (suite *msgServerTestSuite) TestWithdraw_DeadlineExceeded() {
	balance := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
		sdk.NewCoin("usdx", sdk.NewInt(50e6)),
	)
	from := suite.NewAccountFromAddr(sdk.AccAddress("from----------------"), balance)

	withdraw := types.NewMsgWithdraw(
		from.GetAddress().String(),
		sdk.NewInt(2e6),
		sdk.NewCoin("ukava", sdk.NewInt(1e6)),
		sdk.NewCoin("usdx", sdk.NewInt(5e6)),
		suite.Ctx.BlockTime().Add(-1*time.Second).Unix(),
	)

	res, err := suite.msgServer.Withdraw(sdk.WrapSDKContext(suite.Ctx), withdraw)
	suite.Require().Nil(res)
	suite.EqualError(err, fmt.Sprintf("block time %d >= deadline %d: deadline exceeded", suite.Ctx.BlockTime().Unix(), withdraw.GetDeadline().Unix()))
	suite.Nil(res)
}

func (suite *msgServerTestSuite) TestSwapExactForTokens() {
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(1000e6)),
		sdk.NewCoin("usdx", sdk.NewInt(5000e6)),
	)
	err := suite.CreatePool(reserves)
	suite.Require().NoError(err)

	balance := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
	)
	requester := suite.NewAccountFromAddr(sdk.AccAddress("requester-----------"), balance)

	swapInput := sdk.NewCoin("ukava", sdk.NewInt(1e6))
	swapMsg := types.NewMsgSwapExactForTokens(
		requester.GetAddress().String(),
		swapInput,
		sdk.NewCoin("usdx", sdk.NewInt(5e6)),
		sdk.MustNewDecFromStr("0.01"),
		time.Now().Add(10*time.Minute).Unix(),
	)

	suite.Ctx = suite.App.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})
	res, err := suite.msgServer.SwapExactForTokens(sdk.WrapSDKContext(suite.Ctx), swapMsg)
	suite.Require().Equal(&types.MsgSwapExactForTokensResponse{}, res)
	suite.Require().NoError(err)

	expectedSwapOutput := sdk.NewCoin("usdx", sdk.NewInt(4980034))

	suite.AccountBalanceEqual(requester.GetAddress(), balance.Sub(sdk.NewCoins(swapInput)).Add(expectedSwapOutput))
	suite.ModuleAccountBalanceEqual(reserves.Add(swapInput).Sub(sdk.NewCoins(expectedSwapOutput)))
	suite.PoolLiquidityEqual(reserves.Add(swapInput).Sub(sdk.NewCoins(expectedSwapOutput)))

	suite.EventsContains(suite.GetEvents(), sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		sdk.NewAttribute(sdk.AttributeKeySender, requester.GetAddress().String()),
	))

	suite.EventsContains(suite.GetEvents(), sdk.NewEvent(
		bank.EventTypeTransfer,
		sdk.NewAttribute(bank.AttributeKeyRecipient, swapModuleAccountAddress.String()),
		sdk.NewAttribute(bank.AttributeKeySender, requester.GetAddress().String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, swapInput.String()),
	))

	suite.EventsContains(suite.GetEvents(), sdk.NewEvent(
		bank.EventTypeTransfer,
		sdk.NewAttribute(bank.AttributeKeyRecipient, requester.GetAddress().String()),
		sdk.NewAttribute(bank.AttributeKeySender, swapModuleAccountAddress.String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, expectedSwapOutput.String()),
	))

	suite.EventsContains(suite.GetEvents(), sdk.NewEvent(
		types.EventTypeSwapTrade,
		sdk.NewAttribute(types.AttributeKeyPoolID, types.PoolID("ukava", "usdx")),
		sdk.NewAttribute(types.AttributeKeyRequester, requester.GetAddress().String()),
		sdk.NewAttribute(types.AttributeKeySwapInput, swapInput.String()),
		sdk.NewAttribute(types.AttributeKeySwapOutput, expectedSwapOutput.String()),
		sdk.NewAttribute(types.AttributeKeyFeePaid, "3000ukava"),
		sdk.NewAttribute(types.AttributeKeyExactDirection, "input"),
	))
}

func (suite *msgServerTestSuite) TestSwapExactForTokens_SlippageFailure() {
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(1000e6)),
		sdk.NewCoin("usdx", sdk.NewInt(5000e6)),
	)
	err := suite.CreatePool(reserves)
	suite.Require().NoError(err)

	balance := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(100e6)),
	)
	requester := suite.NewAccountFromAddr(sdk.AccAddress("requester-----------"), balance)

	swapInput := sdk.NewCoin("ukava", sdk.NewInt(1e6))
	swapMsg := types.NewMsgSwapExactForTokens(
		requester.GetAddress().String(),
		swapInput,
		sdk.NewCoin("usdx", sdk.NewInt(5030338)),
		sdk.MustNewDecFromStr("0.01"),
		time.Now().Add(10*time.Minute).Unix(),
	)

	suite.Ctx = suite.App.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})
	res, err := suite.msgServer.SwapExactForTokens(sdk.WrapSDKContext(suite.Ctx), swapMsg)
	suite.Require().Nil(res)
	suite.EqualError(err, "slippage 0.010000123252155223 > limit 0.010000000000000000: slippage exceeded")
	suite.Nil(res)
}

func (suite *msgServerTestSuite) TestSwapExactForTokens_DeadlineExceeded() {
	balance := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
	)
	requester := suite.NewAccountFromAddr(sdk.AccAddress("requester-----------"), balance)

	swapMsg := types.NewMsgSwapExactForTokens(
		requester.GetAddress().String(),
		sdk.NewCoin("ukava", sdk.NewInt(5e6)),
		sdk.NewCoin("usdx", sdk.NewInt(25e5)),
		sdk.MustNewDecFromStr("0.01"),
		suite.Ctx.BlockTime().Add(-1*time.Second).Unix(),
	)

	res, err := suite.msgServer.SwapExactForTokens(sdk.WrapSDKContext(suite.Ctx), swapMsg)
	suite.Require().Nil(res)
	suite.EqualError(err, fmt.Sprintf("block time %d >= deadline %d: deadline exceeded", suite.Ctx.BlockTime().Unix(), swapMsg.GetDeadline().Unix()))
	suite.Nil(res)
}

func (suite *msgServerTestSuite) TestSwapForExactTokens() {
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(1000e6)),
		sdk.NewCoin("usdx", sdk.NewInt(5000e6)),
	)
	err := suite.CreatePool(reserves)
	suite.Require().NoError(err)

	balance := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
	)
	requester := suite.NewAccountFromAddr(sdk.AccAddress("requester-----------"), balance)

	swapOutput := sdk.NewCoin("usdx", sdk.NewInt(5e6))
	swapMsg := types.NewMsgSwapForExactTokens(
		requester.GetAddress().String(),
		sdk.NewCoin("ukava", sdk.NewInt(1e6)),
		swapOutput,
		sdk.MustNewDecFromStr("0.01"),
		time.Now().Add(10*time.Minute).Unix(),
	)

	suite.Ctx = suite.App.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})
	res, err := suite.msgServer.SwapForExactTokens(sdk.WrapSDKContext(suite.Ctx), swapMsg)
	suite.Require().Equal(&types.MsgSwapForExactTokensResponse{}, res)
	suite.Require().NoError(err)

	expectedSwapInput := sdk.NewCoin("ukava", sdk.NewInt(1004015))

	suite.AccountBalanceEqual(requester.GetAddress(), balance.Sub(sdk.NewCoins(expectedSwapInput)).Add(swapOutput))
	suite.ModuleAccountBalanceEqual(reserves.Add(expectedSwapInput).Sub(sdk.NewCoins(swapOutput)))
	suite.PoolLiquidityEqual(reserves.Add(expectedSwapInput).Sub(sdk.NewCoins(swapOutput)))

	suite.EventsContains(suite.GetEvents(), sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		sdk.NewAttribute(sdk.AttributeKeySender, requester.GetAddress().String()),
	))

	suite.EventsContains(suite.GetEvents(), sdk.NewEvent(
		bank.EventTypeTransfer,
		sdk.NewAttribute(bank.AttributeKeyRecipient, swapModuleAccountAddress.String()),
		sdk.NewAttribute(bank.AttributeKeySender, requester.GetAddress().String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, expectedSwapInput.String()),
	))

	suite.EventsContains(suite.GetEvents(), sdk.NewEvent(
		bank.EventTypeTransfer,
		sdk.NewAttribute(bank.AttributeKeyRecipient, requester.GetAddress().String()),
		sdk.NewAttribute(bank.AttributeKeySender, swapModuleAccountAddress.String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, swapOutput.String()),
	))

	suite.EventsContains(suite.GetEvents(), sdk.NewEvent(
		types.EventTypeSwapTrade,
		sdk.NewAttribute(types.AttributeKeyPoolID, types.PoolID("ukava", "usdx")),
		sdk.NewAttribute(types.AttributeKeyRequester, requester.GetAddress().String()),
		sdk.NewAttribute(types.AttributeKeySwapInput, expectedSwapInput.String()),
		sdk.NewAttribute(types.AttributeKeySwapOutput, swapOutput.String()),
		sdk.NewAttribute(types.AttributeKeyFeePaid, "3013ukava"),
		sdk.NewAttribute(types.AttributeKeyExactDirection, "output"),
	))
}

func (suite *msgServerTestSuite) TestSwapForExactTokens_SlippageFailure() {
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(1000e6)),
		sdk.NewCoin("usdx", sdk.NewInt(5000e6)),
	)
	err := suite.CreatePool(reserves)
	suite.Require().NoError(err)

	balance := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
	)
	requester := suite.NewAccountFromAddr(sdk.AccAddress("requester-----------"), balance)

	swapOutput := sdk.NewCoin("usdx", sdk.NewInt(5e6))
	swapMsg := types.NewMsgSwapForExactTokens(
		requester.GetAddress().String(),
		sdk.NewCoin("ukava", sdk.NewInt(990991)),
		swapOutput,
		sdk.MustNewDecFromStr("0.01"),
		time.Now().Add(10*time.Minute).Unix(),
	)

	suite.Ctx = suite.App.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})
	res, err := suite.msgServer.SwapForExactTokens(sdk.WrapSDKContext(suite.Ctx), swapMsg)
	suite.Require().Nil(res)
	suite.EqualError(err, "slippage 0.010000979019022939 > limit 0.010000000000000000: slippage exceeded")
	suite.Nil(res)
}

func (suite *msgServerTestSuite) TestSwapForExactTokens_DeadlineExceeded() {
	balance := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
	)
	requester := suite.NewAccountFromAddr(sdk.AccAddress("requester-----------"), balance)

	swapMsg := types.NewMsgSwapForExactTokens(
		requester.GetAddress().String(),
		sdk.NewCoin("ukava", sdk.NewInt(5e6)),
		sdk.NewCoin("usdx", sdk.NewInt(25e5)),
		sdk.MustNewDecFromStr("0.01"),
		suite.Ctx.BlockTime().Add(-1*time.Second).Unix(),
	)

	res, err := suite.msgServer.SwapForExactTokens(sdk.WrapSDKContext(suite.Ctx), swapMsg)
	suite.Require().Nil(res)
	suite.EqualError(err, fmt.Sprintf("block time %d >= deadline %d: deadline exceeded", suite.Ctx.BlockTime().Unix(), swapMsg.GetDeadline().Unix()))
	suite.Nil(res)
}

func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(msgServerTestSuite))
}
