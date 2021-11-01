package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	tmprototypes "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/pricefeed/types"
)

// TestKeeper_SetGetMarket tests adding markets to the pricefeed, getting markets from the store
func TestKeeper_SetGetMarket(t *testing.T) {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, tmprototypes.Header{})
	keeper := tApp.GetPriceFeedKeeper()

	mp := types.Params{
		Markets: []types.Market{
			{MarketID: "tstusd", BaseAsset: "tst", QuoteAsset: "usd", Oracles: []string{}, Active: true},
		},
	}
	keeper.SetParams(ctx, mp)
	markets := keeper.GetMarkets(ctx)
	require.Equal(t, len(markets), 1)
	require.Equal(t, markets[0].MarketID, "tstusd")

	_, found := keeper.GetMarket(ctx, "tstusd")
	require.Equal(t, found, true)

	mp = types.Params{
		Markets: []types.Market{
			{MarketID: "tstusd", BaseAsset: "tst", QuoteAsset: "usd", Oracles: []string{}, Active: true},
			{MarketID: "tst2usd", BaseAsset: "tst2", QuoteAsset: "usd", Oracles: []string{}, Active: true},
		},
	}
	keeper.SetParams(ctx, mp)
	markets = keeper.GetMarkets(ctx)
	require.Equal(t, len(markets), 2)
	require.Equal(t, markets[0].MarketID, "tstusd")
	require.Equal(t, markets[1].MarketID, "tst2usd")

	_, found = keeper.GetMarket(ctx, "nan")
	require.Equal(t, found, false)
}

// TestKeeper_GetSetPrice Test Posting the price by an oracle
func TestKeeper_GetSetPrice(t *testing.T) {
	_, addrs := app.GeneratePrivKeyAddressPairs(2)
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, tmprototypes.Header{})
	keeper := tApp.GetPriceFeedKeeper()

	mp := types.Params{
		Markets: []types.Market{
			{MarketID: "tstusd", BaseAsset: "tst", QuoteAsset: "usd", Oracles: []string{}, Active: true},
		},
	}
	keeper.SetParams(ctx, mp)

	prices := []struct {
		oracle   string
		MarketID string
		price    sdk.Dec
		total    int
	}{
		{addrs[0].String(), "tstusd", sdk.MustNewDecFromStr("0.33"), 1},
		{addrs[1].String(), "tstusd", sdk.MustNewDecFromStr("0.35"), 2},
		{addrs[0].String(), "tstusd", sdk.MustNewDecFromStr("0.37"), 2},
	}

	for _, p := range prices {
		// Set price by oracle 1
		pp, err := keeper.SetPrice(
			ctx,
			p.oracle,
			p.MarketID,
			p.price,
			time.Now().UTC().Add(1*time.Hour),
		)

		require.NoError(t, err)

		// Get raw prices
		rawPrices := keeper.GetRawPrices(ctx, "tstusd")

		require.Equal(t, p.total, len(rawPrices))
		require.Contains(t, rawPrices, pp)

		// Find the oracle and require price to be same
		for _, rp := range rawPrices {
			if p.oracle == rp.OracleAddress {
				require.Equal(t, p.price, rp.Price)
			}
		}
	}
}

// TestKeeper_GetSetCurrentPrice Test Setting the median price of an Asset
func TestKeeper_GetSetCurrentPrice(t *testing.T) {
	_, addrs := app.GeneratePrivKeyAddressPairs(4)
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, tmprototypes.Header{})
	keeper := tApp.GetPriceFeedKeeper()

	mp := types.Params{
		Markets: []types.Market{
			{MarketID: "tstusd", BaseAsset: "tst", QuoteAsset: "usd", Oracles: []string{}, Active: true},
		},
	}
	keeper.SetParams(ctx, mp)
	_, err := keeper.SetPrice(
		ctx, addrs[0].String(), "tstusd",
		sdk.MustNewDecFromStr("0.33"),
		time.Now().Add(time.Hour*1))
	require.NoError(t, err)
	_, err = keeper.SetPrice(
		ctx, addrs[1].String(), "tstusd",
		sdk.MustNewDecFromStr("0.35"),
		time.Now().Add(time.Hour*1))
	require.NoError(t, err)
	_, err = keeper.SetPrice(
		ctx, addrs[2].String(), "tstusd",
		sdk.MustNewDecFromStr("0.34"),
		time.Now().Add(time.Hour*1))
	require.NoError(t, err)
	// Set current price
	err = keeper.SetCurrentPrices(ctx, "tstusd")
	require.NoError(t, err)
	// Get Current price
	price, err := keeper.GetCurrentPrice(ctx, "tstusd")
	require.Nil(t, err)
	require.Equal(t, price.Price.Equal(sdk.MustNewDecFromStr("0.34")), true)

	// Even number of oracles
	_, err = keeper.SetPrice(
		ctx, addrs[3].String(), "tstusd",
		sdk.MustNewDecFromStr("0.36"),
		time.Now().Add(time.Hour*1))
	require.NoError(t, err)
	err = keeper.SetCurrentPrices(ctx, "tstusd")
	require.NoError(t, err)
	price, err = keeper.GetCurrentPrice(ctx, "tstusd")
	require.Nil(t, err)

	exp := sdk.MustNewDecFromStr("0.345")
	require.Truef(t, price.Price.Equal(exp),
		"current price %s should be %s",
		price.Price.String(),
		exp.String(),
	)
}
