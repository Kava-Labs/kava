package savings

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/savings/keeper"
	"github.com/kava-labs/kava/x/savings/types"
)

// InitGenesis initializes genesis state
func InitGenesis(ctx sdk.Context, k keeper.Keeper, ak types.AccountKeeper, gs types.GenesisState) {
	k.SetParams(ctx, gs.Params)

	// check if the module account exists
	SavingsModuleAccount := ak.GetModuleAccount(ctx, types.ModuleAccountName)
	if SavingsModuleAccount == nil {
		panic(fmt.Sprintf("%s module account has not been set", SavingsModuleAccount))
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	params := k.GetParams(ctx)
	return types.NewGenesisState(params)
}
