package v0_11

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	v0_9cdp "github.com/kava-labs/kava/x/cdp/legacy/v0_9"
)

func TestMain(m *testing.M) {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
	app.SetBip44CoinType(config)

	os.Exit(m.Run())
}

func TestMigrateCdp(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "cdp-v09.json"))
	require.NoError(t, err)
	var oldGenState v0_9cdp.GenesisState
	cdc := app.MakeCodec()
	require.NotPanics(t, func() {
		cdc.MustUnmarshalJSON(bz, &oldGenState)
	})

	newGenState := MigrateCDP(oldGenState)
	t.Log(fmt.Sprintf("%d", newGenState.StartingCdpID))
	err = newGenState.Validate()
	require.NoError(t, err)

}
