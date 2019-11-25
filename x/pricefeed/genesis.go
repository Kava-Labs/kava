package pricefeed

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)


// InitGenesis sets distribution information for genesis.
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) {

	// Set the assets and oracles from params
	keeper.SetAssetParams(ctx, data.AssetParams)
	keeper.SetOracleParams(ctx ,data.OracleParams)

	// Iterate through the posted prices and set them in the store
	for _, pp := range data.PostedPrices {
		addr, err := sdk.AccAddressFromBech32(pp.OracleAddress)
		if err != nil {
			panic(err)
		}
		_, err = keeper.SetPrice(ctx, addr, pp.AssetCode, pp.Price, pp.Expiry)
		if err != nil {
			panic(err)
		}
	}

	// Set the current price (if any) based on what's now in the store
	if err := keeper.SetCurrentPrices(ctx); err != nil {
		panic(err)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {

	// Get the params for assets and oracles
	assetParams := keeper.GetAssetParams(ctx)
	oracleParams := keeper.GetOracleParams(ctx)

	var postedPrices []PostedPrice
	for _, asset := range keeper.GetAssets(ctx) {
		pp := keeper.GetRawPrices(ctx, asset.AssetCode)
		postedPrices = append(postedPrices, pp...)
	}

	return GenesisState{
		AssetParams:  assetParams,
		OracleParams: oracleParams,
		PostedPrices: postedPrices,
	}
}
