package kava3

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/genutil"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/kava-labs/kava/app"
)

func TestAddSuggestedParams(t *testing.T) {
	// 1) load an exported kava-2 state that has been migrated to kava v0.8 format
	genDoc, err := tmtypes.GenesisDocFromFile(filepath.Join("testdata", "kava-2-migrated.json"))
	require.NoError(t, err)
	tApp := app.NewTestApp() // also sets the bech32 prefix on sdk.Config
	cdc := app.MakeCodec()

	// 2) add params
	newGenDoc, err := AddSuggestedParams(cdc, *genDoc, "new-chain-id", time.Date(1998, 1, 0, 0, 0, 0, 0, time.UTC))
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
