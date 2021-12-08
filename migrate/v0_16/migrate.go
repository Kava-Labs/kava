package v0_16

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

var (
	GenesisTime = time.Date(2021, 11, 30, 15, 0, 0, 0, time.UTC)
	ChainID     = "kava-8"
)

// Migrate converts v15 genesis doc to v16 genesis doc
func Migrate(genDoc *tmtypes.GenesisDoc, ctx client.Context) (*tmtypes.GenesisDoc, error) {
	var appState genutiltypes.AppMap
	var err error
	if err := json.Unmarshal(genDoc.AppState, &appState); err != nil {
		return nil, fmt.Errorf("failed to marchal app state from genesis doc:  %w", err)
	}

	appState = MigrateCosmosAppState(appState, ctx)

	// TODO: Migrate kava modules

	genDoc.AppState, err = json.Marshal(appState)
	if err != nil {
		return nil, err
	}

	genDoc.GenesisTime = GenesisTime
	genDoc.ChainID = ChainID
	return genDoc, nil
}
