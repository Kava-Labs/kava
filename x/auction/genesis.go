package auction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis - initializes the store state from genesis data
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) {
	keeper.SetParams(ctx, data.AuctionParams)

	for _, a := range data.Auctions {
		keeper.SetAuction(ctx, a)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	params := keeper.GetParams(ctx)

	var genAuctions GenesisAuctions
	iterator := keeper.GetAuctionIterator(ctx)

	for ; iterator.Valid(); iterator.Next() {

		auction := keeper.DecodeAuction(ctx, iterator.Value())
		genAuctions = append(genAuctions, auction)

	}
	return NewGenesisState(params, genAuctions)
}
