package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/kava-labs/kava/x/kavadist/types"
)

// MigrateStore performs in-place migrations from kava 10 to kava 11.
// The migration includes:
//
// - Sets the default InfrastructureParams parameter.
func MigrateStore(ctx sdk.Context, ss paramtypes.Subspace) error {
	if !ss.HasKeyTable() {
		ss = ss.WithKeyTable(types.ParamKeyTable())
	}
	infraParams := types.DefaultInfraParams
	ss.Set(ctx, types.KeyInfra, infraParams)
	return nil
}
