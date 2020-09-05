package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/kava-labs/kava/x/hvt/types"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(r *rand.Rand) []simulation.ParamChange {
	// Hacky way to validate periods since validation is wrapped in params
	active := genRandomActive(r)
	periods := genRandomPeriods(r, simulation.RandTimestamp(r))
	if err := types.NewParams(active, periods).Validate(); err != nil {
		panic(err)
	}

	return []simulation.ParamChange{
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
