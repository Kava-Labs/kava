package pricefeed_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/pricefeed"
	"github.com/kava-labs/kava/x/pricefeed/keeper"
	"github.com/kava-labs/kava/x/pricefeed/types"
	"github.com/stretchr/testify/suite"
	tmprototypes "github.com/tendermint/tendermint/proto/tendermint/types"
)

type grpcQueryTestSuite struct {
	suite.Suite

	tApp        app.TestApp
	ctx         sdk.Context
	keeper      keeper.Keeper
	queryServer types.QueryServer
	addrs       []sdk.AccAddress
	strAddrs    []string
	now         time.Time
}

func (suite *grpcQueryTestSuite) SetupTest() {
	suite.tApp = app.NewTestApp()
	suite.ctx = suite.tApp.NewContext(true, tmprototypes.Header{}).
		WithBlockTime(time.Now().UTC())
	suite.keeper = suite.tApp.GetPriceFeedKeeper()
	suite.queryServer = pricefeed.NewQueryServerImpl(suite.keeper)

	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	suite.addrs = addrs

	var strAddrs []string
	for _, a := range addrs {
		strAddrs = append(strAddrs, a.String())
	}
	suite.strAddrs = strAddrs

	suite.now = time.Now().UTC()
}

func (suite *grpcQueryTestSuite) setTestParams() {
	params := types.NewParams([]types.Market{
		{MarketID: "tstusd", BaseAsset: "tst", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
	})
	suite.keeper.SetParams(suite.ctx, params)
}

func (suite *grpcQueryTestSuite) TestGrpcParams() {
	tests := []struct {
		giveMsg      string
		giveParams   types.Params
		wantAccepted bool
	}{
		{"default params", types.DefaultParams(), true},
		{"test params", types.NewParams([]types.Market{
			{MarketID: "tstusd", BaseAsset: "tst", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
		}), true},
	}

	for _, tt := range tests {
		suite.Run(tt.giveMsg, func() {
			suite.keeper.SetParams(suite.ctx, tt.giveParams)

			res, err := suite.queryServer.Params(sdk.WrapSDKContext(suite.ctx), &types.QueryParamsRequest{})

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
	suite.setTestParams()
	suite.setTstPrice()

	expectedPrice := types.NewCurrentPriceResponse("tstusd", sdk.MustNewDecFromStr("0.34"))

	res, err := suite.queryServer.Price(sdk.WrapSDKContext(suite.ctx), &types.QueryPriceRequest{MarketId: "tstusd"})
	suite.NoError(err)
	suite.Equal(expectedPrice, res.Price)
}

func (suite *grpcQueryTestSuite) TestGrpcPrice_NoPriceSet() {
	suite.setTestParams()

	// No prices set yet, should error
	_, err := suite.queryServer.Price(sdk.WrapSDKContext(suite.ctx), &types.QueryPriceRequest{MarketId: "tstusd"})
	suite.ErrorIs(types.ErrNoValidPrice, err)
}

func (suite *grpcQueryTestSuite) TestGrpcPrice_InvalidMarket() {
	suite.setTestParams()
	suite.setTstPrice()

	_, err := suite.queryServer.Price(sdk.WrapSDKContext(suite.ctx), &types.QueryPriceRequest{MarketId: "invalid"})
	suite.Equal("rpc error: code = NotFound desc = invalid market ID", err.Error())
}

func (suite *grpcQueryTestSuite) TestGrpcPrices() {
	suite.setTestParams()
	suite.setTstPrice()

	expectedPrice := types.NewCurrentPriceResponse("tstusd", sdk.MustNewDecFromStr("0.34"))

	prices, err := suite.queryServer.Prices(sdk.WrapSDKContext(suite.ctx), &types.QueryPricesRequest{})
	suite.NoError(err)

	suite.Contains(prices.Prices, expectedPrice, "all prices should include the tstusd price")
}

func (suite *grpcQueryTestSuite) TestGrpcRawPrices() {
	suite.setTestParams()
	suite.setTstPrice()

	res, err := suite.queryServer.RawPrices(sdk.WrapSDKContext(suite.ctx), &types.QueryRawPricesRequest{MarketId: "tstusd"})
	suite.NoError(err)

	suite.Equal(3, len(res.RawPrices))

	suite.ElementsMatch(
		res.RawPrices,
		[]types.PostedPriceResponse{
			types.NewPostedPriceResponse(
				"tstusd",
				suite.addrs[0],
				sdk.MustNewDecFromStr("0.33"),
				suite.now.Add(time.Hour*1),
			),
			types.NewPostedPriceResponse(
				"tstusd",
				suite.addrs[1],
				sdk.MustNewDecFromStr("0.35"),
				suite.now.Add(time.Hour*1),
			),
			types.NewPostedPriceResponse(
				"tstusd",
				suite.addrs[2],
				sdk.MustNewDecFromStr("0.34"),
				suite.now.Add(time.Hour*1),
			),
		},
	)
}

func (suite *grpcQueryTestSuite) TestGrpcRawPrices_InvalidMarket() {
	suite.setTestParams()
	suite.setTstPrice()

	_, err := suite.queryServer.RawPrices(sdk.WrapSDKContext(suite.ctx), &types.QueryRawPricesRequest{MarketId: "invalid"})
	suite.Equal("rpc error: code = NotFound desc = invalid market ID", err.Error())
}

func (suite *grpcQueryTestSuite) TestGrpcOracles_Empty() {
	params := types.NewParams([]types.Market{
		{MarketID: "tstusd", BaseAsset: "tst", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
	})
	suite.keeper.SetParams(suite.ctx, params)

	res, err := suite.queryServer.Oracles(sdk.WrapSDKContext(suite.ctx), &types.QueryOraclesRequest{MarketId: "tstusd"})
	suite.NoError(err)
	suite.Empty(res.Oracles)

	params = types.NewParams([]types.Market{
		{MarketID: "tstusd", BaseAsset: "tst", QuoteAsset: "usd", Oracles: suite.addrs, Active: true},
	})
	suite.keeper.SetParams(suite.ctx, params)

	res, err = suite.queryServer.Oracles(sdk.WrapSDKContext(suite.ctx), &types.QueryOraclesRequest{MarketId: "tstusd"})
	suite.NoError(err)
	suite.ElementsMatch(suite.strAddrs, res.Oracles)

	_, err = suite.queryServer.Oracles(sdk.WrapSDKContext(suite.ctx), &types.QueryOraclesRequest{MarketId: "invalid"})
	suite.Equal("rpc error: code = NotFound desc = invalid market ID", err.Error())
}

func (suite *grpcQueryTestSuite) TestGrpcOracles() {
	params := types.NewParams([]types.Market{
		{MarketID: "tstusd", BaseAsset: "tst", QuoteAsset: "usd", Oracles: suite.addrs, Active: true},
	})
	suite.keeper.SetParams(suite.ctx, params)

	res, err := suite.queryServer.Oracles(sdk.WrapSDKContext(suite.ctx), &types.QueryOraclesRequest{MarketId: "tstusd"})
	suite.NoError(err)
	suite.ElementsMatch(suite.strAddrs, res.Oracles)
}

func (suite *grpcQueryTestSuite) TestGrpcOracles_InvalidMarket() {
	suite.setTestParams()

	_, err := suite.queryServer.Oracles(sdk.WrapSDKContext(suite.ctx), &types.QueryOraclesRequest{MarketId: "invalid"})
	suite.Equal("rpc error: code = NotFound desc = invalid market ID", err.Error())
}

func (suite *grpcQueryTestSuite) TestGrpcMarkets() {
	params := types.NewParams([]types.Market{
		{MarketID: "tstusd", BaseAsset: "tst", QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
	})
	suite.keeper.SetParams(suite.ctx, params)

	res, err := suite.queryServer.Markets(sdk.WrapSDKContext(suite.ctx), &types.QueryMarketsRequest{})
	suite.NoError(err)
	suite.Len(res.Markets, 1)
	suite.Equal(len(res.Markets), len(params.Markets))
	suite.NoError(res.Markets[0].VerboseEqual(params.Markets[0]))
}

func (suite *grpcQueryTestSuite) setTstPrice() {
	_, err := suite.keeper.SetPrice(
		suite.ctx, suite.addrs[0], "tstusd",
		sdk.MustNewDecFromStr("0.33"),
		suite.now.Add(time.Hour*1))
	suite.NoError(err)

	_, err = suite.keeper.SetPrice(
		suite.ctx, suite.addrs[1], "tstusd",
		sdk.MustNewDecFromStr("0.35"),
		suite.now.Add(time.Hour*1))
	suite.NoError(err)

	_, err = suite.keeper.SetPrice(
		suite.ctx, suite.addrs[2], "tstusd",
		sdk.MustNewDecFromStr("0.34"),
		suite.now.Add(time.Hour*1))
	suite.NoError(err)

	err = suite.keeper.SetCurrentPrices(suite.ctx, "tstusd")
	suite.NoError(err)
}

func TestGrpcQueryTestSuite(t *testing.T) {
	suite.Run(t, new(grpcQueryTestSuite))
}
