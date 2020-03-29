package simulation

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/kava-labs/kava/x/auction/types"
)

// RandomizedGenState generates a random GenesisState for the module
func RandomizedGenState(simState *module.SimulationState) {

	// TODO implement this fully
	// - randomly generating the genesis params
	// - overwriting with genesis provided to simulation
	genesisState := types.DefaultGenesisState()

	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, codec.MustMarshalJSONIndent(simState.Cdc, genesisState))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genesisState)
}
