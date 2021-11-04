package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/pricefeed/keeper"
	"github.com/kava-labs/kava/x/pricefeed/types"
	"github.com/stretchr/testify/suite"
	tmprototypes "github.com/tendermint/tendermint/proto/tendermint/types"
)

type grpcQueryTestSuite struct {
	suite.Suite

	tApp   app.TestApp
	ctx    sdk.Context
	keeper keeper.Keeper
	addrs  []sdk.AccAddress
	now    time.Time
}

func (suite *grpcQueryTestSuite) SetupTest() {
	suite.tApp = app.NewTestApp()
	suite.ctx = suite.tApp.NewContext(true, tmprototypes.Header{}).
		WithBlockTime(time.Now().UTC())
	suite.keeper = suite.tApp.GetPriceFeedKeeper()

	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	suite.addrs = addrs

	suite.now = time.Now().UTC()
}

func (suite *grpcQueryTestSuite) TestGrpcParams() {
	tests := []struct {
		giveMsg      string
		giveParams   types.Params
		wantAccepted bool
	}{
		{"default params", types.DefaultParams(), true},
		{"test params", types.NewParams([]types.Market{
			{MarketID: "tstusd", BaseAsset: "tst", QuoteAsset: "usd", Oracles: []string{}, Active: true},
		}), true},
	}

	for _, tt := range tests {
		suite.Run(tt.giveMsg, func() {
			suite.keeper.SetParams(suite.ctx, tt.giveParams)

			res, err := suite.keeper.Params(sdk.WrapSDKContext(suite.ctx), &types.QueryParamsRequest{})

			if tt.wantAccepted {
				suite.NoError(err)
				suite.NoError(tt.giveParams.VerboseEqual(res.Params), "params query should respond with set params")
			} else {
				suite.Error(err)
			}
		})
	}
}

func (suite *grpcQueryTestSuite) TestGrpcPrice() {
	params := types.NewParams([]types.Market{
		{MarketID: "tstusd", BaseAsset: "tst", QuoteAsset: "usd", Oracles: []string{}, Active: true},
	})
	suite.keeper.SetParams(suite.ctx, params)

	// No prices set yet, should error
	_, err := suite.keeper.Price(sdk.WrapSDKContext(suite.ctx), &types.QueryPriceRequest{MarketId: "tstusd"})
	suite.ErrorIs(types.ErrNoValidPrice, err)

	suite.setTstPrice()

	expectedPrice := types.NewCurrentPrice("tstusd", sdk.MustNewDecFromStr("0.34"))

	res, err := suite.keeper.Price(sdk.WrapSDKContext(suite.ctx), &types.QueryPriceRequest{MarketId: "tstusd"})
	suite.NoError(err)
	// Price median of addTstPrices below
	suite.Equal(expectedPrice, res.Price)

	_, err = suite.keeper.Price(sdk.WrapSDKContext(suite.ctx), &types.QueryPriceRequest{MarketId: "invalid"})
	suite.Equal("rpc error: code = NotFound desc = invalid market ID", err.Error())

	prices, err := suite.keeper.Prices(sdk.WrapSDKContext(suite.ctx), &types.QueryPricesRequest{})
	suite.NoError(err)

	// Includes previous same price
	suite.Contains(prices.Prices, expectedPrice)
}

func (suite *grpcQueryTestSuite) TestGrpcRawPrices() {
	params := types.NewParams([]types.Market{
		{MarketID: "tstusd", BaseAsset: "tst", QuoteAsset: "usd", Oracles: []string{}, Active: true},
	})
	suite.keeper.SetParams(suite.ctx, params)

	suite.setTstPrice()

	res, err := suite.keeper.RawPrices(sdk.WrapSDKContext(suite.ctx), &types.QueryRawPricesRequest{MarketId: "tstusd"})
	suite.NoError(err)

	suite.Equal(3, len(res.RawPrices))

	suite.ElementsMatch(
		res.RawPrices,
		[]types.PostedPrice{
			types.NewPostedPrice(
				"tstusd",
				suite.addrs[0].String(),
				sdk.MustNewDecFromStr("0.33"),
				suite.now.Add(time.Hour*1),
			),
			types.NewPostedPrice(
				"tstusd",
				suite.addrs[1].String(),
				sdk.MustNewDecFromStr("0.35"),
				suite.now.Add(time.Hour*1),
			),
			types.NewPostedPrice(
				"tstusd",
				suite.addrs[2].String(),
				sdk.MustNewDecFromStr("0.34"),
				suite.now.Add(time.Hour*1),
			),
		},
	)

	_, err = suite.keeper.RawPrices(sdk.WrapSDKContext(suite.ctx), &types.QueryRawPricesRequest{MarketId: "invalid"})
	suite.Equal("rpc error: code = NotFound desc = invalid market ID", err.Error())
}

func (suite *grpcQueryTestSuite) TestGrpcOracles() {
	params := types.NewParams([]types.Market{
		{MarketID: "tstusd", BaseAsset: "tst", QuoteAsset: "usd", Oracles: []string{}, Active: true},
	})
	suite.keeper.SetParams(suite.ctx, params)

	res, err := suite.keeper.Oracles(sdk.WrapSDKContext(suite.ctx), &types.QueryOraclesRequest{MarketId: "tstusd"})
	suite.NoError(err)
	suite.Empty(res.Oracles)

	var oracles []string
	for _, a := range suite.addrs {
		oracles = append(oracles, a.String())
	}

	params = types.NewParams([]types.Market{
		{MarketID: "tstusd", BaseAsset: "tst", QuoteAsset: "usd", Oracles: oracles, Active: true},
	})
	suite.keeper.SetParams(suite.ctx, params)

	res, err = suite.keeper.Oracles(sdk.WrapSDKContext(suite.ctx), &types.QueryOraclesRequest{MarketId: "tstusd"})
	suite.NoError(err)
	suite.ElementsMatch(res.Oracles, oracles)

	_, err = suite.keeper.Oracles(sdk.WrapSDKContext(suite.ctx), &types.QueryOraclesRequest{MarketId: "invalid"})
	suite.Equal("rpc error: code = NotFound desc = invalid market ID", err.Error())
}

func (suite *grpcQueryTestSuite) TestGrpcMarkets() {
	params := types.NewParams([]types.Market{
		{MarketID: "tstusd", BaseAsset: "tst", QuoteAsset: "usd", Oracles: []string{}, Active: true},
	})
	suite.keeper.SetParams(suite.ctx, params)

	res, err := suite.keeper.Markets(sdk.WrapSDKContext(suite.ctx), &types.QueryMarketsRequest{})
	suite.NoError(err)
	suite.Len(res.Markets, 1)
	suite.Equal(len(res.Markets), len(params.Markets))
	suite.NoError(res.Markets[0].VerboseEqual(params.Markets[0]))
}

func (suite *grpcQueryTestSuite) setTstPrice() {
	_, err := suite.keeper.SetPrice(
		suite.ctx, suite.addrs[0].String(), "tstusd",
		sdk.MustNewDecFromStr("0.33"),
		suite.now.Add(time.Hour*1))
	suite.NoError(err)

	_, err = suite.keeper.SetPrice(
		suite.ctx, suite.addrs[1].String(), "tstusd",
		sdk.MustNewDecFromStr("0.35"),
		suite.now.Add(time.Hour*1))
	suite.NoError(err)

	_, err = suite.keeper.SetPrice(
		suite.ctx, suite.addrs[2].String(), "tstusd",
		sdk.MustNewDecFromStr("0.34"),
		suite.now.Add(time.Hour*1))
	suite.NoError(err)

	err = suite.keeper.SetCurrentPrices(suite.ctx, "tstusd")
	suite.NoError(err)
}

func TestGrpcQueryTestSuite(t *testing.T) {
	suite.Run(t, new(grpcQueryTestSuite))
}
