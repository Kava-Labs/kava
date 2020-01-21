package bep3

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/Kava-Labs/kava/x/bep3/internal/types"
)

// InitGenesis initialize default parameters
// and the keeper's address to pubkey map
func InitGenesis(ctx sdk.Context, k Keeper, /* TODO: Define what keepers the module needs */, data types.GenesisState) {

	// TODO: Define logic for when you would like to initalize a new genesis

	k.SetParams(ctx, data.Params)
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(ctx sdk.Context, k Keeper) (data types.GenesisState) {
	params := k.GetParams(ctx)

	// TODO: Define logic for exporting state

	return types.NewGenesisState(params, /* TODO: return the other types of your genesis state*/ )
}
