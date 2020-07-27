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

	var incomingSupplies sdk.Coins
	var outgoingSupplies sdk.Coins
	for _, swap := range gs.AtomicSwaps {
		if swap.Validate() != nil {
			panic(fmt.Sprintf("invalid swap %s", swap.GetSwapID()))
		}

		// Atomic swap assets must be both supported and active
		err := keeper.ValidateLiveAsset(ctx, swap.Amount[0])
		if err != nil {
			panic(err)
		}

		keeper.SetAtomicSwap(ctx, swap)

		// Add swap to block index or longterm storage based on swap.Status
		// Increment incoming or outgoing supply based on swap.Direction
		switch swap.Direction {
		case Incoming:
			switch swap.Status {
			case Open:
				// This index expires unclaimed swaps
				keeper.InsertIntoByBlockIndex(ctx, swap)
				incomingSupplies = incomingSupplies.Add(swap.Amount...)
			case Expired:
				incomingSupplies = incomingSupplies.Add(swap.Amount...)
			case Completed:
				// This index stores swaps until deletion
				keeper.InsertIntoLongtermStorage(ctx, swap)
			default:
				panic(fmt.Sprintf("swap %s has invalid status %s", swap.GetSwapID(), swap.Status.String()))
			}
		case Outgoing:
			switch swap.Status {
			case Open:
				keeper.InsertIntoByBlockIndex(ctx, swap)
				outgoingSupplies = outgoingSupplies.Add(swap.Amount...)
			case Expired:
				outgoingSupplies = outgoingSupplies.Add(swap.Amount...)
			case Completed:
				keeper.InsertIntoLongtermStorage(ctx, swap)
			default:
				panic(fmt.Sprintf("swap %s has invalid status %s", swap.GetSwapID(), swap.Status.String()))
			}
		default:
			panic(fmt.Sprintf("swap %s has invalid direction %s", swap.GetSwapID(), swap.Direction.String()))
		}
	}

	// Asset's given incoming/outgoing supply much match the amount of coins in incoming/outgoing atomic swaps
	assets, _ := keeper.GetAssets(ctx)
	for _, asset := range assets {
		incomingSupply := incomingSupplies.AmountOf(asset.Denom)
		if !asset.SupplyLimit.IncomingSupply.Amount.Equal(incomingSupply) {
			panic(fmt.Sprintf("asset's incoming supply %s does not match amount %s in incoming atomic swaps",
				asset.SupplyLimit.IncomingSupply, incomingSupply))
		}
		outgoingSupply := outgoingSupplies.AmountOf(asset.Denom)
		if !asset.SupplyLimit.OutgoingSupply.Amount.Equal(outgoingSupply) {
			panic(fmt.Sprintf("asset's outgoing supply %s does not match amount %s in outgoing atomic swaps",
				asset.SupplyLimit.OutgoingSupply, outgoingSupply))
		}
	}
}

// ExportGenesis writes the current store values to a genesis file, which can be imported again with InitGenesis
func ExportGenesis(ctx sdk.Context, k Keeper) (data GenesisState) {
	params := k.GetParams(ctx)
	swaps := k.GetAllAtomicSwaps(ctx)
	return NewGenesisState(params, swaps)
}
