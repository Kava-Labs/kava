package keeper_test

import (
	"errors"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp/keeper"
	"github.com/kava-labs/kava/x/cdp/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

type DrawTestSuite struct {
	suite.Suite

	keeper keeper.Keeper
	app    app.TestApp
	ctx    sdk.Context
	addrs  []sdk.AccAddress
}

func (suite *DrawTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	_, addrs := app.GeneratePrivKeyAddressPairs(3)
	authGS := app.NewAuthGenState(
		addrs,
		[]sdk.Coins{
			cs(c("xrp", 500000000), c("btc", 500000000), c("usdx", 10000000000)),
			cs(c("xrp", 200000000)),
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
	err := suite.keeper.AddCdp(suite.ctx, addrs[0], c("xrp", 400000000), c("usdx", 10000000))
	suite.NoError(err)
}

func (suite *DrawTestSuite) TestAddRepayPrincipal() {

	err := suite.keeper.AddPrincipal(suite.ctx, suite.addrs[0], "xrp", c("usdx", 10000000))
	suite.NoError(err)

	t, found := suite.keeper.GetCDP(suite.ctx, "xrp", uint64(1))
	suite.True(found)
	suite.Equal(c("usdx", 20000000), t.Principal)
	ctd := suite.keeper.CalculateCollateralToDebtRatio(suite.ctx, t.Collateral, t.Principal.Add(t.AccumulatedFees))
	suite.Equal(d("20.0"), ctd)
	ts := suite.keeper.GetAllCdpsByDenomAndRatio(suite.ctx, "xrp", d("20.0"))
	suite.Equal(0, len(ts))
	ts = suite.keeper.GetAllCdpsByDenomAndRatio(suite.ctx, "xrp", d("20.0").Add(sdk.SmallestDec()))
	suite.Equal(ts[0], t)
	tp := suite.keeper.GetTotalPrincipal(suite.ctx, "xrp", "usdx")
	suite.Equal(i(20000000), tp)
	sk := suite.app.GetSupplyKeeper()
	acc := sk.GetModuleAccount(suite.ctx, types.ModuleName)
	suite.Equal(cs(c("xrp", 400000000), c("debt", 20000000)), acc.GetCoins())

	err = suite.keeper.AddPrincipal(suite.ctx, suite.addrs[0], "xrp", c("susd", 10000000))
	suite.Require().True(errors.Is(err, types.ErrInvalidDebtRequest))

	err = suite.keeper.AddPrincipal(suite.ctx, suite.addrs[1], "xrp", c("usdx", 10000000))
	suite.Require().True(errors.Is(err, types.ErrCdpNotFound))
	err = suite.keeper.AddPrincipal(suite.ctx, suite.addrs[0], "xrp", c("xusd", 10000000))
	suite.Require().True(errors.Is(err, types.ErrInvalidDebtRequest))
	err = suite.keeper.AddPrincipal(suite.ctx, suite.addrs[0], "xrp", c("usdx", 311000000))
	suite.Require().True(errors.Is(err, types.ErrInvalidCollateralRatio))

	err = suite.keeper.RepayPrincipal(suite.ctx, suite.addrs[0], "xrp", c("usdx", 10000000))
	suite.NoError(err)

	t, found = suite.keeper.GetCDP(suite.ctx, "xrp", uint64(1))
	suite.True(found)
	suite.Equal(c("usdx", 10000000), t.Principal)

	ctd = suite.keeper.CalculateCollateralToDebtRatio(suite.ctx, t.Collateral, t.Principal.Add(t.AccumulatedFees))
	suite.Equal(d("40.0"), ctd)
	ts = suite.keeper.GetAllCdpsByDenomAndRatio(suite.ctx, "xrp", d("40.0"))
	suite.Equal(0, len(ts))
	ts = suite.keeper.GetAllCdpsByDenomAndRatio(suite.ctx, "xrp", d("40.0").Add(sdk.SmallestDec()))
	suite.Equal(ts[0], t)
	sk = suite.app.GetSupplyKeeper()
	acc = sk.GetModuleAccount(suite.ctx, types.ModuleName)
	suite.Equal(cs(c("xrp", 400000000), c("debt", 10000000)), acc.GetCoins())

	err = suite.keeper.RepayPrincipal(suite.ctx, suite.addrs[0], "xrp", c("xusd", 10000000))
	suite.Require().True(errors.Is(err, types.ErrInvalidPayment))
	err = suite.keeper.RepayPrincipal(suite.ctx, suite.addrs[1], "xrp", c("xusd", 10000000))
	suite.Require().True(errors.Is(err, types.ErrCdpNotFound))

	err = suite.keeper.RepayPrincipal(suite.ctx, suite.addrs[0], "xrp", c("usdx", 9000000))
	suite.Require().True(errors.Is(err, types.ErrBelowDebtFloor))
	err = suite.keeper.RepayPrincipal(suite.ctx, suite.addrs[0], "xrp", c("usdx", 10000000))
	suite.NoError(err)

	_, found = suite.keeper.GetCDP(suite.ctx, "xrp", uint64(1))
	suite.False(found)
	ts = suite.keeper.GetAllCdpsByDenomAndRatio(suite.ctx, "xrp", types.MaxSortableDec)
	suite.Equal(0, len(ts))
	ts = suite.keeper.GetAllCdpsByDenom(suite.ctx, "xrp")
	suite.Equal(0, len(ts))
	sk = suite.app.GetSupplyKeeper()
	acc = sk.GetModuleAccount(suite.ctx, types.ModuleName)
	suite.Equal(sdk.Coins(nil), acc.GetCoins())

}

func (suite *DrawTestSuite) TestRepayPrincipalOverpay() {
	err := suite.keeper.RepayPrincipal(suite.ctx, suite.addrs[0], "xrp", c("usdx", 20000000))
	suite.NoError(err)
	ak := suite.app.GetAccountKeeper()
	acc := ak.GetAccount(suite.ctx, suite.addrs[0])
	suite.Equal(i(10000000000), (acc.GetCoins().AmountOf("usdx")))
	_, found := suite.keeper.GetCDP(suite.ctx, "xrp", 1)
	suite.False(found)
}

func (suite *DrawTestSuite) TestAddRepayPrincipalFees() {
	err := suite.keeper.AddCdp(suite.ctx, suite.addrs[2], c("xrp", 1000000000000), c("usdx", 100000000000))
	suite.NoError(err)
	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Minute * 10))
	err = suite.keeper.UpdateFeesForAllCdps(suite.ctx, "xrp")
	suite.NoError(err)
	err = suite.keeper.AddPrincipal(suite.ctx, suite.addrs[2], "xrp", c("usdx", 10000000))
	suite.NoError(err)
	t, _ := suite.keeper.GetCDP(suite.ctx, "xrp", uint64(2))
	suite.Equal(c("usdx", 92827), t.AccumulatedFees)
	err = suite.keeper.RepayPrincipal(suite.ctx, suite.addrs[2], "xrp", c("usdx", 100))
	suite.NoError(err)
	t, _ = suite.keeper.GetCDP(suite.ctx, "xrp", uint64(2))
	suite.Equal(c("usdx", 92727), t.AccumulatedFees)
	err = suite.keeper.RepayPrincipal(suite.ctx, suite.addrs[2], "xrp", c("usdx", 100010092727))
	suite.NoError(err)
	_, f := suite.keeper.GetCDP(suite.ctx, "xrp", uint64(2))
	suite.False(f)

	err = suite.keeper.AddCdp(suite.ctx, suite.addrs[2], c("xrp", 1000000000000), c("usdx", 100000000))
	suite.NoError(err)

	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Second * 31536000)) // move forward one year in time
	err = suite.keeper.UpdateFeesForAllCdps(suite.ctx, "xrp")
	suite.NoError(err)
	err = suite.keeper.AddPrincipal(suite.ctx, suite.addrs[2], "xrp", c("usdx", 100000000))
	suite.NoError(err)
	t, _ = suite.keeper.GetCDP(suite.ctx, "xrp", uint64(3))
	suite.Equal(c("usdx", 5000000), t.AccumulatedFees)
}

func (suite *DrawTestSuite) TestPricefeedFailure() {
	ctx := suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Hour * 2))
	pfk := suite.app.GetPriceFeedKeeper()
	pfk.SetCurrentPrices(ctx, "xrp:usd")
	err := suite.keeper.AddPrincipal(ctx, suite.addrs[0], "xrp", c("usdx", 10000000))
	suite.Error(err)
	err = suite.keeper.RepayPrincipal(ctx, suite.addrs[0], "xrp", c("usdx", 10000000))
	suite.NoError(err)
}

func (suite *DrawTestSuite) TestModuleAccountFailure() {
	suite.Panics(func() {
		ctx := suite.ctx.WithBlockHeader(suite.ctx.BlockHeader())
		sk := suite.app.GetSupplyKeeper()
		acc := sk.GetModuleAccount(ctx, types.ModuleName)
		ak := suite.app.GetAccountKeeper()
		ak.RemoveAccount(ctx, acc)
		suite.keeper.RepayPrincipal(ctx, suite.addrs[0], "xrp", c("usdx", 10000000))
	})
}

func TestDrawTestSuite(t *testing.T) {
	suite.Run(t, new(DrawTestSuite))
}
