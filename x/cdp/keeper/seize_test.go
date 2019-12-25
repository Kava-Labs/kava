package keeper_test

import (
	"math/rand"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp/keeper"
	"github.com/kava-labs/kava/x/cdp/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

type SeizeTestSuite struct {
	suite.Suite

	keeper       keeper.Keeper
	addrs        []sdk.AccAddress
	app          app.TestApp
	cdps         types.CDPs
	ctx          sdk.Context
	liquidations liquidationTracker
}

type liquidationTracker struct {
	xrp  []uint64
	btc  []uint64
	debt int64
}

func (suite *SeizeTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	cdps := make(types.CDPs, 100)
	_, addrs := app.GeneratePrivKeyAddressPairs(100)
	coins := []sdk.Coins{}
	tracker := liquidationTracker{}

	for j := 0; j < 100; j++ {
		coins = append(coins, cs(c("btc", 1), c("xrp", 10000)))
	}

	authGS := app.NewAuthGenState(
		addrs, coins)
	tApp.InitializeFromGenesisStates(
		authGS,
		NewPricefeedGenStateMulti(),
		NewCDPGenStateMulti(),
	)

	suite.ctx = ctx
	suite.app = tApp
	suite.keeper = tApp.GetCDPKeeper()

	for j := 0; j < 100; j++ {
		collateral := "xrp"
		amount := 10000
		debt := simulation.RandIntBetween(rand.New(rand.NewSource(int64(j))), 750, 1249)
		if j%2 == 0 {
			collateral = "btc"
			amount = 1
			debt = simulation.RandIntBetween(rand.New(rand.NewSource(int64(j))), 2700, 5332)
			if debt >= 4000 {
				tracker.btc = append(tracker.btc, uint64(j+1))
				tracker.debt += int64(debt)
			}
		} else {
			if debt >= 1000 {
				tracker.xrp = append(tracker.xrp, uint64(j+1))
				tracker.debt += int64(debt)
			}
		}
		suite.Nil(suite.keeper.AddCdp(suite.ctx, addrs[j], cs(c(collateral, int64(amount))), cs(c("usdx", int64(debt)))))
		c, f := suite.keeper.GetCDP(suite.ctx, collateral, uint64(j+1))
		suite.True(f)
		cdps[j] = c
	}

	suite.cdps = cdps
	suite.addrs = addrs
	suite.liquidations = tracker
}

func (suite *SeizeTestSuite) setPrice(price sdk.Dec, market string) {
	pfKeeper := suite.app.GetPriceFeedKeeper()

	pfKeeper.SetPrice(suite.ctx, sdk.AccAddress{}, market, price, suite.ctx.BlockTime().Add(time.Hour*3))
	err := pfKeeper.SetCurrentPrices(suite.ctx, market)
	suite.NoError(err)
	pp, err := pfKeeper.GetCurrentPrice(suite.ctx, market)
	suite.NoError(err)
	suite.Equal(price, pp.Price)
}

func (suite *SeizeTestSuite) TestSeizeCollateral() {
	sk := suite.app.GetSupplyKeeper()
	cdp, _ := suite.keeper.GetCDP(suite.ctx, "xrp", uint64(2))
	p := cdp.Principal[0].Amount
	cl := cdp.Collateral[0].Amount
	tpb := suite.keeper.GetTotalPrincipal(suite.ctx, "xrp", "usdx")
	suite.keeper.SeizeCollateral(suite.ctx, cdp)
	tpa := suite.keeper.GetTotalPrincipal(suite.ctx, "xrp", "usdx")
	suite.Equal(tpb.Sub(tpa), p)
	liqModAcc := sk.GetModuleAccount(suite.ctx, "liquidator")
	suite.Equal(cs(c("debt", p.Int64()), c("xrp", cl.Int64())), liqModAcc.GetCoins())
	ak := suite.app.GetAccountKeeper()
	acc := ak.GetAccount(suite.ctx, suite.addrs[1])
	suite.Equal(p.Int64(), acc.GetCoins().AmountOf("usdx").Int64())
	err := suite.keeper.WithdrawCollateral(suite.ctx, suite.addrs[1], suite.addrs[1], cs(c("xrp", 10)))
	suite.Equal(types.CodeDepositNotAvailable, err.Result().Code)
}

func (suite *SeizeTestSuite) TestLiquidateCdps() {
	sk := suite.app.GetSupplyKeeper()
	acc := sk.GetModuleAccount(suite.ctx, types.ModuleName)
	originalXrpCollateral := acc.GetCoins().AmountOf("xrp")
	suite.setPrice(d("0.2"), "xrp:usd")
	p, _ := suite.keeper.GetCollateral(suite.ctx, "xrp")
	suite.keeper.LiquidateCdps(suite.ctx, "xrp:usd", "xrp", p.LiquidationRatio)
	acc = sk.GetModuleAccount(suite.ctx, types.ModuleName)
	finalXrpCollateral := acc.GetCoins().AmountOf("xrp")
	seizedXrpCollateral := originalXrpCollateral.Sub(finalXrpCollateral)
	xrpLiquidations := int(seizedXrpCollateral.Quo(i(10000)).Int64())
	suite.Equal(len(suite.liquidations.xrp), xrpLiquidations)
}

func (suite *SeizeTestSuite) TestHandleNewDebt() {
	tpb := suite.keeper.GetTotalPrincipal(suite.ctx, "xrp", "usdx")
	suite.keeper.HandleNewDebt(suite.ctx, "xrp", "usdx", i(31536000))
	tpa := suite.keeper.GetTotalPrincipal(suite.ctx, "xrp", "usdx")
	suite.Equal(sdk.NewDec(tpb.Int64()).Mul(d("1.05")).TruncateInt().Int64(), tpa.Int64())
}

func TestSeizeTestSuite(t *testing.T) {
	suite.Run(t, new(SeizeTestSuite))
}
