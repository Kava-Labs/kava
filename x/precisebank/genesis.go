package precisebank

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/precisebank/keeper"
	"github.com/kava-labs/kava/x/precisebank/types"
)

// InitGenesis initializes the store state from a genesis state.
func InitGenesis(
	ctx sdk.Context,
	keeper keeper.Keeper,
	ak types.AccountKeeper,
	gs *types.GenesisState,
) {
	if err := gs.Validate(); err != nil {
		panic(fmt.Sprintf("failed to validate %s genesis state: %s", types.ModuleName, err))
	}

	// initialize module account
	if moduleAcc := ak.GetModuleAccount(ctx, types.ModuleName); moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// TODO:
	// - Set balances
	// - Ensure reserve account exists
	// - Ensure reserve balance matches sum of all fractional balances
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper keeper.Keeper) *types.GenesisState {
	return types.NewGenesisState(nil)
}
