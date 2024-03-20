package v3_test

import (
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	v3 "github.com/kava-labs/kava/x/evmutil/migrations/v3"
	"testing"
)

func TestMigratePrecompiles(t *testing.T) {
	tApp := app.NewTestApp()
	//cdc := tApp.AppCodec()
	storeKey := sdk.NewKVStoreKey("evmutil")
	ctx := testutil.DefaultContext(storeKey, sdk.NewTransientStoreKey("transient_test"))
	//store := ctx.KVStore(storeKey)

	//ctx sdk.Context,
	//evmKeeper *evmkeeper.Keeper,
	v3.Migrate(ctx)

	//require.Nil(
	//	t,
	//	store.Get(types.ParamsKey),
	//	"params shouldn't exist in store before migration",
	//)
	//
	//require.NoError(t, v2.Migrate(ctx, store, cdc))
	//
	//paramsBytes := store.Get(types.ParamsKey)
	//require.NotNil(t, paramsBytes, "params should be in store after migration")
	//
	//var params types.Params
	//cdc.MustUnmarshal(paramsBytes, &params)
	//
	//t.Logf("params: %+v", params)
	//
	//require.Equal(
	//	t,
	//	types.NewParams(
	//		time.Date(2023, 11, 1, 0, 0, 0, 0, time.UTC),
	//		sdk.NewInt(744191),
	//	),
	//	params,
	//	"params should be correct after migration",
	//)
}
