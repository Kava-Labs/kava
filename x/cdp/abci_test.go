package cdp_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	auctiontypes "github.com/kava-labs/kava/x/auction/types"
	"github.com/kava-labs/kava/x/cdp"
	"github.com/kava-labs/kava/x/cdp/keeper"
	"github.com/kava-labs/kava/x/cdp/types"
)

type ModuleTestSuite struct {
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

func (suite *ModuleTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now(), ChainID: app.TestChainID})
	tracker := liquidationTracker{}

	coins := cs(c("btc", 100000000), c("xrp", 10000000000))
	_, addrs := app.GeneratePrivKeyAddressPairs(100)
	authGS := app.NewFundedGenStateWithSameCoins(tApp.AppCodec(), coins, addrs)
	tApp.InitDefaultGenesis(
		ctx,
		authGS,
		NewPricefeedGenStateMulti(tApp.AppCodec()),
		NewCDPGenStateMulti(tApp.AppCodec()),
	)
	suite.ctx = ctx
	suite.app = tApp
	suite.keeper = tApp.GetCDPKeeper()
	suite.cdps = types.CDPs{}
	suite.addrs = addrs
	suite.liquidations = tracker
}

func (suite *ModuleTestSuite) createCdps() {
	cdps := make(types.CDPs, 100)
	tracker := liquidationTracker{}

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
		suite.Nil(suite.keeper.AddCdp(suite.ctx, suite.addrs[j], c(collateral, int64(amount)), c("usdx", int64(debt)), collateral+"-a"))
		c, f := suite.keeper.GetCDP(suite.ctx, collateral+"-a", uint64(j+1))
		suite.True(f)
		cdps[j] = c
	}

	suite.cdps = cdps
	suite.liquidations = tracker
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
	suite.createCdps()
	ak := suite.app.GetAccountKeeper()
	bk := suite.app.GetBankKeeper()

	acc := ak.GetModuleAccount(suite.ctx, types.ModuleName)
	originalXrpCollateral := bk.GetBalance(suite.ctx, acc.GetAddress(), "xrp").Amount
	suite.setPrice(d("0.2"), "xrp:usd")
	cdp.BeginBlocker(suite.ctx, abci.RequestBeginBlock{Header: suite.ctx.BlockHeader()}, suite.keeper)
	acc = ak.GetModuleAccount(suite.ctx, types.ModuleName)
	finalXrpCollateral := bk.GetBalance(suite.ctx, acc.GetAddress(), "xrp").Amount
	seizedXrpCollateral := originalXrpCollateral.Sub(finalXrpCollateral)
	xrpLiquidations := int(seizedXrpCollateral.Quo(i(10000000000)).Int64())
	suite.Equal(10, xrpLiquidations)

	acc = ak.GetModuleAccount(suite.ctx, types.ModuleName)
	originalBtcCollateral := bk.GetBalance(suite.ctx, acc.GetAddress(), "btc").Amount
	suite.setPrice(d("6000"), "btc:usd")
	cdp.BeginBlocker(suite.ctx, abci.RequestBeginBlock{Header: suite.ctx.BlockHeader()}, suite.keeper)
	acc = ak.GetModuleAccount(suite.ctx, types.ModuleName)
	finalBtcCollateral := bk.GetBalance(suite.ctx, acc.GetAddress(), "btc").Amount
	seizedBtcCollateral := originalBtcCollateral.Sub(finalBtcCollateral)
	btcLiquidations := int(seizedBtcCollateral.Quo(i(100000000)).Int64())
	suite.Equal(10, btcLiquidations)

	acc = ak.GetModuleAccount(suite.ctx, auctiontypes.ModuleName)
	suite.Equal(int64(71955653865), bk.GetBalance(suite.ctx, acc.GetAddress(), "debt").Amount.Int64())
}

func (suite *ModuleTestSuite) TestSeizeSingleCdpWithFees() {
	err := suite.keeper.AddCdp(suite.ctx, suite.addrs[0], c("xrp", 10000000000), c("usdx", 1000000000), "xrp-a")
	suite.NoError(err)
	suite.Equal(i(1000000000), suite.keeper.GetTotalPrincipal(suite.ctx, "xrp-a", "usdx"))
	ak := suite.app.GetAccountKeeper()
	bk := suite.app.GetBankKeeper()

	cdpMacc := ak.GetModuleAccount(suite.ctx, types.ModuleName)
	suite.Equal(i(1000000000), bk.GetBalance(suite.ctx, cdpMacc.GetAddress(), "debt").Amount)
	for i := 0; i < 100; i++ {
		suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Second * 6))
		cdp.BeginBlocker(suite.ctx, abci.RequestBeginBlock{Header: suite.ctx.BlockHeader()}, suite.keeper)
	}

	cdpMacc = ak.GetModuleAccount(suite.ctx, types.ModuleName)
	suite.Equal(i(1000000891), (bk.GetBalance(suite.ctx, cdpMacc.GetAddress(), "debt").Amount))
	cdp, _ := suite.keeper.GetCDP(suite.ctx, "xrp-a", 1)

	err = suite.keeper.SeizeCollateral(suite.ctx, cdp)
	suite.NoError(err)
	_, found := suite.keeper.GetCDP(suite.ctx, "xrp-a", 1)
	suite.False(found)
}

func TestModuleTestSuite(t *testing.T) {
	suite.Run(t, new(ModuleTestSuite))
}
