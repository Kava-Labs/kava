package app

import (
	"encoding/json"
)

// GenesisState represents the genesis state of the blockchain. It is a map from module names to module genesis states.
type GenesisState map[string]json.RawMessage

// TODO
// // NewDefaultGenesisState generates the default state for the application.
// func NewDefaultGenesisState() GenesisState {
// 	return ModuleBasics.DefaultGenesis()
// }
