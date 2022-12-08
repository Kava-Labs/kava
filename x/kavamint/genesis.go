package kavamint

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/kava-labs/kava/x/kavamint/keeper"
	"github.com/kava-labs/kava/x/kavamint/types"
)

// InitGenesis new mint genesis
func InitGenesis(ctx sdk.Context, keeper keeper.Keeper, ak types.AccountKeeper, gs *types.GenesisState) {
	// guard against invalid genesis
	if err := gs.Validate(); err != nil {
		panic(fmt.Sprintf("failed to validate %s genesis state: %s", types.ModuleName, err))
	}

	// get module account -- creates one with allowed permissions if it does not exist
	macc := ak.GetModuleAccount(ctx, types.ModuleName)
	if macc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// check module account has minter permissions
	if !macc.HasPermission(authtypes.Minter) {
		panic(fmt.Sprintf("%s module account does not have %s permissions", types.ModuleName, authtypes.Minter))
	}

	// set store state from genesis
	keeper.SetParams(ctx, gs.Params)
	keeper.SetPreviousBlockTime(ctx, gs.PreviousBlockTime)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper keeper.Keeper) *types.GenesisState {
	return types.NewGenesisState(keeper.GetParams(ctx), keeper.GetPreviousBlockTime(ctx))
}
