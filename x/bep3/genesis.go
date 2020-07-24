package bep3

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/bep3/types"
)

// InitGenesis initializes the store state from a genesis state.
func InitGenesis(ctx sdk.Context, keeper Keeper, supplyKeeper types.SupplyKeeper, gs GenesisState) {
	// Check if the module account exists
	moduleAcc := supplyKeeper.GetModuleAccount(ctx, ModuleName)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", ModuleName))
	}

	if err := gs.Validate(); err != nil {
		panic(fmt.Sprintf("failed to validate %s genesis state: %s", ModuleName, err))
	}

	keeper.SetParams(ctx, gs.Params)

	for _, swap := range gs.AtomicSwaps {
		if swap.Validate() != nil {
			panic(fmt.Sprintf("invalid swap %s", swap.GetSwapID()))
		}

		// Atomic swap assets must be both supported and active
		err := keeper.ValidateLiveAsset(ctx, swap.Amount[0])
		if err != nil {
			panic(err)
		}

		err = keeper.CreateAtomicSwap(ctx, swap.RandomNumberHash, swap.Timestamp, types.DefaultMaxBlockLock, swap.Sender, swap.Recipient, swap.SenderOtherChain, swap.RecipientOtherChain, swap.Amount, swap.CrossChain)
		if err != nil {
			panic(err)
		}
	}

}

// ExportGenesis writes the current store values to a genesis file, which can be imported again with InitGenesis
func ExportGenesis(ctx sdk.Context, k Keeper) (data GenesisState) {
	params := k.GetParams(ctx)
	swaps := k.GetAllAtomicSwaps(ctx)
	return NewGenesisState(params, swaps)
}
