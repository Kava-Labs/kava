package auction

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/auction/keeper"
	"github.com/kava-labs/kava/x/auction/types"
)

// InitGenesis initializes the store state from a genesis state.
func InitGenesis(ctx sdk.Context, keeper keeper.Keeper, bankKeeper types.BankKeeper, accountKeeper types.AccountKeeper, gs *types.GenesisState) {
	if err := gs.Validate(); err != nil {
		panic(fmt.Sprintf("failed to validate %s genesis state: %s", types.ModuleName, err))
	}

	keeper.SetNextAuctionID(ctx, gs.NextAuctionId)

	keeper.SetParams(ctx, gs.Params)

	totalAuctionCoins := sdk.NewCoins()

	auctions, err := types.UnpackGenesisAuctions(gs.Auctions)
	if err != nil {
		panic(fmt.Sprintf("failed to unpack genesis auctions: %s", err))
	}
	for _, a := range auctions {
		keeper.SetAuction(ctx, a)
		// find the total coins that should be present in the module account
		totalAuctionCoins = totalAuctionCoins.Add(a.GetModuleAccountCoins()...)
	}

	// check if the module account exists
	moduleAcc := accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	maccCoins := bankKeeper.GetAllBalances(ctx, moduleAcc.GetAddress())

	// check module coins match auction coins
	// Note: Other sdk modules do not check this, instead just using the existing module account coins, or if zero, setting them.
	if !maccCoins.IsEqual(totalAuctionCoins) {
		panic(fmt.Sprintf("total auction coins (%s) do not equal (%s) module account (%s) ", maccCoins, types.ModuleName, totalAuctionCoins))
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper keeper.Keeper) *types.GenesisState {
	nextAuctionID, err := keeper.GetNextAuctionID(ctx)
	if err != nil {
		panic(err)
	}

	params := keeper.GetParams(ctx)

	genAuctions := []types.GenesisAuction{} // return empty list instead of nil if no auctions
	keeper.IterateAuctions(ctx, func(a types.Auction) bool {
		ga, ok := a.(types.GenesisAuction)
		if !ok {
			panic("could not convert stored auction to GenesisAuction type")
		}
		genAuctions = append(genAuctions, ga)
		return false
	})

	gs, err := types.NewGenesisState(nextAuctionID, params, genAuctions)
	if err != nil {
		panic(err)
	}

	return gs
}
