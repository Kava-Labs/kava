package v2

import (
	"time"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/community/types"
)

// Migrate migrates the x/community module state from the consensus version 1 to
// version 2. Specifically, sets new parameters in the module state.
func Migrate(
	ctx sdk.Context,
	store storetypes.KVStore,
	cdc codec.BinaryCodec,
) error {
	params := types.NewParams(
		time.Time{},
		sdkmath.LegacyNewDec(0),
		sdkmath.LegacyNewDec(0),
	)

	if err := params.Validate(); err != nil {
		return err
	}

	bz := cdc.MustMarshal(&params)
	store.Set(types.ParamsKey, bz)

	return nil
}
