package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/kava-labs/kava/x/incentive/types"
)

const (
	keyActive  = "Active"
	keyRewards = "Rewards"
)

// genActive generates active bool with 80% chance of true
func genActive(r *rand.Rand) bool {
	threshold := 80
	value := simulation.RandIntBetween(r, 1, 100)
	if value > threshold {
		return false
	}
	return true
}

// ParamChanges defines the parameters that can be modified by param change proposals
func ParamChanges(r *rand.Rand) []simulation.ParamChange {
	return []simulation.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, keyActive,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%t\"", genActive(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, keyRewards,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%v\"", genRewards(r))
			},
		),
	}
}
