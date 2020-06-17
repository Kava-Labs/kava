package simulation

import (
	"fmt"
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/kava-labs/kava/x/auction/types"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	// Note: params are encoded to JSON before being stored in the param store. These param changes
	// update the raw values in the store so values need to be JSON. This is why values that are represented
	// as strings in JSON (such as time.Duration) have the escaped quotes.
	// TODO should we encode the values properly with ModuleCdc.MustMarshalJSON()?
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyBidDuration),
			func(r *rand.Rand) string {
				return fmt.Sprintf("%d", GenBidDuration(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyMaxAuctionDuration),
			func(r *rand.Rand) string {
				return fmt.Sprintf("%d", GenMaxAuctionDuration(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyIncrementCollateral),
			func(r *rand.Rand) string {
				return fmt.Sprintf("%d", GenIncrementCollateral(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyIncrementDebt),
			func(r *rand.Rand) string {
				return fmt.Sprintf("%d", GenIncrementDebt(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyIncrementSurplus),
			func(r *rand.Rand) string {
				return fmt.Sprintf("%d", GenIncrementSurplus(r))
			},
		),
	}
}
