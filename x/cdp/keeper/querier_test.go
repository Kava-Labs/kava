package keeper_test

import (
	"math/rand"
	"sort"
	"strings"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp/keeper"
	"github.com/kava-labs/kava/x/cdp/types"
	pfkeeper "github.com/kava-labs/kava/x/pricefeed/keeper"
	pftypes "github.com/kava-labs/kava/x/pricefeed/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

const (
	custom = "custom"
)

type QuerierTestSuite struct {
	suite.Suite

	keeper          keeper.Keeper
	pricefeedKeeper pfkeeper.Keeper
	addrs           []sdk.AccAddress
	app             app.TestApp
	cdps            types.CDPs
	augmentedCDPs   types.AugmentedCDPs
	ctx             sdk.Context
	querier         sdk.Querier
}

func (suite *QuerierTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	cdps := make(types.CDPs, 100)
	augmentedCDPs := make(types.AugmentedCDPs, 100)
	_, addrs := app.GeneratePrivKeyAddressPairs(100)
	coins := []sdk.Coins{}

	for j := 0; j < 100; j++ {
		coins = append(coins, cs(c("btc", 10000000000), c("xrp", 10000000000)))
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
	suite.pricefeedKeeper = tApp.GetPriceFeedKeeper()

	// Set up markets
	oracle := addrs[9]
	marketParams := pftypes.Params{
		Markets: pftypes.Markets{
			pftypes.Market{MarketID: "xrp-usd", BaseAsset: "xrp", QuoteAsset: "usd", Oracles: []sdk.AccAddress{oracle}, Active: true},
			pftypes.Market{MarketID: "btc-usd", BaseAsset: "btc", QuoteAsset: "usd", Oracles: []sdk.AccAddress{oracle}, Active: true},
		},
	}
	suite.pricefeedKeeper.SetParams(ctx, marketParams)

	// Set collateral prices for use in collateralization calculations
	_, err := suite.pricefeedKeeper.SetPrice(
		ctx, oracle, "xrp-usd",
		sdk.MustNewDecFromStr("0.75"),
		time.Now().Add(1*time.Hour))
	suite.Nil(err)

	_, err = suite.pricefeedKeeper.SetPrice(
		ctx, oracle, "btc-usd",
		sdk.MustNewDecFromStr("5000"),
		time.Now().Add(1*time.Hour))
	suite.Nil(err)

	for j := 0; j < 100; j++ {
		collateral := "xrp"
		amount := simulation.RandIntBetween(rand.New(rand.NewSource(int64(j))), 2500000000, 9000000000)
		debt := simulation.RandIntBetween(rand.New(rand.NewSource(int64(j))), 50000000, 250000000)
		if j%2 == 0 {
			collateral = "btc"
			amount = simulation.RandIntBetween(rand.New(rand.NewSource(int64(j))), 500000000, 5000000000)
			debt = simulation.RandIntBetween(rand.New(rand.NewSource(int64(j))), 1000000000, 25000000000)
		}
		suite.Nil(suite.keeper.AddCdp(suite.ctx, addrs[j], cs(c(collateral, int64(amount))), cs(c("usdx", int64(debt)))))
		c, f := suite.keeper.GetCDP(suite.ctx, collateral, uint64(j+1))
		suite.True(f)
		cdps[j] = c
		aCDP, _ := suite.keeper.LoadAugmentedCDP(suite.ctx, c)
		augmentedCDPs[j] = aCDP
	}

	suite.cdps = cdps
	suite.augmentedCDPs = augmentedCDPs
	suite.querier = keeper.NewQuerier(suite.keeper)
	suite.addrs = addrs
}

func (suite *QuerierTestSuite) TestQueryCdp() {
	ctx := suite.ctx.WithIsCheckTx(false)
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryGetCdp}, "/"),
		Data: types.ModuleCdc.MustMarshalJSON(types.NewQueryCdpParams(suite.cdps[0].Owner, suite.cdps[0].Collateral[0].Denom)),
	}
	bz, err := suite.querier(ctx, []string{types.QueryGetCdp}, query)
	suite.Nil(err)
	suite.NotNil(bz)

	var c types.AugmentedCDP
	suite.Nil(types.ModuleCdc.UnmarshalJSON(bz, &c))
	suite.Equal(suite.augmentedCDPs[0], c)

	query = abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryGetCdp}, "/"),
		Data: types.ModuleCdc.MustMarshalJSON(types.NewQueryCdpParams(suite.cdps[0].Owner, "lol")),
	}
	_, err = suite.querier(ctx, []string{types.QueryGetCdp}, query)
	suite.Error(err)

	query = abci.RequestQuery{
		Path: strings.Join([]string{custom, "nonsense"}, "/"),
		Data: []byte("nonsense"),
	}

	_, err = suite.querier(ctx, []string{query.Path}, query)
	suite.Error(err)

	_, err = suite.querier(ctx, []string{types.QueryGetCdp}, query)
	suite.Error(err)

	query = abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryGetCdp}, "/"),
		Data: types.ModuleCdc.MustMarshalJSON(types.NewQueryCdpParams(suite.cdps[0].Owner, "xrp")),
	}
	_, err = suite.querier(ctx, []string{types.QueryGetCdp}, query)
	suite.Error(err)

}

func (suite *QuerierTestSuite) TestQueryCdpsByDenom() {
	ctx := suite.ctx.WithIsCheckTx(false)
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryGetCdps}, "/"),
		Data: types.ModuleCdc.MustMarshalJSON(types.NewQueryCdpsParams(suite.cdps[0].Collateral[0].Denom)),
	}
	bz, err := suite.querier(ctx, []string{types.QueryGetCdps}, query)
	suite.Nil(err)
	suite.NotNil(bz)

	var c types.AugmentedCDPs
	suite.Nil(types.ModuleCdc.UnmarshalJSON(bz, &c))
	suite.Equal(50, len(c))

	query = abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryGetCdps}, "/"),
		Data: types.ModuleCdc.MustMarshalJSON(types.NewQueryCdpsParams("lol")),
	}
	_, err = suite.querier(ctx, []string{types.QueryGetCdps}, query)
	suite.Error(err)
}

func (suite *QuerierTestSuite) TestQueryCdpsByRatio() {
	ratioCountBtc := 0
	ratioCountXrp := 0
	xrpRatio := d("2.0")
	btcRatio := d("2500")
	expectedXrpIds := []int{}
	expectedBtcIds := []int{}
	for _, cdp := range suite.cdps {
		absoluteRatio := suite.keeper.CalculateCollateralToDebtRatio(suite.ctx, cdp.Collateral, cdp.Principal)
		collateralizationRatio, err := suite.keeper.CalculateCollateralizationRatioFromAbsoluteRatio(suite.ctx, cdp.Collateral[0].Denom, absoluteRatio)
		suite.Nil(err)
		if cdp.Collateral[0].Denom == "xrp" {
			if collateralizationRatio.LT(xrpRatio) {
				ratioCountXrp += 1
				expectedXrpIds = append(expectedXrpIds, int(cdp.ID))
			}
		} else {
			if collateralizationRatio.LT(btcRatio) {
				ratioCountBtc += 1
				expectedBtcIds = append(expectedBtcIds, int(cdp.ID))
			}
		}
	}

	ctx := suite.ctx.WithIsCheckTx(false)
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryGetCdpsByCollateralization}, "/"),
		Data: types.ModuleCdc.MustMarshalJSON(types.NewQueryCdpsByRatioParams("xrp", xrpRatio)),
	}
	bz, err := suite.querier(ctx, []string{types.QueryGetCdpsByCollateralization}, query)
	suite.Nil(err)
	suite.NotNil(bz)

	var c types.AugmentedCDPs
	actualXrpIds := []int{}
	suite.Nil(types.ModuleCdc.UnmarshalJSON(bz, &c))
	for _, k := range c {
		actualXrpIds = append(actualXrpIds, int(k.ID))
	}
	sort.Ints(actualXrpIds)
	suite.Equal(expectedXrpIds, actualXrpIds)

	query = abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryGetCdpsByCollateralization}, "/"),
		Data: types.ModuleCdc.MustMarshalJSON(types.NewQueryCdpsByRatioParams("btc", btcRatio)),
	}
	bz, err = suite.querier(ctx, []string{types.QueryGetCdpsByCollateralization}, query)
	suite.Nil(err)
	suite.NotNil(bz)

	c = types.AugmentedCDPs{}
	actualBtcIds := []int{}
	suite.Nil(types.ModuleCdc.UnmarshalJSON(bz, &c))
	for _, k := range c {
		actualBtcIds = append(actualBtcIds, int(k.ID))
	}
	sort.Ints(actualBtcIds)
	suite.Equal(expectedBtcIds, actualBtcIds)

	query = abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryGetCdpsByCollateralization}, "/"),
		Data: types.ModuleCdc.MustMarshalJSON(types.NewQueryCdpsByRatioParams("xrp", d("0.003"))),
	}
	bz, err = suite.querier(ctx, []string{types.QueryGetCdpsByCollateralization}, query)
	suite.Nil(err)
	suite.NotNil(bz)
	c = types.AugmentedCDPs{}
	suite.Nil(types.ModuleCdc.UnmarshalJSON(bz, &c))
	suite.Equal(0, len(c))
}

func (suite *QuerierTestSuite) TestQueryParams() {
	ctx := suite.ctx.WithIsCheckTx(false)
	bz, err := suite.querier(ctx, []string{types.QueryGetParams}, abci.RequestQuery{})
	suite.Nil(err)
	suite.NotNil(bz)

	var p types.Params
	suite.Nil(types.ModuleCdc.UnmarshalJSON(bz, &p))

	cdpGS := NewCDPGenStateMulti()
	gs := types.GenesisState{}
	types.ModuleCdc.UnmarshalJSON(cdpGS["cdp"], &gs)
	suite.Equal(gs.Params, p)
}

func (suite *QuerierTestSuite) TestQueryDeposits() {
	ctx := suite.ctx.WithIsCheckTx(false)
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryGetCdpDeposits}, "/"),
		Data: types.ModuleCdc.MustMarshalJSON(types.NewQueryCdpDeposits(suite.cdps[0].Owner, suite.cdps[0].Collateral[0].Denom)),
	}

	bz, err := suite.querier(ctx, []string{types.QueryGetCdpDeposits}, query)
	suite.Nil(err)
	suite.NotNil(bz)

	deposits := suite.keeper.GetDeposits(ctx, suite.cdps[0].ID)

	var d types.Deposits
	suite.Nil(types.ModuleCdc.UnmarshalJSON(bz, &d))
	suite.Equal(deposits, d)

}

func TestQuerierTestSuite(t *testing.T) {
	suite.Run(t, new(QuerierTestSuite))
}
