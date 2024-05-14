package keeper_test

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/precisebank/keeper"
	"github.com/kava-labs/kava/x/precisebank/types"
)

// testKeeper defines necessary fields for testing keeper store methods that
// don't require a full app setup.
type testKeeper struct {
	ctx      sdk.Context
	keeper   keeper.Keeper
	storeKey *storetypes.KVStoreKey
}

func NewTestKeeper() testKeeper {
	storeKey := sdk.NewKVStoreKey(types.ModuleName)
	// Not required by module, but needs to be non-nil for context
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)

	tApp := app.NewTestApp()
	cdc := tApp.AppCodec()
	k := keeper.NewKeeper(cdc, storeKey)

	return testKeeper{
		ctx:      ctx,
		keeper:   k,
		storeKey: storeKey,
	}
}
