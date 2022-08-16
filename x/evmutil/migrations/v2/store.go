package v3

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/kava-labs/kava/x/evmutil/types"
)

// MigrateStore performs in-place migrations from kava 10 to kava 11.
// The migration includes:
//
// - Setting the default evmutil params in the paramstore
func MigrateStore(ctx sdk.Context, ss paramtypes.Subspace) error {
	if !ss.HasKeyTable() {
		ss = ss.WithKeyTable(types.ParamKeyTable())
	}
	params := types.DefaultParams()
	ss.SetParamSet(ctx, &params)
	return nil
}
