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

	totalAssetSupplyCoins := sdk.NewCoins()
	for _, asset := range gs.Assets {
		keeper.SetAssetSupply(ctx, asset, []byte(asset.Denom))
		totalAssetSupplyCoins.Add(sdk.NewCoins(asset))
	}

	totalAtomicSwapCoins := sdk.NewCoins()
	for _, genSwap := range gs.AtomicSwaps {
		swap, ok := genSwap.(types.AtomicSwap)
		if !ok {
			panic("could not convert stored GenesisAtomicSwap to AtomicSwap type")
		}
		keeper.SetAtomicSwap(ctx, swap, swap.GetSwapID())
		totalAtomicSwapCoins.Add(swap.GetModuleAccountCoins())
	}

	keeper.SetParams(ctx, gs.Params)

	totalGenesisCoins := totalAssetSupplyCoins.Add(totalAtomicSwapCoins)

	// Check module coins match expected genesis coins
	moduleAcc := supplyKeeper.GetModuleAccount(ctx, ModuleName)
	if !moduleAcc.GetCoins().IsEqual(totalGenesisCoins) {
		panic(fmt.Sprintf("total coins (%s) do not equal (%s) module account (%s) ", moduleAcc.GetCoins(), ModuleName, totalGenesisCoins))
	}
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(ctx sdk.Context, k Keeper) (data types.GenesisState) {
	params := k.GetParams(ctx)

	genAssetSupplies := []sdk.Coin{}
	// TODO: Add k.IterateAssetSupplies

	genAtomicSwaps := types.GenesisAtomicSwaps{}
	// TODO:
	// k.IterateAtomicSwaps(ctx, func(a types.Swap) bool {
	// 	genAtomicSwap, ok := a.(types.GenesisAtomicSwap)
	// 	if !ok {
	// 		panic("could not convert stored AtomicSwap to GenesisAtomicSwap type")
	// 	}
	// 	genAtomicSwaps = append(genAtomicSwaps, genAtomicSwap)
	// 	return false
	// })

	return types.NewGenesisState(params, genAtomicSwaps, genAssetSupplies)
}
