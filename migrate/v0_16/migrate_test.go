package v0_16

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	genutil "github.com/cosmos/cosmos-sdk/x/genutil/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/kava-labs/kava/app"
)

func TestMigrateGenesisDoc(t *testing.T) {
	expected := getTestDataJSON("genesis-v16.json")
	genDoc, err := tmtypes.GenesisDocFromFile(filepath.Join("testdata", "genesis-v15.json"))
	assert.NoError(t, err)
	actualGenDoc, err := Migrate(genDoc, newClientContext())
	assert.NoError(t, err)
	actualJson, err := tmjson.Marshal(actualGenDoc)
	assert.NoError(t, err)
	assert.JSONEq(t, expected, string(actualJson))
}

func TestMigrateFull(t *testing.T) {
	t.Skip() // avoid committing mainnet state - test also currently fails due to https://github.com/cosmos/cosmos-sdk/issues/10862. If you apply the patch, it will pass
	genDoc, err := tmtypes.GenesisDocFromFile(filepath.Join("testdata", "kava-8-block-864866.json"))
	assert.NoError(t, err)
	newGenDoc, err := Migrate(genDoc, newClientContext())
	assert.NoError(t, err)

	encodingConfig := app.MakeEncodingConfig()
	var appMap genutil.AppMap
	err = tmjson.Unmarshal(newGenDoc.AppState, &appMap)
	assert.NoError(t, err)
	err = app.ModuleBasics.ValidateGenesis(encodingConfig.Marshaler, encodingConfig.TxConfig, appMap)
	assert.NoError(t, err)
	tApp := app.NewTestApp()
	require.NotPanics(t, func() {
		tApp.InitializeFromGenesisStatesWithTimeAndChainID(newGenDoc.GenesisTime, newGenDoc.ChainID, app.GenesisState(appMap))
	})
}
