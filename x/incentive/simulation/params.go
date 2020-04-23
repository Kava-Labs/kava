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

// ParamChanges defines the parameters that can be modified by param change proposals
func ParamChanges(r *rand.Rand) []simulation.ParamChange {
	return []simulation.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, keyActive,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%t\"", GenActive(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, keyRewards,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%v\"", GenRewards(r))
			},
		),
	}
}
