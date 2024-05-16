package keeper_test

import (
	"testing"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/precisebank/keeper"
	"github.com/kava-labs/kava/x/precisebank/types"
	"github.com/kava-labs/kava/x/precisebank/types/mocks"
)

// testData defines necessary fields for testing keeper store methods and mocks
// for unit tests without full app setup.
type testData struct {
	ctx      sdk.Context
	keeper   keeper.Keeper
	storeKey *storetypes.KVStoreKey
	bk       *mocks.MockBankKeeper
	ak       *mocks.MockAccountKeeper
}

// NewMockedTestData creates a new testData instance with mocked bank and
// account keepers.
func NewMockedTestData(t *testing.T) testData {
	storeKey := sdk.NewKVStoreKey(types.ModuleName)
	// Not required by module, but needs to be non-nil for context
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)

	bk := mocks.NewMockBankKeeper(t)
	ak := mocks.NewMockAccountKeeper(t)

	tApp := app.NewTestApp()
	cdc := tApp.AppCodec()
	k := keeper.NewKeeper(cdc, storeKey, bk, ak)

	return testData{
		ctx:      ctx,
		keeper:   k,
		storeKey: storeKey,
		bk:       bk,
		ak:       ak,
	}
}
