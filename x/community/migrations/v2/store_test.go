package v2_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	v2 "github.com/kava-labs/kava/x/community/migrations/v2"
	"github.com/kava-labs/kava/x/community/types"
)

func TestMigrateStore(t *testing.T) {
	tApp := app.NewTestApp()
	cdc := tApp.AppCodec()
	storeKey := sdk.NewKVStoreKey("community")
	ctx := testutil.DefaultContext(storeKey, sdk.NewTransientStoreKey("transient_test"))
	store := ctx.KVStore(storeKey)

	require.Nil(
		t,
		store.Get(types.ParamsKey),
		"params shouldn't exist in store before migration",
	)

	require.NoError(t, v2.Migrate(ctx, store, cdc))

	paramsBytes := store.Get(types.ParamsKey)
	require.NotNil(t, paramsBytes, "params should be in store after migration")

	var params types.Params
	cdc.MustUnmarshal(paramsBytes, &params)

	t.Logf("params: %+v", params)

	require.Equal(
		t,
		types.NewParams(
			time.Time{},
			sdkmath.LegacyNewDec(0),
			sdkmath.LegacyNewDec(0),
		),
		params,
		"params should be correct after migration",
	)
}
