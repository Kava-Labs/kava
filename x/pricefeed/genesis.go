package pricefeed

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/pricefeed/keeper"
	"github.com/kava-labs/kava/x/pricefeed/types"
)

// InitGenesis sets distribution information for genesis.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, gs types.GenesisState) {
	// Set the markets and oracles from params
	k.SetParams(ctx, gs.Params)

	// Iterate through the posted prices and set them in the store if they are not expired
	for _, pp := range gs.PostedPrices {
		if pp.Expiry.After(ctx.BlockTime()) {
			_, err := k.SetPrice(ctx, pp.OracleAddress, pp.MarketID, pp.Price, pp.Expiry)
			if err != nil {
				panic(err)
			}
		}
	}
	params := k.GetParams(ctx)

	// Set the current price (if any) based on what's now in the store
	for _, market := range params.Markets {
		if !market.Active {
			continue
		}
		rps := k.GetRawPrices(ctx, market.MarketID)

		if len(rps) == 0 {
			continue
		}
		err := k.SetCurrentPrices(ctx, market.MarketID)
		if err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {

	// Get the params for markets and oracles
	params := k.GetParams(ctx)

	var postedPrices []types.PostedPrice
	for _, market := range k.GetMarkets(ctx) {
		pp := k.GetRawPrices(ctx, market.MarketID)
		postedPrices = append(postedPrices, pp...)
	}

	return types.NewGenesisState(params, postedPrices)
}
