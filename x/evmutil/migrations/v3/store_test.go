package v3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	v3 "github.com/kava-labs/kava/x/evmutil/migrations/v3"
	v3types "github.com/kava-labs/kava/x/evmutil/migrations/v3/types"
	"github.com/kava-labs/kava/x/evmutil/types"
	"github.com/kava-labs/kava/x/evmutil/types/mocks"
)

func TestStoreMigrationToPreciseBank(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	evmutilKey := sdk.NewKVStoreKey(types.ModuleName)
	tEvmutilKey := sdk.NewTransientStoreKey("transient_test")

	ctx := testutil.DefaultContext(evmutilKey, tEvmutilKey)

	store := ctx.KVStore(evmutilKey)

	preciseBankKeeper := mocks.NewMockPreciseBankKeeper(t)

	accounts := []*v3types.Account{}
	for i := 0; i < 10; i++ {
		bal := sdk.NewInt(1000 * int64(i))
		acc := v3types.NewAccount(
			sdk.AccAddress([]byte{byte(i)}),
			bal,
		)

		accounts = append(accounts, acc)

		// Set in store
		bz := encCfg.Codec.MustMarshal(acc)
		accountKey := v3.AccountStoreKey(acc.Address)
		store.Set(accountKey, bz)

		// Register expectation call for precise bank
		preciseBankKeeper.EXPECT().
			SetFractionalBalance(ctx, acc.Address, acc.Balance).
			Once()
	}

	// Run migrations
	err := v3.MigrateStore(ctx, encCfg.Codec, evmutilKey, preciseBankKeeper)
	require.NoError(t, err)

	// Check removed from store
	for _, acc := range accounts {
		accountKey := types.AccountStoreKey(acc.Address)
		require.False(t, store.Has(accountKey))
	}

	// Check all accounts are removed
	iterator := sdk.KVStorePrefixIterator(store, v3.AccountStoreKeyPrefix)
	defer iterator.Close()
	require.False(t, iterator.Valid(), "there should be no accounts left in the store")
}
