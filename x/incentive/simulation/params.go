package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/x/simulation"
)

const (
	keyActive  = "Active"
	keyRewards = "Rewards"
)

// genActive generates active bool with 80% chance of true
func genActive(r *rand.Rand) bool {
	threshold := 80
	value := simulation.RandIntBetween(r, 1, 100)
	return value <= threshold
}

// ParamChanges defines the parameters that can be modified by param change proposals
func ParamChanges(r *rand.Rand) []simulation.ParamChange {
	return []simulation.ParamChange{}
}
