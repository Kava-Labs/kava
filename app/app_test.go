package app

import (
	"encoding/json"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	db "github.com/tendermint/tm-db"
)

func TestNewApp(t *testing.T) {

	NewApp(
		log.NewTMLogger(log.NewSyncWriter(os.Stdout)),
		db.NewMemDB(),
		DefaultNodeHome,
		nil,
		MakeEncodingConfig(),
		Options{},
		simapp.EmptyAppOptions{},
	)
}

func TestExport(t *testing.T) {
	db := db.NewMemDB()
	app := NewApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, DefaultNodeHome, nil, MakeEncodingConfig(), Options{})

	stateBytes, err := json.Marshal(NewDefaultGenesisState())
	require.NoError(t, err)

	initRequest := abci.RequestInitChain{
		Time:            time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
		ChainId:         "kava-test",
		InitialHeight:   1,
		ConsensusParams: tmtypes.TM2PB.ConsensusParams(tmtypes.DefaultConsensusParams()),
		Validators:      nil,
		AppStateBytes:   stateBytes,
	}
	app.InitChain(initRequest)
	app.Commit()

	exportedApp, err := app.ExportAppStateAndValidators(false, []string{})
	require.NoError(t, err)

	// Assume each module is exported correctly, so only check modules in genesis are present in export
	initialModules, err := unmarshalJSONKeys(initRequest.AppStateBytes)
	require.NoError(t, err)
	exportedModules, err := unmarshalJSONKeys(exportedApp.AppState)
	require.NoError(t, err)
	assert.ElementsMatch(t, initialModules, exportedModules)

	assert.Equal(t, initRequest.InitialHeight+1, exportedApp.Height) // app.Commit() increments height
	assert.Equal(t, initRequest.ConsensusParams, exportedApp.ConsensusParams)
	assert.Equal(t, []tmtypes.GenesisValidator(nil), exportedApp.Validators) // no validators set in default genesis
}

// unmarshalJSONKeys extracts keys from the top level of a json blob.
func unmarshalJSONKeys(jsonBytes []byte) ([]string, error) {
	var jsonMap map[string]json.RawMessage
	err := json.Unmarshal(jsonBytes, &jsonMap)
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(jsonMap))
	for k, _ := range jsonMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys, nil
}
