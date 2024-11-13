package app

import (
	"cosmossdk.io/log"
	"encoding/json"
	dbm "github.com/cosmos/cosmos-db"
)

// GenesisState represents the genesis state of the blockchain. It is a map from module names to module genesis states.
type GenesisState map[string]json.RawMessage

// NewDefaultGenesisState generates the default state for the application.
func NewDefaultGenesisState() GenesisState {
	encCfg := MakeEncodingConfig()
	tempApp := NewApp(log.NewNopLogger(), dbm.NewMemDB(), DefaultNodeHome, nil, encCfg, DefaultOptions)

	return tempApp.BasicModuleManager.DefaultGenesis(encCfg.Marshaler)
}
