package simulation

import (
	"fmt"
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/kava-labs/kava/x/kavadist/types"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	// Hacky way to validate periods since validation is wrapped in params
	active := genRandomActive(r)
	periods := genRandomPeriods(r, simtypes.RandTimestamp(r))
	if err := types.NewParams(active, periods).Validate(); err != nil {
		panic(err)
	}

	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyActive),
			func(r *rand.Rand) string {
				return fmt.Sprintf("%t", active)
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyPeriods),
			func(r *rand.Rand) string {
				return fmt.Sprintf("%v", periods)
			},
		),
	}
}
