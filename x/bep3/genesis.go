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

	keeper.SetParams(ctx, gs.Params)

	for _, swap := range gs.AtomicSwaps {
		if swap.Validate() != nil {
			panic(fmt.Sprintf("invalid swap %s", swap.GetSwapID()))
		}

		// Confirm that the asset is supported and active
		err := keeper.ValidateLiveAsset(ctx, swap.Amount[0])
		if err != nil {
			panic(err)
		}

		// Validate that this amount is within supply limits
		err = keeper.ValidateCreateSwapAgainstSupplyLimit(ctx, swap.Amount[0])
		if err != nil {
			panic(err)
		}

		keeper.SetAtomicSwap(ctx, swap)

		// Add swap to correct storage indexes and/or increment in swap supply based on status
		switch swap.Status {
		case types.Open:
			keeper.IncrementInSwapSupply(ctx, swap.Amount[0])
			keeper.InsertIntoByBlockIndex(ctx, swap) // used to expire swaps
		case types.Expired:
			keeper.IncrementInSwapSupply(ctx, swap.Amount[0])
		case types.Completed:
			keeper.InsertIntoLongtermStorage(ctx, swap) // used to delete swaps
		default:
			panic(fmt.Sprintf("swap %s has invalid status %s", swap.GetSwapID(), swap.Status.String()))
		}
	}

	// Build map for assets in swap supplies
	inSwapSupplyMap := make(map[string]sdk.Int)
	inSwapSupplies := keeper.GetAllInSwapSupplies(ctx)
	for _, swapSupply := range inSwapSupplies {
		inSwapSupplyMap[swapSupply.Denom] = swapSupply.Amount
	}

	// Set each asset supply after validating that it is supported and within valid
	// supply limits. Assets must be supported but do not have to be active.
	var totalAssetSupply sdk.Coins
	for _, asset := range gs.AssetSupplies {
		coin, found := keeper.GetAssetByDenom(ctx, asset.Denom)
		if !found {
			panic(fmt.Sprintf("invalid asset supply: %s is not a supported asset", coin.Denom))
		}
		if asset.Amount.Add(inSwapSupplyMap[asset.Denom]).GT(coin.Limit) {
			panic(fmt.Sprintf("invalid asset supply: %s has a supply limit of %s", coin.Denom, coin.Limit))
		}
		keeper.SetAssetSupply(ctx, asset, []byte(asset.Denom))
		totalAssetSupply = totalAssetSupply.Add(sdk.NewCoins(asset))
	}
}

// ExportGenesis writes the current store values to a genesis file, which can be imported again with InitGenesis
func ExportGenesis(ctx sdk.Context, k Keeper) (data types.GenesisState) {
	params := k.GetParams(ctx)
	swaps := k.GetAllAtomicSwaps(ctx)
	assets := k.GetAllAssetSupplies(ctx)
	return types.NewGenesisState(params, swaps, assets)
}
