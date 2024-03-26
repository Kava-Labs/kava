package testutil

import (
	"testing"
	"time"

	tmprototypes "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/pricefeed/keeper"
	"github.com/kava-labs/kava/x/pricefeed/types"
)

func SetCurrentPrices_PriceCalculations(t *testing.T, f func(ctx sdk.Context, keeper keeper.Keeper)) {
	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, tmprototypes.Header{}).
		WithBlockTime(time.Now().UTC())
	keeper := tApp.GetPriceFeedKeeper()

	params := types.Params{
		Markets: []types.Market{
			// valid previous price, expired prices, price change, and active
			{MarketID: "asset1:usd", BaseAsset: "asset1", QuoteAsset: "usd", Oracles: addrs, Active: true},
			// same data as asset1, but not active and should be ignored
			{MarketID: "asset2:usd", BaseAsset: "asset2", QuoteAsset: "usd", Oracles: addrs, Active: false},
			// same data as asset1 except no valid previous price
			{MarketID: "asset3:usd", BaseAsset: "asset3", QuoteAsset: "usd", Oracles: addrs, Active: true},
			// previous price set, but no valid prices
			{MarketID: "asset4:usd", BaseAsset: "asset4", QuoteAsset: "usd", Oracles: addrs, Active: true},
			// same as market one except different prices
			{MarketID: "asset5:usd", BaseAsset: "asset5", QuoteAsset: "usd", Oracles: addrs, Active: true},
		},
	}
	keeper.SetParams(ctx, params)

	// need price equal to block time and after block time
	blockTime := time.Now()
	initialPriceExpiry := blockTime.Add(1 * time.Hour)

	_, err := keeper.SetPrice(ctx, addrs[0], "asset1:usd", sdk.MustNewDecFromStr("1"), initialPriceExpiry)
	require.NoError(t, err)
	_, err = keeper.SetPrice(ctx, addrs[0], "asset2:usd", sdk.MustNewDecFromStr("1"), initialPriceExpiry)
	require.NoError(t, err)
	_, err = keeper.SetPrice(ctx, addrs[0], "asset4:usd", sdk.MustNewDecFromStr("1"), initialPriceExpiry)
	require.NoError(t, err)
	_, err = keeper.SetPrice(ctx, addrs[0], "asset5:usd", sdk.MustNewDecFromStr("10"), initialPriceExpiry)
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(blockTime)
	f(ctx, keeper)

	// price should be set
	price, err := keeper.GetCurrentPrice(ctx, "asset1:usd")
	require.NoError(t, err)
	require.Equal(t, sdk.OneDec(), price.Price)
	// not an active market, so price is not set
	price, err = keeper.GetCurrentPrice(ctx, "asset2:usd")
	require.Equal(t, types.ErrNoValidPrice, err)
	// no price posted
	price, err = keeper.GetCurrentPrice(ctx, "asset3:usd")
	require.Equal(t, types.ErrNoValidPrice, err)
	// price set initially
	price, err = keeper.GetCurrentPrice(ctx, "asset4:usd")
	require.NoError(t, err)
	require.Equal(t, sdk.OneDec(), price.Price)
	price, err = keeper.GetCurrentPrice(ctx, "asset5:usd")
	require.NoError(t, err)
	require.Equal(t, sdk.MustNewDecFromStr("10.0"), price.Price)

	_, err = keeper.SetPrice(ctx, addrs[1], "asset1:usd", sdk.MustNewDecFromStr("2"), initialPriceExpiry.Add(1*time.Hour))
	require.NoError(t, err)
	_, err = keeper.SetPrice(ctx, addrs[1], "asset2:usd", sdk.MustNewDecFromStr("2"), initialPriceExpiry.Add(1*time.Hour))
	require.NoError(t, err)
	_, err = keeper.SetPrice(ctx, addrs[1], "asset5:usd", sdk.MustNewDecFromStr("20"), initialPriceExpiry.Add(1*time.Hour))
	require.NoError(t, err)

	blockTime = blockTime.Add(30 * time.Minute)
	ctx = ctx.WithBlockTime(blockTime)
	f(ctx, keeper)

	// price should be set
	price, err = keeper.GetCurrentPrice(ctx, "asset1:usd")
	require.NoError(t, err)
	require.Equal(t, sdk.MustNewDecFromStr("1.5"), price.Price)
	// not an active market, so price is not set
	price, err = keeper.GetCurrentPrice(ctx, "asset2:usd")
	require.Equal(t, types.ErrNoValidPrice, err)
	// no price posted
	price, err = keeper.GetCurrentPrice(ctx, "asset3:usd")
	require.Equal(t, types.ErrNoValidPrice, err)
	// price set initially
	price, err = keeper.GetCurrentPrice(ctx, "asset4:usd")
	require.NoError(t, err)
	require.Equal(t, sdk.OneDec(), price.Price)
	price, err = keeper.GetCurrentPrice(ctx, "asset5:usd")
	require.NoError(t, err)
	require.Equal(t, sdk.MustNewDecFromStr("15.0"), price.Price)

	_, err = keeper.SetPrice(ctx, addrs[2], "asset1:usd", sdk.MustNewDecFromStr("30"), initialPriceExpiry.Add(1*time.Hour))
	require.NoError(t, err)
	_, err = keeper.SetPrice(ctx, addrs[2], "asset2:usd", sdk.MustNewDecFromStr("30"), initialPriceExpiry.Add(1*time.Hour))
	require.NoError(t, err)
	_, err = keeper.SetPrice(ctx, addrs[2], "asset5:usd", sdk.MustNewDecFromStr("30"), initialPriceExpiry.Add(1*time.Hour))
	require.NoError(t, err)

	blockTime = blockTime.Add(15 * time.Minute)
	ctx = ctx.WithBlockTime(blockTime)
	f(ctx, keeper)

	// price should be set
	price, err = keeper.GetCurrentPrice(ctx, "asset1:usd")
	require.NoError(t, err)
	require.Equal(t, sdk.MustNewDecFromStr("2.0"), price.Price)
	// not an active market, so price is not set
	price, err = keeper.GetCurrentPrice(ctx, "asset2:usd")
	require.Equal(t, types.ErrNoValidPrice, err)
	// no price posted
	price, err = keeper.GetCurrentPrice(ctx, "asset3:usd")
	require.Equal(t, types.ErrNoValidPrice, err)
	// price set initially
	price, err = keeper.GetCurrentPrice(ctx, "asset4:usd")
	require.NoError(t, err)
	require.Equal(t, sdk.OneDec(), price.Price)
	price, err = keeper.GetCurrentPrice(ctx, "asset5:usd")
	require.NoError(t, err)
	require.Equal(t, sdk.MustNewDecFromStr("20.0"), price.Price)

	blockTime = blockTime.Add(15 * time.Minute)
	ctx = ctx.WithBlockTime(blockTime)
	f(ctx, keeper)

	// price should be set
	price, err = keeper.GetCurrentPrice(ctx, "asset1:usd")
	require.NoError(t, err)
	require.Equal(t, sdk.MustNewDecFromStr("16"), price.Price)
	// not an active market, so price is not set
	price, err = keeper.GetCurrentPrice(ctx, "asset2:usd")
	require.Equal(t, types.ErrNoValidPrice, err)
	// no price posted
	price, err = keeper.GetCurrentPrice(ctx, "asset3:usd")
	require.Equal(t, types.ErrNoValidPrice, err)
	// price set initially, now expired
	price, err = keeper.GetCurrentPrice(ctx, "asset4:usd")
	require.Equal(t, types.ErrNoValidPrice, err)
	price, err = keeper.GetCurrentPrice(ctx, "asset5:usd")
	require.NoError(t, err)
	require.Equal(t, sdk.MustNewDecFromStr("25.0"), price.Price)

	blockTime = blockTime.Add(10 * time.Hour)
	ctx = ctx.WithBlockTime(blockTime)
	f(ctx, keeper)

	// all prices expired now
	price, err = keeper.GetCurrentPrice(ctx, "asset1:usd")
	require.Equal(t, types.ErrNoValidPrice, err)
	price, err = keeper.GetCurrentPrice(ctx, "asset2:usd")
	require.Equal(t, types.ErrNoValidPrice, err)
	price, err = keeper.GetCurrentPrice(ctx, "asset3:usd")
	require.Equal(t, types.ErrNoValidPrice, err)
	price, err = keeper.GetCurrentPrice(ctx, "asset4:usd")
	require.Equal(t, types.ErrNoValidPrice, err)
	price, err = keeper.GetCurrentPrice(ctx, "asset5:usd")
	require.Equal(t, types.ErrNoValidPrice, err)
}

func SetCurrentPrices_EventEmission(t *testing.T, f func(ctx sdk.Context, keeper keeper.Keeper)) {
	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, tmprototypes.Header{}).
		WithBlockTime(time.Now().UTC())
	keeper := tApp.GetPriceFeedKeeper()

	params := types.Params{
		Markets: []types.Market{
			{MarketID: "asset1:usd", BaseAsset: "asset1", QuoteAsset: "usd", Oracles: addrs, Active: true},
		},
	}
	keeper.SetParams(ctx, params)

	blockTime := time.Now()
	initialPriceExpiry := blockTime.Add(1 * time.Hour)

	// post a price
	_, err := keeper.SetPrice(ctx, addrs[0], "asset1:usd", sdk.MustNewDecFromStr("1"), initialPriceExpiry)
	require.NoError(t, err)

	// reset context with fresh event manager
	ctx = ctx.WithBlockTime(blockTime).WithEventManager(sdk.NewEventManager())
	f(ctx, keeper)

	// no previous price so no event
	require.Equal(t, 0, len(ctx.EventManager().Events()))

	// post same price from another oracle
	_, err = keeper.SetPrice(ctx, addrs[1], "asset1:usd", sdk.MustNewDecFromStr("1"), initialPriceExpiry)
	require.NoError(t, err)

	blockTime = blockTime.Add(10 * time.Second)
	ctx = ctx.WithBlockTime(blockTime).WithEventManager(sdk.NewEventManager())
	f(ctx, keeper)

	// no price change so no event
	require.Equal(t, 0, len(ctx.EventManager().Events()))

	// post price changes
	_, err = keeper.SetPrice(ctx, addrs[2], "asset1:usd", sdk.MustNewDecFromStr("2"), initialPriceExpiry)
	require.NoError(t, err)
	_, err = keeper.SetPrice(ctx, addrs[3], "asset1:usd", sdk.MustNewDecFromStr("10"), initialPriceExpiry)
	require.NoError(t, err)
	_, err = keeper.SetPrice(ctx, addrs[4], "asset1:usd", sdk.MustNewDecFromStr("10"), initialPriceExpiry)
	require.NoError(t, err)

	blockTime = blockTime.Add(10 * time.Second)
	ctx = ctx.WithBlockTime(blockTime).WithEventManager(sdk.NewEventManager())
	f(ctx, keeper)

	// price is changes so event should be emitted
	require.Equal(t, 1, len(ctx.EventManager().Events()))

	event := ctx.EventManager().Events()[0]

	// has correct event type
	assert.Equal(t, types.EventTypeMarketPriceUpdated, event.Type)
	// has correct attributes
	marketID, found := event.GetAttribute(types.AttributeMarketID)
	require.True(t, found)
	marketPrice, found := event.GetAttribute(types.AttributeMarketPrice)
	require.True(t, found)
	// attributes have correct values
	assert.Equal(t, "asset1:usd", marketID.Value)
	assert.Equal(t, sdk.MustNewDecFromStr("2").String(), marketPrice.Value)
}
