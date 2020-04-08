package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/kava-labs/kava/x/auction/types"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(r *rand.Rand) []simulation.ParamChange {

	return []simulation.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyBidDuration), "",
			func(r *rand.Rand) string {
				return fmt.Sprintf("%d", GenBidDuration(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyMaxAuctionDuration), "",
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%d\"", GenMaxAuctionDuration(r)) // TODO why the escaped quotes?
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyIncrementCollateral), "",
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%d\"", GenIncrementCollateral(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyIncrementDebt), "",
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%d\"", GenIncrementDebt(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyIncrementSurplus), "",
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%d\"", GenIncrementSurplus(r))
			},
		),
	}
}
