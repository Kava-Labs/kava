package issuance

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/issuance/keeper"
	"github.com/kava-labs/kava/x/issuance/types"
)

// InitGenesis initializes the store state from a genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, supplyKeeper types.SupplyKeeper, gs types.GenesisState) {

	if err := gs.Validate(); err != nil {
		panic(fmt.Sprintf("failed to validate %s genesis state: %s", ModuleName, err))
	}

	// check if the module account exists
	moduleAcc := supplyKeeper.GetModuleAccount(ctx, types.ModuleAccountName)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleAccountName))
	}

	k.SetParams(ctx, gs.Params)

	for _, supply := range gs.Supplies {
		k.SetAssetSupply(ctx, supply, supply.GetDenom())
	}

}

// ExportGenesis export genesis state for issuance module
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	params := k.GetParams(ctx)
	supplies := k.GetAllAssetSupplies(ctx)
	return types.NewGenesisState(params, supplies)
}
