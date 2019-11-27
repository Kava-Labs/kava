package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/x/pricefeed/types"
)

// TestKeeper_SetGetAsset tests adding assets to the pricefeed, getting assets from the store
func TestKeeper_SetGetAsset(t *testing.T) {
	helper := getMockApp(t, 0, types.GenesisState{}, nil)
	header := abci.Header{
		Height: helper.mApp.LastBlockHeight() + 1,
		Time:   tmtime.Now()}
	helper.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := helper.mApp.BaseApp.NewContext(false, header)

	ap := types.Params{
		Assets: []types.Asset{
			types.Asset{AssetCode: "tstusd", BaseAsset: "tst", QuoteAsset: "usd", Oracles: types.Oracles{}, Active: true},
		},
	}
	helper.keeper.SetParams(ctx, ap)
	assets := helper.keeper.GetAssetParams(ctx)
	require.Equal(t, len(assets), 1)
	require.Equal(t, assets[0].AssetCode, "tstusd")

	_, found := helper.keeper.GetAsset(ctx, "tstusd")
	require.Equal(t, found, true)

	ap = types.Params{
		Assets: []types.Asset{
			types.Asset{AssetCode: "tstusd", BaseAsset: "tst", QuoteAsset: "usd", Oracles: types.Oracles{}, Active: true},
			types.Asset{AssetCode: "tst2usd", BaseAsset: "tst2", QuoteAsset: "usd", Oracles: types.Oracles{}, Active: true},
		},
	}
	helper.keeper.SetParams(ctx, ap)
	assets = helper.keeper.GetAssetParams(ctx)
	require.Equal(t, len(assets), 2)
	require.Equal(t, assets[0].AssetCode, "tstusd")
	require.Equal(t, assets[1].AssetCode, "tst2usd")

	_, found = helper.keeper.GetAsset(ctx, "nan")
	require.Equal(t, found, false)
}

// TestKeeper_GetSetPrice Test Posting the price by an oracle
func TestKeeper_GetSetPrice(t *testing.T) {
	helper := getMockApp(t, 2, types.GenesisState{}, nil)
	header := abci.Header{
		Height: helper.mApp.LastBlockHeight() + 1,
		Time:   tmtime.Now()}
	helper.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := helper.mApp.BaseApp.NewContext(false, header)
	ap := types.Params{
		Assets: []types.Asset{
			types.Asset{AssetCode: "tstusd", BaseAsset: "tst", QuoteAsset: "usd", Oracles: types.Oracles{}, Active: true},
		},
	}
	helper.keeper.SetParams(ctx, ap)
	// Set price by oracle 1
	_, err := helper.keeper.SetPrice(
		ctx, helper.addrs[0], "tstusd",
		sdk.MustNewDecFromStr("0.33"),
		header.Time.Add(1*time.Hour))
	require.NoError(t, err)
	// Get raw prices
	rawPrices := helper.keeper.GetRawPrices(ctx, "tstusd")
	require.Equal(t, len(rawPrices), 1)
	require.Equal(t, rawPrices[0].Price.Equal(sdk.MustNewDecFromStr("0.33")), true)
	// Set price by oracle 2
	_, err = helper.keeper.SetPrice(
		ctx, helper.addrs[1], "tstusd",
		sdk.MustNewDecFromStr("0.35"),
		header.Time.Add(time.Hour*1))
	require.NoError(t, err)

	rawPrices = helper.keeper.GetRawPrices(ctx, "tstusd")
	require.Equal(t, len(rawPrices), 2)
	require.Equal(t, rawPrices[1].Price.Equal(sdk.MustNewDecFromStr("0.35")), true)

	// Update Price by Oracle 1
	_, err = helper.keeper.SetPrice(
		ctx, helper.addrs[0], "tstusd",
		sdk.MustNewDecFromStr("0.37"),
		header.Time.Add(time.Hour*1))
	require.NoError(t, err)
	rawPrices = helper.keeper.GetRawPrices(ctx, "tstusd")
	require.Equal(t, rawPrices[0].Price.Equal(sdk.MustNewDecFromStr("0.37")), true)
}

// TestKeeper_GetSetCurrentPrice Test Setting the median price of an Asset
func TestKeeper_GetSetCurrentPrice(t *testing.T) {
	helper := getMockApp(t, 4, types.GenesisState{}, nil)
	header := abci.Header{
		Height: helper.mApp.LastBlockHeight() + 1,
		Time:   tmtime.Now()}
	helper.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := helper.mApp.BaseApp.NewContext(false, header)
	ap := types.Params{
		Assets: []types.Asset{
			types.Asset{AssetCode: "tstusd", BaseAsset: "tst", QuoteAsset: "usd", Oracles: types.Oracles{}, Active: true},
		},
	}
	helper.keeper.SetParams(ctx, ap)
	helper.keeper.SetPrice(
		ctx, helper.addrs[0], "tstusd",
		sdk.MustNewDecFromStr("0.33"),
		header.Time.Add(time.Hour*1))
	helper.keeper.SetPrice(
		ctx, helper.addrs[1], "tstusd",
		sdk.MustNewDecFromStr("0.35"),
		header.Time.Add(time.Hour*1))
	helper.keeper.SetPrice(
		ctx, helper.addrs[2], "tstusd",
		sdk.MustNewDecFromStr("0.34"),
		header.Time.Add(time.Hour*1))
	// Set current price
	err := helper.keeper.SetCurrentPrices(ctx, "tstusd")
	require.NoError(t, err)
	// Get Current price
	price := helper.keeper.GetCurrentPrice(ctx, "tstusd")
	require.Equal(t, price.Price.Equal(sdk.MustNewDecFromStr("0.34")), true)

	// Even number of oracles
	helper.keeper.SetPrice(
		ctx, helper.addrs[3], "tstusd",
		sdk.MustNewDecFromStr("0.36"),
		header.Time.Add(time.Hour*1))
	err = helper.keeper.SetCurrentPrices(ctx, "tstusd")
	require.NoError(t, err)
	price = helper.keeper.GetCurrentPrice(ctx, "tstusd")
	require.Equal(t, price.Price.Equal(sdk.MustNewDecFromStr("0.345")), true)

}
