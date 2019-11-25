package liquidator

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis sets the genesis state in the keeper.
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) {

	keeper.SetParams(ctx, data.Params)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	params := keeper.GetParams(ctx)
	return GenesisState{
		Params: params,
	}
}
