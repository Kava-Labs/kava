package swap_test

import (
	"testing"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/swap"
	"github.com/kava-labs/kava/x/swap/types"

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
	pool := types.NewAllowedPool("ukava", "usdx")
	suite.Require().NoError(pool.Validate())
	suite.keeper.SetParams(suite.ctx, types.NewParams(types.NewAllowedPools(pool), types.DefaultSwapFee))

	balance := sdk.NewCoins(
		sdk.NewCoin(pool.TokenA, sdk.NewInt(10e6)),
		sdk.NewCoin(pool.TokenB, sdk.NewInt(50e6)),
	)
	depositor := suite.GetAccount(balance)

	deposit := swap.NewMsgDeposit{
		depositor: depositor.GetAddress(),
		depositor.GetCoins().AmountOf(pool.TokenA),
		depositor.GetCoins().AmountOf(pool.TokenB),
	}

	res, err := suite.handler(suite.ctx, deposit)
	suite.Require().NoError(err)

	suite.AccountBalanceEqual(depositor, sdk.Coins{})
	suite.ModuleAccountBalanceEqual(balance)
	suite.PoolLiquidtyEqual(pool, balance)
	suite.PoolShareValueEqual(depositor, pool, balance)

	suite.EventsContains(sdk.NewEvent(
		types.EventTypeSwapPoolDeposit,
		sdk.NewAttribute(types.AttributeKeyPoolName, pool.Name()),
		sdk.NewAttribute(types.AttributeKeyDepositor, depositor.GetAddress.String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, balance.String()),
	))
}

func (suite *handlerTestSuite) TestInvalidMsg() {
	res, err := suite.handler(suite.ctx, sdk.NewTestMsg())
	suite.Nil(res)
	suite.EqualError(err, "unknown request: unrecognized swap message type: *types.TestMsg")
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(handlerTestSuite))
}
