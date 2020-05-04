package cdp_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp"

	"github.com/stretchr/testify/suite"
)

type HandlerTestSuite struct {
	suite.Suite

	ctx     sdk.Context
	app     app.TestApp
	handler sdk.Handler
	keeper  cdp.Keeper
}

func (suite *HandlerTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	tApp.InitializeFromGenesisStates(
		NewPricefeedGenStateMulti(),
		NewCDPGenStateMulti(),
	)
	keeper := tApp.GetCDPKeeper()
	suite.handler = cdp.NewHandler(keeper)
	suite.app = tApp
	suite.keeper = keeper
	suite.ctx = ctx
}

func (suite *HandlerTestSuite) TestMsgCreateCdp() {
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	ak := suite.app.GetAccountKeeper()
	acc := ak.NewAccountWithAddress(suite.ctx, addrs[0])
	acc.SetCoins(cs(c("xrp", 200000000), c("btc", 500000000)))
	ak.SetAccount(suite.ctx, acc)
	msg := cdp.NewMsgCreateCDP(
		addrs[0],
		c("xrp", 200000000),
		c("usdx", 10000000),
	)
	res, err := suite.handler(suite.ctx, msg)
	suite.Require().NoError(err)
	suite.Require().Equal(cdp.GetCdpIDBytes(uint64(1)), res.Data)

}

func (suite *HandlerTestSuite) TestInvalidMsg() {
	res, err := suite.handler(suite.ctx, sdk.NewTestMsg())
	suite.Require().Error(err)
	suite.Require().Nil(res)
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
