package v0_17

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kava-labs/kava/app"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	tmtypes "github.com/tendermint/tendermint/types"
)

var (
	// Configure go live time for the next version of the chain
	GenesisTime = time.Date(2022, 5, 11, 16, 0, 0, 0, time.UTC)
	// Configure the id of the next version of the chain
	ChainID = "kava-10"
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
		return nil, fmt.Errorf("failed to marchal app state from genesis doc:  %w", err)
	}

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
