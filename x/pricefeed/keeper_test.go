package pricefeed

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

// TestKeeper_SetGetAsset tests adding assets to the pricefeed, getting assets from the store
func TestKeeper_SetGetAsset(t *testing.T) {
	helper := getMockApp(t, 0, GenesisState{}, nil)
	header := abci.Header{Height: helper.mApp.LastBlockHeight() + 1}
	helper.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := helper.mApp.BaseApp.NewContext(false, abci.Header{})
	ap := AssetParams{
		Assets: []Asset{Asset{AssetCode: "tst", Description: "the future of finance"}},
	}
	helper.keeper.SetAssetParams(ctx, ap)
	assets := helper.keeper.GetAssets(ctx)
	require.Equal(t, len(assets), 1)
	require.Equal(t, assets[0].AssetCode, "tst")

	_, found := helper.keeper.GetAsset(ctx, "tst")
	require.Equal(t, found, true)

	ap = AssetParams{
		Assets: []Asset{
			Asset{AssetCode: "tst", Description: "the future of finance"},
			Asset{AssetCode: "tst2", Description: "the future of finance"}},
	}
	helper.keeper.SetAssetParams(ctx, ap)
	assets = helper.keeper.GetAssets(ctx)
	require.Equal(t, len(assets), 2)
	require.Equal(t, assets[0].AssetCode, "tst")
	require.Equal(t, assets[1].AssetCode, "tst2")

	_, found = helper.keeper.GetAsset(ctx, "nan")
	require.Equal(t, found, false)
}

// TestKeeper_GetSetPrice Test Posting the price by an oracle
func TestKeeper_GetSetPrice(t *testing.T) {
	helper := getMockApp(t, 2, GenesisState{}, nil)
	header := abci.Header{Height: helper.mApp.LastBlockHeight() + 1}
	helper.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := helper.mApp.BaseApp.NewContext(false, abci.Header{})
	ap := AssetParams{
		Assets: []Asset{Asset{AssetCode: "tst", Description: "the future of finance"}},
	}
	helper.keeper.SetAssetParams(ctx, ap)
	// Set price by oracle 1
	_, err := helper.keeper.SetPrice(
		ctx, helper.addrs[0], "tst",
		sdk.MustNewDecFromStr("0.33"),
		sdk.NewInt(10))
	require.NoError(t, err)
	// Get raw prices
	rawPrices := helper.keeper.GetRawPrices(ctx, "tst")
	require.Equal(t, len(rawPrices), 1)
	require.Equal(t, rawPrices[0].Price.Equal(sdk.MustNewDecFromStr("0.33")), true)
	// Set price by oracle 2
	_, err = helper.keeper.SetPrice(
		ctx, helper.addrs[1], "tst",
		sdk.MustNewDecFromStr("0.35"),
		sdk.NewInt(10))
	require.NoError(t, err)

	rawPrices = helper.keeper.GetRawPrices(ctx, "tst")
	require.Equal(t, len(rawPrices), 2)
	require.Equal(t, rawPrices[1].Price.Equal(sdk.MustNewDecFromStr("0.35")), true)

	// Update Price by Oracle 1
	_, err = helper.keeper.SetPrice(
		ctx, helper.addrs[0], "tst",
		sdk.MustNewDecFromStr("0.37"),
		sdk.NewInt(10))
	require.NoError(t, err)
	rawPrices = helper.keeper.GetRawPrices(ctx, "tst")
	require.Equal(t, rawPrices[0].Price.Equal(sdk.MustNewDecFromStr("0.37")), true)
}

// TestKeeper_GetSetCurrentPrice Test Setting the median price of an Asset
func TestKeeper_GetSetCurrentPrice(t *testing.T) {
	helper := getMockApp(t, 4, GenesisState{}, nil)
	header := abci.Header{Height: helper.mApp.LastBlockHeight() + 1}
	helper.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := helper.mApp.BaseApp.NewContext(false, abci.Header{})
	// Odd number of oracles
	ap := AssetParams{
		Assets: []Asset{Asset{AssetCode: "tst", Description: "the future of finance"}},
	}
	helper.keeper.SetAssetParams(ctx, ap)
	helper.keeper.SetPrice(
		ctx, helper.addrs[0], "tst",
		sdk.MustNewDecFromStr("0.33"),
		sdk.NewInt(10))
	helper.keeper.SetPrice(
		ctx, helper.addrs[1], "tst",
		sdk.MustNewDecFromStr("0.35"),
		sdk.NewInt(10))
	helper.keeper.SetPrice(
		ctx, helper.addrs[2], "tst",
		sdk.MustNewDecFromStr("0.34"),
		sdk.NewInt(10))
	// Set current price
	err := helper.keeper.SetCurrentPrices(ctx)
	require.NoError(t, err)
	// Get Current price
	price := helper.keeper.GetCurrentPrice(ctx, "tst")
	require.Equal(t, price.Price.Equal(sdk.MustNewDecFromStr("0.34")), true)

	// Even number of oracles
	helper.keeper.SetPrice(
		ctx, helper.addrs[3], "tst",
		sdk.MustNewDecFromStr("0.36"),
		sdk.NewInt(10))
	err = helper.keeper.SetCurrentPrices(ctx)
	require.NoError(t, err)
	price = helper.keeper.GetCurrentPrice(ctx, "tst")
	require.Equal(t, price.Price.Equal(sdk.MustNewDecFromStr("0.345")), true)

}
