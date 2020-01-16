package pricefeed

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis sets distribution information for genesis.
func InitGenesis(ctx sdk.Context, keeper Keeper, gs GenesisState) {
	// Set the markets and oracles from params
	keeper.SetParams(ctx, gs.Params)

	// Iterate through the posted prices and set them in the store
	for _, pp := range gs.PostedPrices {
		_, err := keeper.SetPrice(ctx, pp.OracleAddress, pp.MarketID, pp.Price, pp.Expiry)
		if err != nil {
			panic(err)
		}
	}
	params := keeper.GetParams(ctx)

	// Set the current price (if any) based on what's now in the store
	for _, market := range params.Markets {
		if market.Active {
			err := keeper.SetCurrentPrices(ctx, market.MarketID)
			if err != nil {
				panic(err)
			}
		}
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {

	// Get the params for markets and oracles
	params := keeper.GetParams(ctx)

	var postedPrices []PostedPrice
	for _, market := range keeper.GetMarkets(ctx) {
		pp := keeper.GetRawPrices(ctx, market.MarketID)
		postedPrices = append(postedPrices, pp...)
	}

	return GenesisState{
		Params:       params,
		PostedPrices: postedPrices,
	}
}
