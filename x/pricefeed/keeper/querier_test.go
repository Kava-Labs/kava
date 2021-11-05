package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/x/pricefeed/keeper"
	"github.com/kava-labs/kava/x/pricefeed/types"
)

func (suite *KeeperTestSuite) TestQuerierGetParams() {
	querier := keeper.NewQuerier(suite.keeper, types.ModuleCdc.LegacyAmino)
	bz, err := querier(suite.ctx, []string{types.QueryGetParams}, abci.RequestQuery{})
	suite.Require().NoError(err)
	suite.NotNil(bz)

	expParams := types.Params{
		Markets: []types.Market{
			{MarketID: "btc:usd", BaseAsset: "btc", QuoteAsset: "usd", Oracles: []string{}, Active: true},
			{MarketID: "xrp:usd", BaseAsset: "xrp", QuoteAsset: "usd", Oracles: []string{}, Active: true},
		},
	}
	var p types.Params
	suite.Nil(types.ModuleCdc.UnmarshalJSON(bz, &p))
	suite.Require().NoError(expParams.VerboseEqual(p))
}

func (suite *KeeperTestSuite) TestQuerierGetPrice() {
	// Invalid market
	requestParams := types.NewQueryWithMarketIDParams("invalid")
	data, err := types.ModuleCdc.LegacyAmino.MarshalJSON(requestParams)
	suite.Require().NoError(err)

	querier := keeper.NewQuerier(suite.keeper, types.ModuleCdc.LegacyAmino)
	bz, err := querier(suite.ctx, []string{types.QueryPrice}, abci.RequestQuery{Data: data})
	suite.Require().Error(err)
	suite.Nil(bz)

	// Valid market
	requestParams = types.NewQueryWithMarketIDParams("btc:usd")
	data, err = types.ModuleCdc.LegacyAmino.MarshalJSON(requestParams)
	suite.Require().NoError(err)

	bz, err = querier(suite.ctx, []string{types.QueryPrice}, abci.RequestQuery{Data: data})
	suite.Require().NoError(err)
	suite.NotNil(bz)

	expPrice := types.CurrentPrice{
		MarketID: "btc:usd",
		Price:    sdk.MustNewDecFromStr("8000.00"),
	}
	var p types.CurrentPrice
	suite.Nil(types.ModuleCdc.UnmarshalJSON(bz, &p))
	suite.Require().NoError(expPrice.VerboseEqual(p))
}

func (suite *KeeperTestSuite) TestQuerierGetPrices() {
	querier := keeper.NewQuerier(suite.keeper, types.ModuleCdc.LegacyAmino)
	bz, err := querier(suite.ctx, []string{types.QueryPrices}, abci.RequestQuery{})
	suite.Require().NoError(err)
	suite.NotNil(bz)

	expPrices := []types.CurrentPrice{
		{
			MarketID: "btc:usd",
			Price:    sdk.MustNewDecFromStr("8000.00"),
		},
		{
			MarketID: "xrp:usd",
			Price:    sdk.MustNewDecFromStr("0.25"),
		},
	}
	var p []types.CurrentPrice
	suite.Nil(types.ModuleCdc.LegacyAmino.UnmarshalJSON(bz, &p))
	suite.Require().Len(p, 2, "should have 2 prices")

	suite.Equal(expPrices[0], p[0])
	suite.Equal(expPrices[1], p[1])
}

func (suite *KeeperTestSuite) TestQuerierGetRawPrices() {
	requestParams := types.NewQueryWithMarketIDParams("btc:usd")
	data, err := types.ModuleCdc.LegacyAmino.MarshalJSON(requestParams)
	suite.Require().NoError(err)

	querier := keeper.NewQuerier(suite.keeper, types.ModuleCdc.LegacyAmino)
	bz, err := querier(suite.ctx, []string{types.QueryRawPrices}, abci.RequestQuery{Data: data})
	suite.Require().NoError(err)
	suite.NotNil(bz)

	expPrices := []types.PostedPrice{
		{
			MarketID: "btc:usd",
			Price:    sdk.MustNewDecFromStr("8000.00"),
		},
		{
			MarketID: "xrp:usd",
			Price:    sdk.MustNewDecFromStr("0.25"),
		},
	}
	var p []types.PostedPrice
	suite.Nil(types.ModuleCdc.LegacyAmino.UnmarshalJSON(bz, &p))
	suite.Require().Len(p, 1, "should have 1 raw price for btc:usd")

	// Expire time different so only compare prices
	suite.True(expPrices[0].Price.Equal(p[0].Price), "first posted price should be equal")
}

func (suite *KeeperTestSuite) TestQuerierGetOracles() {
	requestParams := types.NewQueryWithMarketIDParams("btc:usd")
	data, err := types.ModuleCdc.LegacyAmino.MarshalJSON(requestParams)
	suite.Require().NoError(err)

	querier := keeper.NewQuerier(suite.keeper, types.ModuleCdc.LegacyAmino)
	bz, err := querier(suite.ctx, []string{types.QueryOracles}, abci.RequestQuery{Data: data})
	suite.Require().NoError(err)
	suite.NotNil(bz)

	var p []string
	suite.Nil(types.ModuleCdc.LegacyAmino.UnmarshalJSON(bz, &p))
	suite.Require().Empty(p, "there should be no oracles")
}

func (suite *KeeperTestSuite) TestQuerierGetMarkets() {
	querier := keeper.NewQuerier(suite.keeper, types.ModuleCdc.LegacyAmino)
	bz, err := querier(suite.ctx, []string{types.QueryMarkets}, abci.RequestQuery{})
	suite.Require().NoError(err)
	suite.NotNil(bz)

	expMarkets := []types.Market{
		types.NewMarket("btc:usd", "btc", "usd", []string(nil), true),
		types.NewMarket("xrp:usd", "xrp", "usd", []string(nil), true),
	}

	var p []types.Market
	suite.Nil(types.ModuleCdc.LegacyAmino.UnmarshalJSON(bz, &p))
	suite.Require().Len(p, 2)

	suite.Equal(expMarkets[0], p[0])
}

func (suite *KeeperTestSuite) TestQuerierInvalid() {
	querier := keeper.NewQuerier(suite.keeper, types.ModuleCdc.LegacyAmino)
	bz, err := querier(suite.ctx, []string{"invalidpath"}, abci.RequestQuery{})
	suite.Require().Error(err)
	suite.Nil(bz)
}
