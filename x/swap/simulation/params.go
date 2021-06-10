package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/kava-labs/kava/x/swap/types"
)

const (
	keyAllowedPools = "AllowedPools"
	keySwapFee      = "SwapFee"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(r *rand.Rand) []simulation.ParamChange {
	return []simulation.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, keyAllowedPools,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenAllowedPools(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, keySwapFee,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenSwapFee(r))
			},
		),
	}
}
