package cdp_test

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

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
		cs(c("xrp", 200000000)),
		cs(c("usdx", 10000000)),
	)
	res := suite.handler(suite.ctx, msg)
	suite.True(res.IsOK())
	suite.Equal(cdp.GetCdpIDBytes(uint64(1)), res.Data)

}

func (suite *HandlerTestSuite) TestInvalidMsg() {
	res := suite.handler(suite.ctx, sdk.NewTestMsg())
	suite.False(res.IsOK())
	suite.True(strings.Contains(res.Log, "unrecognized cdp msg type"))
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
