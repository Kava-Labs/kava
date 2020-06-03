package kava3

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/genutil"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/migrate/v0_8"
	v032tendermint "github.com/kava-labs/kava/migrate/v0_8/tendermint/v0_32"
)

func TestAddSuggestedParams(t *testing.T) {
	tApp := app.NewTestApp() // also sets the bech32 prefix on sdk.Config
	cdc := app.MakeCodec()

	// 1) load an exported kava-2 state and migrate to kava v0.8 format (avoids storing v0.8 state that can get out of date)
	oldGenDoc, err := v032tendermint.GenesisDocFromFile(filepath.Join("../../migrate/v0_8/testdata", "kava-2.json"))
	require.NoError(t, err)
	genDoc := v0_8.Migrate(*oldGenDoc)

	// 2) add params
	newGenDoc, err := AddSuggestedParams(cdc, genDoc, "new-chain-id", time.Date(1998, 1, 0, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)

	// 3) check new genesis is valid
	var newAppState genutil.AppMap
	require.NoError(t,
		cdc.UnmarshalJSON(newGenDoc.AppState, &newAppState),
	)
	require.NoError(t,
		app.ModuleBasics.ValidateGenesis(newAppState),
	)
	require.NotPanics(t, func() {
		// this runs both InitGenesis for all modules (which panic on errors) and runs all invariants
		tApp.InitializeFromGenesisStates(app.GenesisState(newAppState))
	})
}
