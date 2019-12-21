package keeper_test

import (
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
}

func (suite *DrawTestSuite) TestAddRepayPrincipal() {

	err := suite.keeper.AddPrincipal(suite.ctx, suite.addrs[0], "xrp", cs(c("usdx", 10)))
	suite.NoError(err)

	t, _ := suite.keeper.GetCDP(suite.ctx, "xrp", uint64(1))
	suite.Equal(cs(c("usdx", 20)), t.Principal)
	ctd := suite.keeper.CalculateCollateralToDebtRatio(suite.ctx, t.Collateral, t.Principal.Add(t.AccumulatedFees))
	suite.Equal(d("20.0"), ctd)
	ts := suite.keeper.GetAllCdpsByDenomAndRatio(suite.ctx, "xrp", d("20.0"))
	suite.Equal(0, len(ts))
	ts = suite.keeper.GetAllCdpsByDenomAndRatio(suite.ctx, "xrp", d("20.0").Add(sdk.SmallestDec()))
	suite.Equal(ts[0], t)
	tp := suite.keeper.GetTotalPrincipal(suite.ctx, "xrp", "usdx")
	suite.Equal(i(20), tp)
	sk := suite.app.GetSupplyKeeper()
	acc := sk.GetModuleAccount(suite.ctx, types.ModuleName)
	suite.Equal(cs(c("xrp", 400), c("debt", 20)), acc.GetCoins())

	err = suite.keeper.AddPrincipal(suite.ctx, suite.addrs[0], "xrp", cs(c("susd", 10)))
	suite.NoError(err)
	t, _ = suite.keeper.GetCDP(suite.ctx, "xrp", uint64(1))
	suite.Equal(cs(c("usdx", 20), c("susd", 10)), t.Principal)
	ctd = suite.keeper.CalculateCollateralToDebtRatio(suite.ctx, t.Collateral, t.Principal.Add(t.AccumulatedFees))
	suite.Equal(d("400").Quo(d("30")), ctd)
	ts = suite.keeper.GetAllCdpsByDenomAndRatio(suite.ctx, "xrp", d("400").Quo(d("30")))
	suite.Equal(0, len(ts))
	ts = suite.keeper.GetAllCdpsByDenomAndRatio(suite.ctx, "xrp", d("400").Quo(d("30")).Add(sdk.SmallestDec()))
	suite.Equal(ts[0], t)
	tp = suite.keeper.GetTotalPrincipal(suite.ctx, "xrp", "susd")
	suite.Equal(i(10), tp)
	sk = suite.app.GetSupplyKeeper()
	acc = sk.GetModuleAccount(suite.ctx, types.ModuleName)
	suite.Equal(cs(c("xrp", 400), c("debt", 30)), acc.GetCoins())

	err = suite.keeper.AddPrincipal(suite.ctx, suite.addrs[1], "xrp", cs(c("usdx", 10)))
	suite.Equal(types.CodeCdpNotFound, err.Result().Code)
	err = suite.keeper.AddPrincipal(suite.ctx, suite.addrs[0], "xrp", cs(c("xusd", 10)))
	suite.Equal(types.CodeDebtNotSupported, err.Result().Code)
	err = suite.keeper.AddPrincipal(suite.ctx, suite.addrs[0], "xrp", cs(c("usdx", 311)))
	suite.Equal(types.CodeInvalidCollateralRatio, err.Result().Code)

	err = suite.keeper.RepayPrincipal(suite.ctx, suite.addrs[0], "xrp", cs(c("usdx", 10)))
	suite.NoError(err)

	t, _ = suite.keeper.GetCDP(suite.ctx, "xrp", uint64(1))
	suite.Equal(cs(c("usdx", 10), c("susd", 10)), t.Principal)
	ctd = suite.keeper.CalculateCollateralToDebtRatio(suite.ctx, t.Collateral, t.Principal.Add(t.AccumulatedFees))
	suite.Equal(d("20.0"), ctd)
	ts = suite.keeper.GetAllCdpsByDenomAndRatio(suite.ctx, "xrp", d("20.0"))
	suite.Equal(0, len(ts))
	ts = suite.keeper.GetAllCdpsByDenomAndRatio(suite.ctx, "xrp", d("20.0").Add(sdk.SmallestDec()))
	suite.Equal(ts[0], t)
	tp = suite.keeper.GetTotalPrincipal(suite.ctx, "xrp", "usdx")
	suite.Equal(i(10), tp)
	sk = suite.app.GetSupplyKeeper()
	acc = sk.GetModuleAccount(suite.ctx, types.ModuleName)
	suite.Equal(cs(c("xrp", 400), c("debt", 20)), acc.GetCoins())

	err = suite.keeper.RepayPrincipal(suite.ctx, suite.addrs[0], "xrp", cs(c("susd", 10)))
	suite.NoError(err)

	t, _ = suite.keeper.GetCDP(suite.ctx, "xrp", uint64(1))
	suite.Equal(cs(c("usdx", 10)), t.Principal)
	ctd = suite.keeper.CalculateCollateralToDebtRatio(suite.ctx, t.Collateral, t.Principal.Add(t.AccumulatedFees))
	suite.Equal(d("40.0"), ctd)
	ts = suite.keeper.GetAllCdpsByDenomAndRatio(suite.ctx, "xrp", d("40.0"))
	suite.Equal(0, len(ts))
	ts = suite.keeper.GetAllCdpsByDenomAndRatio(suite.ctx, "xrp", d("40.0").Add(sdk.SmallestDec()))
	suite.Equal(ts[0], t)
	tp = suite.keeper.GetTotalPrincipal(suite.ctx, "xrp", "susd")
	suite.Equal(i(0), tp)
	sk = suite.app.GetSupplyKeeper()
	acc = sk.GetModuleAccount(suite.ctx, types.ModuleName)
	suite.Equal(cs(c("xrp", 400), c("debt", 10)), acc.GetCoins())

	err = suite.keeper.RepayPrincipal(suite.ctx, suite.addrs[0], "xrp", cs(c("xusd", 10)))
	suite.Equal(types.CodeInvalidPaymentDenom, err.Result().Code)
	err = suite.keeper.RepayPrincipal(suite.ctx, suite.addrs[1], "xrp", cs(c("xusd", 10)))
	suite.Equal(types.CodeCdpNotFound, err.Result().Code)
	err = suite.keeper.RepayPrincipal(suite.ctx, suite.addrs[0], "xrp", cs(c("usdx", 100)))
	suite.Error(err)

	err = suite.keeper.RepayPrincipal(suite.ctx, suite.addrs[0], "xrp", cs(c("usdx", 10)))
	suite.NoError(err)

	_, found := suite.keeper.GetCDP(suite.ctx, "xrp", uint64(1))
	suite.False(found)
	ts = suite.keeper.GetAllCdpsByDenomAndRatio(suite.ctx, "xrp", types.MaxSortableDec)
	suite.Equal(0, len(ts))
	ts = suite.keeper.GetAllCdpsByDenom(suite.ctx, "xrp")
	suite.Equal(0, len(ts))
	sk = suite.app.GetSupplyKeeper()
	acc = sk.GetModuleAccount(suite.ctx, types.ModuleName)
	suite.Equal(sdk.Coins(nil), acc.GetCoins())

}

func (suite *DrawTestSuite) TestAddRepayPrincipalFees() {
	err := suite.keeper.AddCdp(suite.ctx, suite.addrs[2], cs(c("xrp", 1000000000000)), cs(c("usdx", 100000000000)))
	suite.NoError(err)
	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Minute * 10))
	err = suite.keeper.AddPrincipal(suite.ctx, suite.addrs[2], "xrp", cs(c("usdx", 10000000)))
	suite.NoError(err)
	t, _ := suite.keeper.GetCDP(suite.ctx, "xrp", uint64(2))
	suite.Equal(cs(c("usdx", 92827)), t.AccumulatedFees)
	_ = suite.keeper.MintDebtCoins(suite.ctx, types.ModuleName, "debt", cs(c("usdx", 92827)))
	err = suite.keeper.RepayPrincipal(suite.ctx, suite.addrs[2], "xrp", cs(c("usdx", 100)))
	suite.NoError(err)
	t, _ = suite.keeper.GetCDP(suite.ctx, "xrp", uint64(2))
	suite.Equal(cs(c("usdx", 92727)), t.AccumulatedFees)
	err = suite.keeper.RepayPrincipal(suite.ctx, suite.addrs[2], "xrp", cs(c("usdx", 100010092727)))
	suite.NoError(err)
	_, f := suite.keeper.GetCDP(suite.ctx, "xrp", uint64(2))
	suite.False(f)

	err = suite.keeper.AddCdp(suite.ctx, suite.addrs[2], cs(c("xrp", 1000000)), cs(c("usdx", 100)))
	suite.NoError(err)

	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Second * 31536000))
	err = suite.keeper.AddPrincipal(suite.ctx, suite.addrs[2], "xrp", cs(c("usdx", 100)))
	suite.NoError(err)
	t, _ = suite.keeper.GetCDP(suite.ctx, "xrp", uint64(3))
	suite.Equal(cs(c("usdx", 5)), t.AccumulatedFees)
}

func (suite *DrawTestSuite) TestPricefeedFailure() {
	ctx := suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Hour * 2))
	pfk := suite.app.GetPriceFeedKeeper()
	pfk.SetCurrentPrices(ctx, "xrp:usd")
	err := suite.keeper.AddPrincipal(ctx, suite.addrs[0], "xrp", cs(c("usdx", 10)))
	suite.Error(err)
	err = suite.keeper.RepayPrincipal(ctx, suite.addrs[0], "xrp", cs(c("usdx", 10)))
	suite.NoError(err)
}

func (suite *DrawTestSuite) TestModuleAccountFailure() {
	suite.Panics(func() {
		ctx := suite.ctx.WithBlockHeader(suite.ctx.BlockHeader())
		sk := suite.app.GetSupplyKeeper()
		acc := sk.GetModuleAccount(ctx, types.ModuleName)
		ak := suite.app.GetAccountKeeper()
		ak.RemoveAccount(ctx, acc)
		_ = suite.keeper.RepayPrincipal(ctx, suite.addrs[0], "xrp", cs(c("usdx", 10)))
	})
}

func TestDrawTestSuite(t *testing.T) {
	suite.Run(t, new(DrawTestSuite))
}
