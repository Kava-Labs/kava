package cdp_test

import (
	"testing"

	testdata "github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp"
	"github.com/kava-labs/kava/x/cdp/keeper"
	"github.com/kava-labs/kava/x/cdp/types"

	"github.com/stretchr/testify/suite"
)

type HandlerTestSuite struct {
	suite.Suite

	ctx     sdk.Context
	app     app.TestApp
	handler sdk.Handler
	keeper  keeper.Keeper
}

func (suite *HandlerTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})
	tApp.InitializeFromGenesisStates(
		NewPricefeedGenStateMulti(tApp.AppCodec()),
		NewCDPGenStateMulti(tApp.AppCodec()),
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
	suite.app.FundAccount(suite.ctx, acc.GetAddress(), cs(c("xrp", 200000000), c("btc", 500000000)))
	ak.SetAccount(suite.ctx, acc)
	msg := types.NewMsgCreateCDP(
		addrs[0],
		c("xrp", 200000000),
		c("usdx", 10000000),
		"xrp-a",
	)
	res, err := suite.handler(suite.ctx, &msg)
	suite.Require().NoError(err)
	suite.Require().Equal(types.GetCdpIDBytes(uint64(1)), res.Data)

}

func (suite *HandlerTestSuite) TestInvalidMsg() {
	res, err := suite.handler(suite.ctx, testdata.NewTestMsg())
	suite.Require().Error(err)
	suite.Require().Nil(res)
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
