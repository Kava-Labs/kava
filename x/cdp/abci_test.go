package cdp_test

import (
	"math/rand"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/auction"
	"github.com/kava-labs/kava/x/cdp"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

type ModuleTestSuite struct {
	suite.Suite

	keeper       cdp.Keeper
	addrs        []sdk.AccAddress
	app          app.TestApp
	cdps         cdp.CDPs
	ctx          sdk.Context
	liquidations liquidationTracker
}

type liquidationTracker struct {
	xrp  []uint64
	btc  []uint64
	debt int64
}

func (suite *ModuleTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	coins := []sdk.Coins{}
	tracker := liquidationTracker{}

	for j := 0; j < 100; j++ {
		coins = append(coins, cs(c("btc", 100000000), c("xrp", 10000000000)))
	}
	_, addrs := app.GeneratePrivKeyAddressPairs(100)

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
	suite.cdps = cdp.CDPs{}
	suite.addrs = addrs
	suite.liquidations = tracker
}

func (suite *ModuleTestSuite) createCdps() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	cdps := make(cdp.CDPs, 100)
	_, addrs := app.GeneratePrivKeyAddressPairs(100)
	coins := []sdk.Coins{}
	tracker := liquidationTracker{}

	for j := 0; j < 100; j++ {
		coins = append(coins, cs(c("btc", 100000000), c("xrp", 10000000000)))
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
		amount := 10000000000
		debt := simulation.RandIntBetween(rand.New(rand.NewSource(int64(j))), 750000000, 1249000000)
		if j%2 == 0 {
			collateral = "btc"
			amount = 100000000
			debt = simulation.RandIntBetween(rand.New(rand.NewSource(int64(j))), 2700000000, 5332000000)
			if debt >= 4000000000 {
				tracker.btc = append(tracker.btc, uint64(j+1))
				tracker.debt += int64(debt)
			}
		} else {
			if debt >= 1000000000 {
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

func (suite *ModuleTestSuite) setPrice(price sdk.Dec, market string) {
	pfKeeper := suite.app.GetPriceFeedKeeper()

	pfKeeper.SetPrice(suite.ctx, sdk.AccAddress{}, market, price, suite.ctx.BlockTime().Add(time.Hour*3))
	err := pfKeeper.SetCurrentPrices(suite.ctx, market)
	suite.NoError(err)
	pp, err := pfKeeper.GetCurrentPrice(suite.ctx, market)
	suite.NoError(err)
	suite.Equal(price, pp.Price)
}
func (suite *ModuleTestSuite) TestBeginBlock() {
	suite.createCdps()
	sk := suite.app.GetSupplyKeeper()
	acc := sk.GetModuleAccount(suite.ctx, cdp.ModuleName)
	originalXrpCollateral := acc.GetCoins().AmountOf("xrp")
	suite.setPrice(d("0.2"), "xrp:usd")
	cdp.BeginBlocker(suite.ctx, abci.RequestBeginBlock{Header: suite.ctx.BlockHeader()}, suite.keeper)
	acc = sk.GetModuleAccount(suite.ctx, cdp.ModuleName)
	finalXrpCollateral := acc.GetCoins().AmountOf("xrp")
	seizedXrpCollateral := originalXrpCollateral.Sub(finalXrpCollateral)
	xrpLiquidations := int(seizedXrpCollateral.Quo(i(10000000000)).Int64())
	suite.Equal(len(suite.liquidations.xrp), xrpLiquidations)

	acc = sk.GetModuleAccount(suite.ctx, cdp.ModuleName)
	originalBtcCollateral := acc.GetCoins().AmountOf("btc")
	suite.setPrice(d("6000"), "btc:usd")
	cdp.BeginBlocker(suite.ctx, abci.RequestBeginBlock{Header: suite.ctx.BlockHeader()}, suite.keeper)
	acc = sk.GetModuleAccount(suite.ctx, cdp.ModuleName)
	finalBtcCollateral := acc.GetCoins().AmountOf("btc")
	seizedBtcCollateral := originalBtcCollateral.Sub(finalBtcCollateral)
	btcLiquidations := int(seizedBtcCollateral.Quo(i(100000000)).Int64())
	suite.Equal(len(suite.liquidations.btc), btcLiquidations)

	acc = sk.GetModuleAccount(suite.ctx, auction.ModuleName)
	suite.Equal(suite.liquidations.debt, acc.GetCoins().AmountOf("debt").Int64())

}

func (suite *ModuleTestSuite) TestSeizeSingleCdpWithFees() {
	err := suite.keeper.AddCdp(suite.ctx, suite.addrs[0], cs(c("xrp", 10000000000)), cs(c("usdx", 1000000000)))
	suite.NoError(err)
	suite.keeper.SetPreviousBlockTime(suite.ctx, suite.ctx.BlockTime())
	suite.Equal(i(1000000000), suite.keeper.GetTotalPrincipal(suite.ctx, "xrp", "usdx"))
	sk := suite.app.GetSupplyKeeper()
	cdpMacc := sk.GetModuleAccount(suite.ctx, cdp.ModuleName)
	suite.Equal(i(1000000000), cdpMacc.GetCoins().AmountOf("debt"))
	for i := 0; i < 100; i++ {
		suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Second * 6))
		cdp.BeginBlocker(suite.ctx, abci.RequestBeginBlock{Header: suite.ctx.BlockHeader()}, suite.keeper)
	}

	cdpMacc = sk.GetModuleAccount(suite.ctx, cdp.ModuleName)
	suite.Equal(i(1000000900), (cdpMacc.GetCoins().AmountOf("debt")))
	cdp, _ := suite.keeper.GetCDP(suite.ctx, "xrp", 1)

	err = suite.keeper.SeizeCollateral(suite.ctx, cdp)
	suite.NoError(err)
	_, found := suite.keeper.GetCDP(suite.ctx, "xrp", 1)
	suite.False(found)
}

func TestModuleTestSuite(t *testing.T) {
	suite.Run(t, new(ModuleTestSuite))
}
