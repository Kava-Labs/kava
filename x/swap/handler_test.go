package swap_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/kava-labs/kava/x/swap"
	"github.com/kava-labs/kava/x/swap/testutil"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	tmtime "github.com/tendermint/tendermint/types/time"
)

var swapModuleAccountAddress = sdk.AccAddress(crypto.AddressHash([]byte(swap.ModuleAccountName)))

type handlerTestSuite struct {
	testutil.Suite
	handler sdk.Handler
}

func (suite *handlerTestSuite) SetupTest() {
	suite.Suite.SetupTest()
	suite.handler = swap.NewHandler(suite.Keeper)
}

func (suite *handlerTestSuite) TestDeposit_CreatePool() {
	pool := swap.NewAllowedPool("ukava", "usdx")
	suite.Require().NoError(pool.Validate())
	suite.Keeper.SetParams(suite.Ctx, swap.NewParams(swap.NewAllowedPools(pool), swap.DefaultSwapFee))

	balance := sdk.NewCoins(
		sdk.NewCoin(pool.TokenA, sdk.NewInt(10e6)),
		sdk.NewCoin(pool.TokenB, sdk.NewInt(50e6)),
	)
	depositor := suite.CreateAccount(balance)

	deposit := swap.NewMsgDeposit(
		depositor.GetAddress(),
		sdk.NewCoin(pool.TokenA, depositor.GetCoins().AmountOf(pool.TokenA)),
		sdk.NewCoin(pool.TokenB, depositor.GetCoins().AmountOf(pool.TokenB)),
		sdk.MustNewDecFromStr("0.01"),
		time.Now().Add(10*time.Minute).Unix(),
	)

	res, err := suite.handler(suite.Ctx, deposit)
	suite.Require().NoError(err)

	suite.AccountBalanceEqual(depositor, sdk.Coins(nil))
	suite.ModuleAccountBalanceEqual(balance)
	suite.PoolLiquidityEqual(balance)
	suite.PoolShareValueEqual(depositor, pool, balance)

	suite.EventsContains(res.Events, sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, swap.AttributeValueCategory),
		sdk.NewAttribute(sdk.AttributeKeySender, depositor.GetAddress().String()),
	))

	suite.EventsContains(res.Events, sdk.NewEvent(
		bank.EventTypeTransfer,
		sdk.NewAttribute(bank.AttributeKeyRecipient, swapModuleAccountAddress.String()),
		sdk.NewAttribute(bank.AttributeKeySender, depositor.GetAddress().String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, balance.String()),
	))

	suite.EventsContains(res.Events, sdk.NewEvent(
		swap.EventTypeSwapDeposit,
		sdk.NewAttribute(swap.AttributeKeyPoolID, swap.PoolID(pool.TokenA, pool.TokenB)),
		sdk.NewAttribute(swap.AttributeKeyDepositor, depositor.GetAddress().String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, balance.String()),
		sdk.NewAttribute(swap.AttributeKeyShares, "22360679"),
	))
}

func (suite *handlerTestSuite) TestDeposit_DeadlineExceeded() {
	pool := swap.NewAllowedPool("ukava", "usdx")
	suite.Require().NoError(pool.Validate())
	suite.Keeper.SetParams(suite.Ctx, swap.NewParams(swap.NewAllowedPools(pool), swap.DefaultSwapFee))

	balance := sdk.NewCoins(
		sdk.NewCoin(pool.TokenA, sdk.NewInt(10e6)),
		sdk.NewCoin(pool.TokenB, sdk.NewInt(50e6)),
	)
	depositor := suite.CreateAccount(balance)

	deposit := swap.NewMsgDeposit(
		depositor.GetAddress(),
		sdk.NewCoin(pool.TokenA, depositor.GetCoins().AmountOf(pool.TokenA)),
		sdk.NewCoin(pool.TokenB, depositor.GetCoins().AmountOf(pool.TokenB)),
		sdk.MustNewDecFromStr("0.01"),
		suite.Ctx.BlockTime().Add(-1*time.Second).Unix(),
	)

	res, err := suite.handler(suite.Ctx, deposit)
	suite.EqualError(err, fmt.Sprintf("deadline exceeded: block time %d >= deadline %d", suite.Ctx.BlockTime().Unix(), deposit.GetDeadline().Unix()))
	suite.Nil(res)
}

func (suite *handlerTestSuite) TestDeposit_ExistingPool() {
	pool := swap.NewAllowedPool("ukava", "usdx")
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
	depositor := suite.NewAccountFromAddr(sdk.AccAddress("new depositor"), balance)

	deposit := swap.NewMsgDeposit(
		depositor.GetAddress(),
		sdk.NewCoin("usdx", depositor.GetCoins().AmountOf("usdx")),
		sdk.NewCoin("ukava", depositor.GetCoins().AmountOf("ukava")),
		sdk.MustNewDecFromStr("0.01"),
		time.Now().Add(10*time.Minute).Unix(),
	)

	res, err := suite.handler(suite.Ctx, deposit)
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

	suite.EventsContains(res.Events, sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, swap.AttributeValueCategory),
		sdk.NewAttribute(sdk.AttributeKeySender, depositor.GetAddress().String()),
	))

	suite.EventsContains(res.Events, sdk.NewEvent(
		bank.EventTypeTransfer,
		sdk.NewAttribute(bank.AttributeKeyRecipient, swapModuleAccountAddress.String()),
		sdk.NewAttribute(bank.AttributeKeySender, depositor.GetAddress().String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, expectedDeposit.String()),
	))

	suite.EventsContains(res.Events, sdk.NewEvent(
		swap.EventTypeSwapDeposit,
		sdk.NewAttribute(swap.AttributeKeyPoolID, swap.PoolID(pool.TokenA, pool.TokenB)),
		sdk.NewAttribute(swap.AttributeKeyDepositor, depositor.GetAddress().String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, expectedDeposit.String()),
		sdk.NewAttribute(swap.AttributeKeyShares, "2236067"),
	))
}

func (suite *handlerTestSuite) TestDeposit_ExistingPool_SlippageFailure() {
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

	deposit := swap.NewMsgDeposit(
		depositor.GetAddress(),
		sdk.NewCoin("usdx", depositor.GetCoins().AmountOf("usdx")),
		sdk.NewCoin("ukava", depositor.GetCoins().AmountOf("ukava")),
		sdk.MustNewDecFromStr("0.01"),
		time.Now().Add(10*time.Minute).Unix(),
	)

	res, err := suite.handler(suite.Ctx, deposit)
	suite.EqualError(err, "slippage exceeded: slippage 4.000000000000000000 > limit 0.010000000000000000")
	suite.Nil(res)
}

func (suite *handlerTestSuite) TestWithdraw_AllShares() {
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
		sdk.NewCoin("usdx", sdk.NewInt(50e6)),
	)
	depositor := suite.CreateAccount(reserves)
	pool := swap.NewAllowedPool(reserves[0].Denom, reserves[1].Denom)
	suite.Require().NoError(pool.Validate())
	suite.Keeper.SetParams(suite.Ctx, swap.NewParams(swap.NewAllowedPools(pool), swap.DefaultSwapFee))

	err := suite.Keeper.Deposit(suite.Ctx, depositor.GetAddress(), reserves[0], reserves[1], sdk.MustNewDecFromStr("1"))
	suite.Require().NoError(err)

	withdraw := swap.NewMsgWithdraw(
		depositor.GetAddress(),
		sdk.NewInt(22360679),
		reserves[0],
		reserves[1],
		time.Now().Add(10*time.Minute).Unix(),
	)

	ctx := suite.App.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	res, err := suite.handler(ctx, withdraw)
	suite.Require().NoError(err)

	suite.AccountBalanceEqual(depositor, reserves)
	suite.ModuleAccountBalanceEqual(sdk.Coins(nil))
	suite.PoolDeleted("ukava", "usdx")
	suite.PoolSharesDeleted(depositor.GetAddress(), "ukava", "usdx")

	suite.EventsContains(res.Events, sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, swap.AttributeValueCategory),
		sdk.NewAttribute(sdk.AttributeKeySender, depositor.GetAddress().String()),
	))

	suite.EventsContains(res.Events, sdk.NewEvent(
		bank.EventTypeTransfer,
		sdk.NewAttribute(bank.AttributeKeyRecipient, depositor.GetAddress().String()),
		sdk.NewAttribute(bank.AttributeKeySender, swapModuleAccountAddress.String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, reserves.String()),
	))

	suite.EventsContains(res.Events, sdk.NewEvent(
		swap.EventTypeSwapWithdraw,
		sdk.NewAttribute(swap.AttributeKeyPoolID, swap.PoolID(pool.TokenA, pool.TokenB)),
		sdk.NewAttribute(swap.AttributeKeyOwner, depositor.GetAddress().String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, reserves.String()),
		sdk.NewAttribute(swap.AttributeKeyShares, "22360679"),
	))
}

func (suite *handlerTestSuite) TestWithdraw_PartialShares() {
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
		sdk.NewCoin("usdx", sdk.NewInt(50e6)),
	)
	depositor := suite.CreateAccount(reserves)
	pool := swap.NewAllowedPool(reserves[0].Denom, reserves[1].Denom)
	suite.Require().NoError(pool.Validate())
	suite.Keeper.SetParams(suite.Ctx, swap.NewParams(swap.NewAllowedPools(pool), swap.DefaultSwapFee))

	err := suite.Keeper.Deposit(suite.Ctx, depositor.GetAddress(), reserves[0], reserves[1], sdk.MustNewDecFromStr("1"))
	suite.Require().NoError(err)

	minTokenA := sdk.NewCoin("ukava", sdk.NewInt(4999999))
	minTokenB := sdk.NewCoin("usdx", sdk.NewInt(24999998))

	withdraw := swap.NewMsgWithdraw(
		depositor.GetAddress(),
		sdk.NewInt(11180339),
		minTokenA,
		minTokenB,
		time.Now().Add(10*time.Minute).Unix(),
	)

	ctx := suite.App.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	res, err := suite.handler(ctx, withdraw)
	suite.Require().NoError(err)

	expectedCoinsReceived := sdk.NewCoins(minTokenA, minTokenB)

	suite.AccountBalanceEqual(depositor, expectedCoinsReceived)
	suite.ModuleAccountBalanceEqual(reserves.Sub(expectedCoinsReceived))
	suite.PoolLiquidityEqual(reserves.Sub(expectedCoinsReceived))
	suite.PoolShareValueEqual(depositor, swap.NewAllowedPool("ukava", "usdx"), reserves.Sub(expectedCoinsReceived))

	suite.EventsContains(res.Events, sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, swap.AttributeValueCategory),
		sdk.NewAttribute(sdk.AttributeKeySender, depositor.GetAddress().String()),
	))

	suite.EventsContains(res.Events, sdk.NewEvent(
		bank.EventTypeTransfer,
		sdk.NewAttribute(bank.AttributeKeyRecipient, depositor.GetAddress().String()),
		sdk.NewAttribute(bank.AttributeKeySender, swapModuleAccountAddress.String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, expectedCoinsReceived.String()),
	))

	suite.EventsContains(res.Events, sdk.NewEvent(
		swap.EventTypeSwapWithdraw,
		sdk.NewAttribute(swap.AttributeKeyPoolID, swap.PoolID(pool.TokenA, pool.TokenB)),
		sdk.NewAttribute(swap.AttributeKeyOwner, depositor.GetAddress().String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, expectedCoinsReceived.String()),
		sdk.NewAttribute(swap.AttributeKeyShares, "11180339"),
	))
}

func (suite *handlerTestSuite) TestWithdraw_SlippageFailure() {
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
		sdk.NewCoin("usdx", sdk.NewInt(50e6)),
	)
	depositor := suite.CreateAccount(reserves)
	pool := swap.NewAllowedPool(reserves[0].Denom, reserves[1].Denom)
	suite.Require().NoError(pool.Validate())
	suite.Keeper.SetParams(suite.Ctx, swap.NewParams(swap.NewAllowedPools(pool), swap.DefaultSwapFee))

	err := suite.Keeper.Deposit(suite.Ctx, depositor.GetAddress(), reserves[0], reserves[1], sdk.MustNewDecFromStr("1"))
	suite.Require().NoError(err)

	minTokenA := sdk.NewCoin("ukava", sdk.NewInt(5e6))
	minTokenB := sdk.NewCoin("usdx", sdk.NewInt(25e6))

	withdraw := swap.NewMsgWithdraw(
		depositor.GetAddress(),
		sdk.NewInt(11180339),
		minTokenA,
		minTokenB,
		time.Now().Add(10*time.Minute).Unix(),
	)

	res, err := suite.handler(suite.Ctx, withdraw)
	suite.EqualError(err, "slippage exceeded: minimum withdraw not met")
	suite.Nil(res)
}

func (suite *handlerTestSuite) TestWithdraw_DeadlineExceeded() {
	balance := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(10e6)),
		sdk.NewCoin("usdx", sdk.NewInt(50e6)),
	)
	from := suite.CreateAccount(balance)

	withdraw := swap.NewMsgWithdraw(
		from.GetAddress(),
		sdk.NewInt(2e6),
		sdk.NewCoin("ukava", sdk.NewInt(1e6)),
		sdk.NewCoin("usdx", sdk.NewInt(5e6)),
		suite.Ctx.BlockTime().Add(-1*time.Second).Unix(),
	)

	res, err := suite.handler(suite.Ctx, withdraw)
	suite.EqualError(err, fmt.Sprintf("deadline exceeded: block time %d >= deadline %d", suite.Ctx.BlockTime().Unix(), withdraw.GetDeadline().Unix()))
	suite.Nil(res)
}

func (suite *handlerTestSuite) TestInvalidMsg() {
	res, err := suite.handler(suite.Ctx, sdk.NewTestMsg())
	suite.Nil(res)
	suite.EqualError(err, "unknown request: unrecognized swap message type: *types.TestMsg")
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(handlerTestSuite))
}
