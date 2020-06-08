package app

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/std"
)

// GenesisState represents the genesis state of the blockchain. It is a map from module names to module genesis states.
type GenesisState map[string]json.RawMessage

// NewDefaultGenesisState generates the default state for the application.
func NewDefaultGenesisState() GenesisState {
	cdc := std.MakeCodec(ModuleBasics)
	return ModuleBasics.DefaultGenesis(cdc)
}
