package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp/keeper"
	"github.com/kava-labs/kava/x/cdp/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

type SeizeTestSuite struct {
	suite.Suite

	keeper keeper.Keeper
	app    app.TestApp
	ctx    sdk.Context
	addrs  []sdk.AccAddress
}

func (suite *SeizeTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	_, addrs := app.GeneratePrivKeyAddressPairs(3)
	authGS := app.NewAuthGenState(
		addrs,
		[]sdk.Coins{
			cs(c("xrp", 500), c("btc", 5)),
			cs(c("xrp", 200)),
			cs(c("xrp", 10000000000000), c("usdx", 100000000000))})
	tApp.InitializeFromGenesisStates(
		authGS,
		NewPricefeedGenStateMulti(),
		NewCDPGenStateMulti(),
	)
	keeper := tApp.GetCDPKeeper()
	suite.app = tApp
	suite.keeper = keeper
	suite.ctx = ctx
	suite.addrs = addrs
	err := suite.keeper.AddCdp(suite.ctx, addrs[0], cs(c("xrp", 400)), cs(c("usdx", 10)))
	suite.NoError(err)
	err = suite.keeper.DepositCollateral(suite.ctx, suite.addrs[0], suite.addrs[1], cs(c("xrp", 10)))
	suite.NoError(err)
}

func (suite *SeizeTestSuite) TestSeizeCollateral() {
	sk := suite.app.GetSupplyKeeper()
	cdp, _ := suite.keeper.GetCDP(suite.ctx, "xrp", uint64(1))
	suite.keeper.SeizeCollateral(suite.ctx, cdp)
	cdpModAcc := sk.GetModuleAccount(suite.ctx, "cdp")
	suite.Equal(sdk.Coins(nil), cdpModAcc.GetCoins())
	liqModAcc := sk.GetModuleAccount(suite.ctx, "liquidator")
	suite.Equal(cs(c("debt", 10), c("xrp", 410)), liqModAcc.GetCoins())
	ak := suite.app.GetAccountKeeper()
	acc := ak.GetAccount(suite.ctx, suite.addrs[0])
	suite.Equal(i(10), acc.GetCoins().AmountOf("usdx"))
	err := suite.keeper.WithdrawCollateral(suite.ctx, suite.addrs[0], suite.addrs[0], cs(c("xrp", 10)))
	suite.Equal(types.CodeDepositNotAvailable, err.Result().Code)
}

func TestSeizeTestSuite(t *testing.T) {
	suite.Run(t, new(SeizeTestSuite))
}
