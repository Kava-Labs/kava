package simulation

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/kava-labs/kava/x/cdp/types"
)

// RandomizedGenState generates a random GenesisState for cdp
func RandomizedGenState(simState *module.SimulationState) {

	// TODO implement this fully
	// - randomly generating the genesis params
	// - overwriting with genesis provided to simulation
	cdpGenesis := types.DefaultGenesisState()

	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, codec.MustMarshalJSONIndent(simState.Cdc, cdpGenesis))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(cdpGenesis)
}
