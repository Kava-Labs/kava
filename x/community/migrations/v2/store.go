package v2

import (
	"time"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/community/types"
)

const (
	ModuleName = "mint"
)

// Migrate migrates the x/community module state from the consensus version 1 to
// version 2. Specifically, sets new parameters in the module state.
func Migrate(
	ctx sdk.Context,
	store storetypes.KVStore,
	cdc codec.BinaryCodec,
) error {
	params := types.NewParams(
		// 2023-11-01T00:00:00Z
		time.Date(2023, 11, 1, 0, 0, 0, 0, time.UTC),
		sdk.NewInt(744191),
	)

	if err := params.Validate(); err != nil {
		return err
	}

	bz := cdc.MustMarshal(&params)
	store.Set(types.ParamsKey, bz)

	return nil
}
