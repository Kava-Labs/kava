package v0_17

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/kava-labs/kava/app"
	tmtypes "github.com/tendermint/tendermint/types"
)

var (
	// TODO: needs verification before release
	GenesisTime = time.Date(2022, 5, 10, 17, 0, 0, 0, time.UTC)
	ChainID     = "kava-2222-10"
)

func setConfigIfUnsealed() {
	config := sdk.GetConfig()
	if config.GetBech32AccountAddrPrefix() == "kava" {
		return
	}
	app.SetSDKConfig()
}

// Migrate converts v16 genesis doc to v17 genesis doc
func Migrate(genDoc *tmtypes.GenesisDoc, ctx client.Context) (*tmtypes.GenesisDoc, error) {
	setConfigIfUnsealed()

	var appState genutiltypes.AppMap
	var err error
	if err := json.Unmarshal(genDoc.AppState, &appState); err != nil {
		return nil, fmt.Errorf("failed to unmarshal app state from genesis doc:  %w", err)
	}

	MigrateCosmosAppState(appState, ctx, GenesisTime)
	migrateAppState(appState, ctx)

	genDoc.AppState, err = json.Marshal(appState)
	if err != nil {
		return nil, err
	}

	genDoc.GenesisTime = GenesisTime
	genDoc.ChainID = ChainID
	genDoc.InitialHeight = 1

	return genDoc, nil
}
