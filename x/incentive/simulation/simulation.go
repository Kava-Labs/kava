package simulation

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/kava-labs/kava/x/incentive/types"
)

// RandomizedGenState generates a random GenesisState for cdp
func RandomizedGenState(simState *module.SimulationState) {

	// TODO implement this fully
	// - randomly generating the genesis params
	// - overwriting with genesis provided to simulation
	genesis := types.DefaultGenesisState()

	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, codec.MustMarshalJSONIndent(simState.Cdc, genesis))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genesis)
}
