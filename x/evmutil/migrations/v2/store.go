package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/kava-labs/kava/x/evmutil/types"
)

// MigrateStore performs in-place store migrations for consensus version 2
// V2 adds the allowed_cosmos_denoms param to parameters.
func MigrateStore(ctx sdk.Context, paramstore paramtypes.Subspace) error {
	migrateParamsStore(ctx, paramstore)
	return nil
}

// migrateParamsStore ensures the param key table exists and has the allowed_cosmos_denoms property
func migrateParamsStore(ctx sdk.Context, paramstore paramtypes.Subspace) {
	if !paramstore.HasKeyTable() {
		paramstore.WithKeyTable(types.ParamKeyTable())
	}
	paramstore.Set(ctx, types.KeyAllowedCosmosDenoms, types.DefaultAllowedCosmosDenoms)
}
