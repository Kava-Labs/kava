package pricefeed

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis sets distribution information for genesis.
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) {

	// Set the assets and oracles from params
	keeper.SetParams(ctx, data.Params)

	// Iterate through the posted prices and set them in the store
	for _, pp := range data.PostedPrices {
		_, err := keeper.SetPrice(ctx, pp.OracleAddress, pp.MarketID, pp.Price, pp.Expiry)
		if err != nil {
			panic(err)
		}
	}

	// Set the current price (if any) based on what's now in the store
	for _, a := range data.Params.Markets {
		if a.Active {
			err := keeper.SetCurrentPrices(ctx, a.MarketID)
			if err != nil {
				panic(err)
			}
		}
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {

	// Get the params for assets and oracles
	params := keeper.GetParams(ctx)

	var postedPrices []PostedPrice
	for _, asset := range keeper.GetMarketParams(ctx) {
		pp := keeper.GetRawPrices(ctx, asset.MarketID)
		postedPrices = append(postedPrices, pp...)
	}

	return GenesisState{
		Params:       params,
		PostedPrices: postedPrices,
	}
}
