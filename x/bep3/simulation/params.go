package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/kava-labs/kava/x/bep3/types"
)

const (
	keyBnbDeputyAddress = "BnbDeputyAddress"
	keyMinBlockLock     = "MinBlockLock"
	keyMaxBlockLock     = "MaxBlockLock"
	keySupportedAssets  = "SupportedAssets"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(r *rand.Rand) []simulation.ParamChange {
	// We generate MinBlockLock first because the result is required by GenMaxBlockLock()
	minBlockLockVal := GenMinBlockLock(r)

	return []simulation.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, keyBnbDeputyAddress, "",
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenBnbDeputyAddress(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, keyMinBlockLock, "",
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%d\"", minBlockLockVal)
			},
		),
		simulation.NewSimParamChange(types.ModuleName, keyMaxBlockLock, "",
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%d\"", GenMaxBlockLock(r, minBlockLockVal))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, keySupportedAssets, "",
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%v\"", GenSupportedAssets(r))
			},
		),
	}
}
