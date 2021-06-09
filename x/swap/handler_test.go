package swap_test

import (
	"testing"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/swap"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

type handlerTestSuite struct {
	suite.Suite
	keeper  swap.Keeper
	handler sdk.Handler
	app     app.TestApp
	ctx     sdk.Context
}

func (suite *handlerTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	keeper := tApp.GetSwapKeeper()

	suite.ctx = ctx
	suite.app = tApp
	suite.keeper = keeper
	suite.handler = swap.NewHandler(keeper)
}

func (suite *handlerTestSuite) TestDeposit_CreatePool() {
	pool := swap.NewAllowedPool("ukava", "usdx")
	suite.Require().NoError(pool.Validate())
	suite.keeper.SetParams(suite.ctx, swap.NewParams(swap.NewAllowedPools(pool), swap.DefaultSwapFee))

	balance := sdk.NewCoins(
		sdk.NewCoin(pool.TokenA, sdk.NewInt(10e6)),
		sdk.NewCoin(pool.TokenB, sdk.NewInt(50e6)),
	)
	depositor := suite.GetAccount(balance)

	deposit := swap.NewMsgDeposit(
		depositor.GetAddress(),
		sdk.NewCoin(pool.TokenA, depositor.GetCoins().AmountOf(pool.TokenA)),
		sdk.NewCoin(pool.TokenB, depositor.GetCoins().AmountOf(pool.TokenB)),
	)

	res, err := suite.handler(suite.ctx, deposit)
	suite.Require().NoError(err)

	suite.AccountBalanceEqual(depositor, sdk.Coins{})
	suite.ModuleAccountBalanceEqual(balance)
	suite.PoolLiquidtyEqual(pool, balance)
	suite.PoolShareValueEqual(depositor, pool, balance)

	suite.EventsContains(sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, swap.AttributeValueCategory),
		sdk.NewAttribute(sdk.AttributeKeySender, depositor.String()),
	))

	suite.EventsContains(sdk.NewEvent(
		swap.EventTypeSwapDeposit,
		sdk.NewAttribute(swap.AttributeKeyPoolName, pool.Name()),
		sdk.NewAttribute(swap.AttributeKeyDepositor, depositor.GetAddress().String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, balance.String()),
	))
}

func (suite *handlerTestSuite) GetAccount(initialBalance sdk.Coins) authexported.Account {
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	ak := suite.app.GetAccountKeeper()

	acc := ak.NewAccountWithAddress(suite.ctx, addrs[0])
	acc.SetCoins(initialBalance)

	ak.SetAccount(suite.ctx, acc)
	return acc
}

func (suite *handlerTestSuite) TestInvalidMsg() {
	res, err := suite.handler(suite.ctx, sdk.NewTestMsg())
	suite.Nil(res)
	suite.EqualError(err, "unknown request: unrecognized swap message type: *types.TestMsg")
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(handlerTestSuite))
}
