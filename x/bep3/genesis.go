package bep3

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/bep3/types"
)

// InitGenesis initializes the store state from a genesis state.
func InitGenesis(ctx sdk.Context, keeper Keeper, supplyKeeper types.SupplyKeeper, gs GenesisState) {
	if err := gs.Validate(); err != nil {
		panic(fmt.Sprintf("failed to validate %s genesis state: %s", ModuleName, err))
	}

	// Set each AssetSupply and store total coin count
	totalGenesisCoins := sdk.NewCoins()
	for _, asset := range gs.AssetSupplies {
		keeper.SetAssetSupply(ctx, asset, []byte(asset.Denom))
		totalGenesisCoins.Add(sdk.NewCoins(asset))
	}

	// Set each AtomicSwap and store total coin count
	for _, swap := range gs.AtomicSwaps {
		if swap.Validate() != nil {
			panic(fmt.Sprintf("invalid swap %s", swap.GetSwapID()))
		}
		keeper.SetAtomicSwap(ctx, swap)
		keeper.InsertIntoByBlockIndex(ctx, swap)
		totalGenesisCoins.Add(swap.GetModuleAccountCoins())
	}

	keeper.SetParams(ctx, gs.Params)

	// Check module coins match expected genesis coins
	moduleAcc := supplyKeeper.GetModuleAccount(ctx, ModuleName)
	if !moduleAcc.GetCoins().IsEqual(totalGenesisCoins) {
		panic(fmt.Sprintf("total coins (%s) do not equal (%s) module account (%s) ", moduleAcc.GetCoins(), ModuleName, totalGenesisCoins))
	}
}

// ExportGenesis writes the current store values to a genesis file, which can be imported again with InitGenesis
func ExportGenesis(ctx sdk.Context, k Keeper) (data types.GenesisState) {
	params := k.GetParams(ctx)
	swaps := k.GetAllAtomicSwaps(ctx)
	assets := k.GetAllAssetSupplies(ctx)

	return types.NewGenesisState(params, swaps, assets)
}
