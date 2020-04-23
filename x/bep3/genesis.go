package bep3

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the store state from a genesis state.
func InitGenesis(ctx sdk.Context, keeper Keeper, supplyKeeper SupplyKeeper, gs GenesisState) {
	if err := gs.Validate(); err != nil {
		panic(fmt.Sprintf("failed to validate %s genesis state: %s", ModuleName, err))
	}

	keeper.SetParams(ctx, gs.Params)

	// Initialize supported assets
	for _, asset := range gs.Params.SupportedAssets {
		zeroCoin := sdk.NewCoin(asset.Denom, sdk.NewInt(0))
		supply := NewAssetSupply(asset.Denom, zeroCoin, zeroCoin, zeroCoin, sdk.NewCoin(asset.Denom, asset.Limit))
		keeper.SetAssetSupply(ctx, supply, []byte(asset.Denom))
	}

	// Increment an asset's incoming, current, and outgoing supply
	// It it required that assets are supported but they do not have to be active
	for _, supply := range gs.AssetSupplies {
		// Asset must be supported but does not have to be active
		coin, found := keeper.GetAssetByDenom(ctx, supply.Denom)
		if !found {
			panic(fmt.Sprintf("invalid asset supply: %s is not a supported asset", coin.Denom))
		}
		if !coin.Limit.Equal(supply.Limit.Amount) {
			panic(fmt.Sprintf("supported asset limit %s does not equal asset supply %s", coin.Limit, supply.Limit.Amount))
		}

		// Increment current, incoming, and outgoing asset supplies
		err := keeper.IncrementCurrentAssetSupply(ctx, supply.CurrentSupply)
		if err != nil {
			panic(err)
		}
		err = keeper.IncrementIncomingAssetSupply(ctx, supply.IncomingSupply)
		if err != nil {
			panic(err)
		}
		err = keeper.IncrementOutgoingAssetSupply(ctx, supply.OutgoingSupply)
		if err != nil {
			panic(err)
		}
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
	supplies := keeper.GetAllAssetSupplies(ctx)
	for _, supply := range supplies {
		incomingSupply := incomingSupplies.AmountOf(supply.Denom)
		if !supply.IncomingSupply.Amount.Equal(incomingSupply) {
			panic(fmt.Sprintf("asset's incoming supply %s does not match amount %s in incoming atomic swaps",
				supply.IncomingSupply, incomingSupply))
		}
		outgoingSupply := outgoingSupplies.AmountOf(supply.Denom)
		if !supply.OutgoingSupply.Amount.Equal(outgoingSupply) {
			panic(fmt.Sprintf("asset's outgoing supply %s does not match amount %s in outgoing atomic swaps",
				supply.OutgoingSupply, outgoingSupply))
		}
	}
}

// ExportGenesis writes the current store values to a genesis file, which can be imported again with InitGenesis
func ExportGenesis(ctx sdk.Context, k Keeper) (data GenesisState) {
	params := k.GetParams(ctx)
	swaps := k.GetAllAtomicSwaps(ctx)
	assets := k.GetAllAssetSupplies(ctx)
	return NewGenesisState(params, swaps, assets)
}
