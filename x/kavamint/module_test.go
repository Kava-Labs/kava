package kavamint_test

import (
	"testing"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/kavamint/types"
)

func TestKavaMintModuleAccountWithPermissionsOnAppInit(t *testing.T) {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, tmproto.Header{Height: 1})
	tApp.InitializeFromGenesisStates()

	ak := tApp.GetAccountKeeper()

	// by pass auto creation of module accounts
	addr, _ := ak.GetModuleAddressAndPermissions(types.ModuleName)
	acc := ak.GetAccount(ctx, addr)

	require.NotNil(t, acc, "expected module account to exist")

	macc, ok := acc.(authtypes.ModuleAccountI)
	require.True(t, ok)

	require.True(t, macc.HasPermission(authtypes.Minter), "expected module account to have mint permissions")
}
