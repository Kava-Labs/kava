package v2_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	v2cdp "github.com/kava-labs/kava/x/cdp/migrations/v2"
	"github.com/kava-labs/kava/x/cdp/types"
)

func TestStoreMigrationAddsKeyTableIncludingNewParam(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	cdpKey := sdk.NewKVStoreKey(types.ModuleName)
	tcdpKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(cdpKey, tcdpKey)
	paramstore := paramtypes.NewSubspace(encCfg.Codec, encCfg.Amino, cdpKey, tcdpKey, types.ModuleName)

	// Check param doesn't exist before
	require.False(t, paramstore.Has(ctx, types.KeyBeginBlockerExecutionBlockInterval))

	// Run migrations.
	err := v2cdp.MigrateStore(ctx, paramstore)
	require.NoError(t, err)

	// Make sure the new params are set.
	require.True(t, paramstore.Has(ctx, types.KeyBeginBlockerExecutionBlockInterval))
	// Assert the value is what we expect
	result := types.DefaultBeginBlockerExecutionBlockInterval
	paramstore.Get(ctx, types.KeyBeginBlockerExecutionBlockInterval, &result)
	require.Equal(t, result, types.DefaultBeginBlockerExecutionBlockInterval)
}

func TestStoreMigrationSetsNewParamOnExistingKeyTable(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	cdpKey := sdk.NewKVStoreKey(types.ModuleName)
	tcdpKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(cdpKey, tcdpKey)
	paramstore := paramtypes.NewSubspace(encCfg.Codec, encCfg.Amino, cdpKey, tcdpKey, types.ModuleName)
	paramstore.WithKeyTable(types.ParamKeyTable())

	// expect it to have key table
	require.True(t, paramstore.HasKeyTable())
	// expect it to not have new param
	require.False(t, paramstore.Has(ctx, types.KeyBeginBlockerExecutionBlockInterval))

	// Run migrations.
	err := v2cdp.MigrateStore(ctx, paramstore)
	require.NoError(t, err)

	// Make sure the new params are set.
	require.True(t, paramstore.Has(ctx, types.KeyBeginBlockerExecutionBlockInterval))

	// Assert the value is what we expect
	result := types.DefaultBeginBlockerExecutionBlockInterval
	paramstore.Get(ctx, types.KeyBeginBlockerExecutionBlockInterval, &result)
	require.Equal(t, result, types.DefaultBeginBlockerExecutionBlockInterval)
}
