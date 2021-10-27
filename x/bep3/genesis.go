package bep3

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/bep3/keeper"
	"github.com/kava-labs/kava/x/bep3/types"
)

// InitGenesis initializes the store state from a genesis state.
func InitGenesis(ctx sdk.Context, keeper keeper.Keeper, supplyKeeper types.SupplyKeeper, gs types.GenesisState) {
	// Check if the module account exists
	moduleAcc := supplyKeeper.GetModuleAccount(ctx, types.ModuleName)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	if err := gs.Validate(); err != nil {
		panic(fmt.Sprintf("failed to validate %s genesis state: %s", types.ModuleName, err))
	}

	keeper.SetPreviousBlockTime(ctx, gs.PreviousBlockTime)

	keeper.SetParams(ctx, gs.Params)
	for _, supply := range gs.Supplies {
		keeper.SetAssetSupply(ctx, supply, supply.GetDenom())
	}

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
		case types.SWAP_DIRECTION_INCOMING:
			switch swap.Status {
			case types.SWAP_STATUS_OPEN:
				// This index expires unclaimed swaps
				keeper.InsertIntoByBlockIndex(ctx, swap)
				incomingSupplies = incomingSupplies.Add(swap.Amount...)
			case types.SWAP_STATUS_EXPIRED:
				incomingSupplies = incomingSupplies.Add(swap.Amount...)
			case types.SWAP_STATUS_COMPLETED:
				// This index stores swaps until deletion
				keeper.InsertIntoLongtermStorage(ctx, swap)
			default:
				panic(fmt.Sprintf("swap %s has invalid status %s", swap.GetSwapID(), swap.Status.String()))
			}
		case types.SWAP_DIRECTION_OUTGOING:
			switch swap.Status {
			case types.SWAP_STATUS_OPEN:
				keeper.InsertIntoByBlockIndex(ctx, swap)
				outgoingSupplies = outgoingSupplies.Add(swap.Amount...)
			case types.SWAP_STATUS_EXPIRED:
				outgoingSupplies = outgoingSupplies.Add(swap.Amount...)
			case types.SWAP_STATUS_COMPLETED:
				keeper.InsertIntoLongtermStorage(ctx, swap)
			default:
				panic(fmt.Sprintf("swap %s has invalid status %s", swap.GetSwapID(), swap.Status.String()))
			}
		default:
			panic(fmt.Sprintf("swap %s has invalid direction %s", swap.GetSwapID(), swap.Direction.String()))
		}
	}

	// Asset's given incoming/outgoing supply much match the amount of coins in incoming/outgoing atomic swaps
	supplies := keeper.GetAllAssetSupplies(ctx)
	for _, supply := range supplies {
		incomingSupply := incomingSupplies.AmountOf(supply.GetDenom())
		if !supply.IncomingSupply.Amount.Equal(incomingSupply) {
			panic(fmt.Sprintf("asset's incoming supply %s does not match amount %s in incoming atomic swaps",
				supply.IncomingSupply, incomingSupply))
		}
		outgoingSupply := outgoingSupplies.AmountOf(supply.GetDenom())
		if !supply.OutgoingSupply.Amount.Equal(outgoingSupply) {
			panic(fmt.Sprintf("asset's outgoing supply %s does not match amount %s in outgoing atomic swaps",
				supply.OutgoingSupply, outgoingSupply))
		}
		limit, err := keeper.GetSupplyLimit(ctx, supply.GetDenom())
		if err != nil {
			panic(err)
		}
		if supply.CurrentSupply.Amount.GT(limit.Limit) {
			panic(fmt.Sprintf("asset's current supply %s is over the supply limit %s", supply.CurrentSupply, limit.Limit))
		}
		if supply.IncomingSupply.Amount.GT(limit.Limit) {
			panic(fmt.Sprintf("asset's incoming supply %s is over the supply limit %s", supply.IncomingSupply, limit.Limit))
		}
		if supply.IncomingSupply.Amount.Add(supply.CurrentSupply.Amount).GT(limit.Limit) {
			panic(fmt.Sprintf("asset's incoming supply + current supply %s is over the supply limit %s", supply.IncomingSupply.Add(supply.CurrentSupply), limit.Limit))
		}
		if supply.OutgoingSupply.Amount.GT(limit.Limit) {
			panic(fmt.Sprintf("asset's outgoing supply %s is over the supply limit %s", supply.OutgoingSupply, limit.Limit))
		}

	}
}

// ExportGenesis writes the current store values to a genesis file, which can be imported again with InitGenesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) (data types.GenesisState) {
	params := k.GetParams(ctx)
	swaps := k.GetAllAtomicSwaps(ctx)
	supplies := k.GetAllAssetSupplies(ctx)
	previousBlockTime, found := k.GetPreviousBlockTime(ctx)
	if !found {
		previousBlockTime = types.DefaultPreviousBlockTime
	}
	return types.NewGenesisState(params, swaps, supplies, previousBlockTime)
}
