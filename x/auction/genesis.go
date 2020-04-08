package auction

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/auction/types"
)

// InitGenesis initializes the store state from a genesis state.
func InitGenesis(ctx sdk.Context, keeper Keeper, supplyKeeper types.SupplyKeeper, gs GenesisState) {
	if err := gs.Validate(); err != nil {
		panic(fmt.Sprintf("failed to validate %s genesis state: %s", ModuleName, err))
	}

	keeper.SetNextAuctionID(ctx, gs.NextAuctionID)

	keeper.SetParams(ctx, gs.Params)

	totalAuctionCoins := sdk.NewCoins()
	for _, a := range gs.Auctions {
		keeper.SetAuction(ctx, a)
		// find the total coins that should be present in the module account
		totalAuctionCoins.Add(a.GetModuleAccountCoins()...)
	}

	// check if the module account exists
	moduleAcc := supplyKeeper.GetModuleAccount(ctx, ModuleName)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", ModuleName))
	}
	// check module coins match auction coins
	// Note: Other sdk modules do not check this, instead just using the existing module account coins, or if zero, setting them.
	if !moduleAcc.GetCoins().IsEqual(totalAuctionCoins) {
		panic(fmt.Sprintf("total auction coins (%s) do not equal (%s) module account (%s) ", moduleAcc.GetCoins(), ModuleName, totalAuctionCoins))
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	nextAuctionID, err := keeper.GetNextAuctionID(ctx)
	if err != nil {
		panic(err)
	}

	params := keeper.GetParams(ctx)

	genAuctions := GenesisAuctions{} // return empty list instead of nil if no auctions
	keeper.IterateAuctions(ctx, func(a Auction) bool {
		ga, ok := a.(types.GenesisAuction)
		if !ok {
			panic("could not convert stored auction to GenesisAuction type")
		}
		genAuctions = append(genAuctions, ga)
		return false
	})

	return NewGenesisState(nextAuctionID, params, genAuctions)
}
