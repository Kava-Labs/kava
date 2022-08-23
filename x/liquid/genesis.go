package liquid

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/liquid/keeper"
	"github.com/kava-labs/kava/x/liquid/types"
)

// InitGenesis initializes genesis state
func InitGenesis(ctx sdk.Context, k keeper.Keeper, ak types.AccountKeeper, gs types.GenesisState) {
	if err := gs.Validate(); err != nil {
		panic(fmt.Sprintf("failed to validate %s genesis state: %s", types.ModuleName, err))
	}

	k.SetParams(ctx, gs.Params)
}

// ExportGenesis returns a GenesisState for a given context and keeper
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	params := k.GetParams(ctx)
	return types.NewGenesisState(params)
}
