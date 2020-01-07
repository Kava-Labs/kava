package auction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis - initializes the store state from genesis data
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) {
	keeper.SetNextAuctionID(ctx, data.NextAuctionID)

	keeper.SetParams(ctx, data.Params)

	for _, a := range data.Auctions {
		keeper.SetAuction(ctx, a)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	nextAuctionID, err := keeper.GetNextAuctionID(ctx)
	if err != nil {
		panic(err)
	}

	params := keeper.GetParams(ctx)

	var genAuctions Auctions
	keeper.IterateAuctions(ctx, func(a Auction) bool {
		genAuctions = append(genAuctions, a)
		return false
	})

	return NewGenesisState(nextAuctionID, params, genAuctions)
}
