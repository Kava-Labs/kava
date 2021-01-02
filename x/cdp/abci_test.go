package cdp_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/auction"
	"github.com/kava-labs/kava/x/cdp"
)

type ModuleTestSuite struct {
	suite.Suite

	keeper cdp.Keeper
	addrs  []sdk.AccAddress
	app    app.TestApp
	ctx    sdk.Context
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

	const numAddrs = 100
	for j := 0; j < numAddrs; j++ {
		coins = append(coins, cs(c("btc", 100000000), c("xrp", 10000000000)))
	}
	_, addrs := app.GeneratePrivKeyAddressPairs(numAddrs)

	tApp.InitializeFromGenesisStates(
		app.NewAuthGenState(addrs, coins),
		NewPricefeedGenStateMulti(),
		NewCDPGenStateMulti(),
	)
	suite.ctx = ctx
	suite.app = tApp
	suite.keeper = tApp.GetCDPKeeper()
	suite.addrs = addrs
}

func createCDPs(ctx sdk.Context, keeper cdp.Keeper, addrs []sdk.AccAddress, numCDPs int) (liquidationTracker, error) {
	tracker := liquidationTracker{}

	for j := 0; j < numCDPs; j++ {
		var collateral string
		var amount, debt int

		if j%2 == 0 {
			collateral = "btc"
			amount = 100000000
			debt = simulation.RandIntBetween(rand.New(rand.NewSource(int64(j))), 2700000000, 5332000000)
			if debt >= 4000000000 {
				tracker.btc = append(tracker.btc, uint64(j+1))
				tracker.debt += int64(debt)
			}
		} else {
			collateral = "xrp"
			amount = 10000000000
			debt = simulation.RandIntBetween(rand.New(rand.NewSource(int64(j))), 750000000, 1249000000)
			if debt >= 1000000000 {
				tracker.xrp = append(tracker.xrp, uint64(j+1))
				tracker.debt += int64(debt)
			}
		}
		err := keeper.AddCdp(ctx, addrs[j], c(collateral, int64(amount)), c("usdx", int64(debt)), collateral+"-a")
		if err != nil {
			return liquidationTracker{}, err
		}
		_, f := keeper.GetCDP(ctx, collateral+"-a", uint64(j+1))
		if !f {
			return liquidationTracker{}, fmt.Errorf("created cdp, but could not find in store")
		}
	}
	return tracker, nil
}

func (suite *ModuleTestSuite) setPrice(price sdk.Dec, market string) {
	pfKeeper := suite.app.GetPriceFeedKeeper()

	_, err := pfKeeper.SetPrice(suite.ctx, sdk.AccAddress{}, market, price, suite.ctx.BlockTime().Add(time.Hour*3))
	suite.NoError(err)
	err = pfKeeper.SetCurrentPrices(suite.ctx, market)
	suite.NoError(err)
	pp, err := pfKeeper.GetCurrentPrice(suite.ctx, market)
	suite.NoError(err)
	suite.Equal(price, pp.Price)
}
func (suite *ModuleTestSuite) TestBeginBlock() {
	liquidations, err := createCDPs(suite.ctx, suite.keeper, suite.addrs, 100)
	suite.Require().NoError(err)

	sk := suite.app.GetSupplyKeeper()
	acc := sk.GetModuleAccount(suite.ctx, cdp.ModuleName)
	originalXrpCollateral := acc.GetCoins().AmountOf("xrp")
	originalBtcCollateral := acc.GetCoins().AmountOf("btc")

	suite.setPrice(d("0.2"), "xrp:usd")
	suite.setPrice(d("6000"), "btc:usd")
	cdp.BeginBlocker(suite.ctx, abci.RequestBeginBlock{Header: suite.ctx.BlockHeader()}, suite.keeper)

	acc = sk.GetModuleAccount(suite.ctx, cdp.ModuleName)
	finalXrpCollateral := acc.GetCoins().AmountOf("xrp")
	seizedXrpCollateral := originalXrpCollateral.Sub(finalXrpCollateral)
	xrpLiquidations := int(seizedXrpCollateral.Quo(i(10000000000)).Int64())
	suite.Equal(len(liquidations.xrp), xrpLiquidations)

	acc = sk.GetModuleAccount(suite.ctx, cdp.ModuleName)
	finalBtcCollateral := acc.GetCoins().AmountOf("btc")
	seizedBtcCollateral := originalBtcCollateral.Sub(finalBtcCollateral)
	btcLiquidations := int(seizedBtcCollateral.Quo(i(100000000)).Int64())
	suite.Equal(len(liquidations.btc), btcLiquidations)

	acc = sk.GetModuleAccount(suite.ctx, auction.ModuleName)
	suite.Equal(liquidations.debt, acc.GetCoins().AmountOf("debt").Int64())
}

func (suite *ModuleTestSuite) TestSeizeSingleCdpWithFees() {
	err := suite.keeper.AddCdp(suite.ctx, suite.addrs[0], c("xrp", 10000000000), c("usdx", 1000000000), "xrp-a")
	suite.NoError(err)

	suite.Equal(i(1000000000), suite.keeper.GetTotalPrincipal(suite.ctx, "xrp-a", "usdx"))

	sk := suite.app.GetSupplyKeeper()
	cdpMacc := sk.GetModuleAccount(suite.ctx, cdp.ModuleName)
	suite.Equal(i(1000000000), cdpMacc.GetCoins().AmountOf("debt"))

	for i := 0; i < 100; i++ {
		suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Second * 6))
		cdp.BeginBlocker(suite.ctx, abci.RequestBeginBlock{Header: suite.ctx.BlockHeader()}, suite.keeper)
	}

	cdpMacc = sk.GetModuleAccount(suite.ctx, cdp.ModuleName)
	suite.Equal(i(1000000900), (cdpMacc.GetCoins().AmountOf("debt")))
	cdp, _ := suite.keeper.GetCDP(suite.ctx, "xrp-a", 1)

	err = suite.keeper.SeizeCollateral(suite.ctx, cdp)
	suite.NoError(err)
	_, found := suite.keeper.GetCDP(suite.ctx, "xrp-a", 1)
	suite.False(found)
}

func TestModuleTestSuite(t *testing.T) {
	suite.Run(t, new(ModuleTestSuite))
}

func BenchmarkBeginBlocker(b *testing.B) {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})

	const numAddrs = 10_000
	coins := []sdk.Coins{}
	for j := 0; j < numAddrs; j++ {
		coins = append(coins, cs(c("btc", 100_000_000), c("xrp", 10_000_000_000)))
	}
	_, addrs := app.GeneratePrivKeyAddressPairs(numAddrs)

	tApp.InitializeFromGenesisStates(
		app.NewAuthGenState(addrs, coins),
		NewPricefeedGenStateMulti(),
		NewCDPGenStateMulti(),
	)
	_, err := createCDPs(ctx, tApp.GetCDPKeeper(), addrs, 2000)
	if err != nil {
		b.Fatal(err)
	}
	// note: price has not been lowered, so there will be no liquidations in the begin blocker

	b.ResetTimer() // don't count the expensive cdp creation in the benchmark
	for n := 0; n < b.N; n++ {
		// Use a copy of the store in the begin blocker to discard any writes and avoid sequential runs interfering.
		// Exclude this operation from the benchmark time
		b.StopTimer()
		cacheCtx, _ := ctx.CacheContext()
		b.StartTimer()

		cdp.BeginBlocker(cacheCtx, abci.RequestBeginBlock{Header: cacheCtx.BlockHeader()}, tApp.GetCDPKeeper())
	}
}
