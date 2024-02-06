package keeper_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp/keeper"
	"github.com/kava-labs/kava/x/cdp/types"
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
	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})
	cdc := tApp.AppCodec()
	_, addrs := app.GeneratePrivKeyAddressPairs(3)
	coins := []sdk.Coins{
		cs(c("xrp", 500000000), c("btc", 500000000), c("usdx", 10000000000)),
		cs(c("xrp", 200000000)),
		cs(c("xrp", 10000000000000), c("usdx", 100000000000)),
	}

	authGS := app.NewFundedGenStateWithCoins(cdc, coins, addrs)
	tApp.InitializeFromGenesisStates(
		authGS,
		NewPricefeedGenStateMulti(cdc),
		NewCDPGenStateMulti(cdc),
	)
	keeper := tApp.GetCDPKeeper()
	suite.app = tApp
	suite.keeper = keeper
	suite.ctx = ctx
	suite.addrs = addrs
	err := suite.keeper.AddCdp(suite.ctx, addrs[0], c("xrp", 400000000), c("usdx", 10000000), "xrp-a")
	suite.NoError(err)
}

func (suite *DrawTestSuite) TestAddRepayPrincipal() {
	err := suite.keeper.AddPrincipal(suite.ctx, suite.addrs[0], "xrp-a", c("usdx", 10000000))
	suite.NoError(err)

	t, found := suite.keeper.GetCDP(suite.ctx, "xrp-a", uint64(1))
	suite.True(found)
	suite.Equal(c("usdx", 20000000), t.Principal)
	ctd := suite.keeper.CalculateCollateralToDebtRatio(suite.ctx, t.Collateral, "xrp-a", t.Principal.Add(t.AccumulatedFees))
	suite.Equal(d("20.0"), ctd)
	ts := suite.keeper.GetAllCdpsByCollateralTypeAndRatio(suite.ctx, "xrp-a", d("20.0"))
	suite.Equal(0, len(ts))
	ts = suite.keeper.GetAllCdpsByCollateralTypeAndRatio(suite.ctx, "xrp-a", d("20.0").Add(sdk.SmallestDec()))
	suite.Equal(ts[0], t)
	tp := suite.keeper.GetTotalPrincipal(suite.ctx, "xrp-a", "usdx")
	suite.Equal(i(20000000), tp)

	ak := suite.app.GetAccountKeeper()
	bk := suite.app.GetBankKeeper()

	acc := ak.GetModuleAccount(suite.ctx, types.ModuleName)
	suite.Equal(cs(c("xrp", 400000000), c("debt", 20000000)), bk.GetAllBalances(suite.ctx, acc.GetAddress()))

	err = suite.keeper.AddPrincipal(suite.ctx, suite.addrs[0], "xrp-a", c("susd", 10000000))
	suite.Require().True(errors.Is(err, types.ErrInvalidDebtRequest))

	err = suite.keeper.AddPrincipal(suite.ctx, suite.addrs[1], "xrp-a", c("usdx", 10000000))
	suite.Require().True(errors.Is(err, types.ErrCdpNotFound))
	err = suite.keeper.AddPrincipal(suite.ctx, suite.addrs[0], "xrp-a", c("xusd", 10000000))
	suite.Require().True(errors.Is(err, types.ErrInvalidDebtRequest))
	err = suite.keeper.AddPrincipal(suite.ctx, suite.addrs[0], "xrp-a", c("usdx", 311000000))
	suite.Require().True(errors.Is(err, types.ErrInvalidCollateralRatio))

	err = suite.keeper.RepayPrincipal(suite.ctx, suite.addrs[0], "xrp-a", c("usdx", 10000000))
	suite.NoError(err)

	t, found = suite.keeper.GetCDP(suite.ctx, "xrp-a", uint64(1))
	suite.True(found)
	suite.Equal(c("usdx", 10000000), t.Principal)

	ctd = suite.keeper.CalculateCollateralToDebtRatio(suite.ctx, t.Collateral, "xrp-a", t.Principal.Add(t.AccumulatedFees))
	suite.Equal(d("40.0"), ctd)
	ts = suite.keeper.GetAllCdpsByCollateralTypeAndRatio(suite.ctx, "xrp-a", d("40.0"))
	suite.Equal(0, len(ts))
	ts = suite.keeper.GetAllCdpsByCollateralTypeAndRatio(suite.ctx, "xrp-a", d("40.0").Add(sdk.SmallestDec()))
	suite.Equal(ts[0], t)

	ak = suite.app.GetAccountKeeper()
	bk = suite.app.GetBankKeeper()
	acc = ak.GetModuleAccount(suite.ctx, types.ModuleName)
	suite.Equal(cs(c("xrp", 400000000), c("debt", 10000000)), bk.GetAllBalances(suite.ctx, acc.GetAddress()))

	err = suite.keeper.RepayPrincipal(suite.ctx, suite.addrs[0], "xrp-a", c("xusd", 10000000))
	suite.Require().True(errors.Is(err, types.ErrInvalidPayment))
	err = suite.keeper.RepayPrincipal(suite.ctx, suite.addrs[1], "xrp-a", c("xusd", 10000000))
	suite.Require().True(errors.Is(err, types.ErrCdpNotFound))

	err = suite.keeper.RepayPrincipal(suite.ctx, suite.addrs[0], "xrp-a", c("usdx", 9000000))
	suite.Require().True(errors.Is(err, types.ErrBelowDebtFloor))
	err = suite.keeper.RepayPrincipal(suite.ctx, suite.addrs[0], "xrp-a", c("usdx", 10000000))
	suite.NoError(err)

	_, found = suite.keeper.GetCDP(suite.ctx, "xrp-a", uint64(1))
	suite.False(found)
	ts = suite.keeper.GetAllCdpsByCollateralTypeAndRatio(suite.ctx, "xrp-a", types.MaxSortableDec)
	suite.Equal(0, len(ts))
	ts = suite.keeper.GetAllCdpsByCollateralType(suite.ctx, "xrp-a")
	suite.Equal(0, len(ts))

	ak = suite.app.GetAccountKeeper()
	bk = suite.app.GetBankKeeper()
	acc = ak.GetModuleAccount(suite.ctx, types.ModuleName)
	suite.Equal(sdk.Coins{}, bk.GetAllBalances(suite.ctx, acc.GetAddress()))
}

func (suite *DrawTestSuite) TestRepayPrincipalOverpay() {
	err := suite.keeper.RepayPrincipal(suite.ctx, suite.addrs[0], "xrp-a", c("usdx", 20000000))
	suite.NoError(err)
	ak := suite.app.GetAccountKeeper()
	bk := suite.app.GetBankKeeper()

	acc := ak.GetAccount(suite.ctx, suite.addrs[0])
	suite.Equal(i(10000000000), (bk.GetBalance(suite.ctx, acc.GetAddress(), "usdx")).Amount)
	_, found := suite.keeper.GetCDP(suite.ctx, "xrp-a", 1)
	suite.False(found)
}

func (suite *DrawTestSuite) TestPricefeedFailure() {
	ctx := suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Hour * 2))
	pfk := suite.app.GetPriceFeedKeeper()
	err := pfk.SetCurrentPrices(ctx, "xrp:usd")
	suite.Error(err)

	err = suite.keeper.AddPrincipal(ctx, suite.addrs[0], "xrp-a", c("usdx", 10000000))
	suite.Error(err)
	err = suite.keeper.RepayPrincipal(ctx, suite.addrs[0], "xrp-a", c("usdx", 10000000))
	suite.NoError(err)
}

func (suite *DrawTestSuite) TestModuleAccountFailure() {
	ctx := suite.ctx.WithBlockHeader(suite.ctx.BlockHeader())
	ak := suite.app.GetAccountKeeper()
	bk := suite.app.GetBankKeeper()
	acc := ak.GetModuleAccount(ctx, types.ModuleName)

	// Remove module account balance
	ak.RemoveAccount(ctx, acc)
	// Also need to burn coins as account keeper no longer stores balances
	err := bk.BurnCoins(ctx, types.ModuleName, bk.GetAllBalances(ctx, acc.GetAddress()))
	suite.Require().NoError(err)

	suite.Panics(func() {
		// Error ignored here since this should panic
		_ = suite.keeper.RepayPrincipal(ctx, suite.addrs[0], "xrp-a", c("usdx", 10000000))
	})
}

func TestDrawTestSuite(t *testing.T) {
	suite.Run(t, new(DrawTestSuite))
}
