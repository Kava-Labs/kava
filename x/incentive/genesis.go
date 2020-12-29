package incentive

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
)

// InitGenesis initializes the store state from a genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, supplyKeeper types.SupplyKeeper, gs types.GenesisState) {

	// check if the module account exists
	moduleAcc := supplyKeeper.GetModuleAccount(ctx, types.IncentiveMacc)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.IncentiveMacc))
	}

	if err := gs.Validate(); err != nil {
		panic(fmt.Sprintf("failed to validate %s genesis state: %s", types.ModuleName, err))
	}

	k.SetParams(ctx, gs.Params)

	for _, gat := range gs.PreviousAccumulationTimes {
		k.SetPreviousAccrualTime(ctx, gat.CollateralType, gat.PreviousAccumulationTime)
	}

}

// ExportGenesis export genesis state for incentive module
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {

	return types.DefaultGenesisState()
}
