package v0_16

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/kava-labs/kava/app"
)

var (
	GenesisTime = time.Date(2022, 1, 19, 16, 0, 0, 0, time.UTC)
	ChainID     = "kava-9"
)

// Migrate converts v15 genesis doc to v16 genesis doc
func Migrate(genDoc *tmtypes.GenesisDoc, ctx client.Context) (*tmtypes.GenesisDoc, error) {
	app.SetSDKConfig()

	var appState genutiltypes.AppMap
	var err error
	if err := json.Unmarshal(genDoc.AppState, &appState); err != nil {
		return nil, fmt.Errorf("failed to marchal app state from genesis doc:  %w", err)
	}

	MigrateCosmosAppState(appState, ctx, GenesisTime)
	migrateKavaAppState(appState, ctx)

	genDoc.AppState, err = json.Marshal(appState)
	if err != nil {
		return nil, err
	}

	genDoc.GenesisTime = GenesisTime
	genDoc.ChainID = ChainID

	genDoc.InitialHeight = 1

	genDoc.ConsensusParams.Version.AppVersion = 1

	genDoc.ConsensusParams.Evidence.MaxBytes = 50000

	return genDoc, nil
}
