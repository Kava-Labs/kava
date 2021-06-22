package v0_15

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/stretchr/testify/require"

	v0_14incentive "github.com/kava-labs/kava/x/incentive/legacy/v0_14"
	v0_15incentive "github.com/kava-labs/kava/x/incentive/types"
)

func TestIncentive(t *testing.T) {
	bz, err := ioutil.ReadFile(filepath.Join("testdata", "v0_14-incentive-state.json"))
	require.NoError(t, err)
	appState := genutil.AppMap{v0_14incentive.ModuleName: bz}

	MigrateAppState(appState)

	bz, err = ioutil.ReadFile(filepath.Join("testdata", "v0_15-incentive-state.json"))
	require.NoError(t, err)

	fmt.Println(string(appState[v0_15incentive.ModuleName]))
	require.JSONEq(t, string(bz), string(appState[v0_15incentive.ModuleName]))
}
