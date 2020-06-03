package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/kava-labs/kava/x/bep3/types"
)

const (
	keyBnbDeputyAddress  = "BnbDeputyAddress"
	keyBnbDeputyFixedFee = "BnbDeputyFixedFee"
	keyMinAmount         = "MinAmount"
	keyMaxAmount         = "MaxAmount"
	keyMinBlockLock      = "MinBlockLock"
	keyMaxBlockLock      = "MaxBlockLock"
	keySupportedAssets   = "SupportedAssets"
)

// ParamChanges defines the parameters that can be modified by param change proposals
func ParamChanges(r *rand.Rand) []simulation.ParamChange {
	// We generate MinBlockLock first because the result is required by GenMaxBlockLock()
	minBlockLockVal := GenMinBlockLock(r)
	minAmount := GenMinAmount(r)

	return []simulation.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, keyBnbDeputyAddress,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenRandBnbDeputy(r).Address)
			},
		),
		simulation.NewSimParamChange(types.ModuleName, keyBnbDeputyFixedFee,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%d\"", GenRandBnbDeputyFixedFee(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, keyMinAmount,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%d\"", minAmount)
			},
		),
		simulation.NewSimParamChange(types.ModuleName, keyMaxAmount,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%d\"", GenMaxAmount(r, minAmount))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, keyMinBlockLock,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%d\"", minBlockLockVal)
			},
		),
		simulation.NewSimParamChange(types.ModuleName, keyMaxBlockLock,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%d\"", GenMaxBlockLock(r, minBlockLockVal))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, keySupportedAssets,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%v\"", GenSupportedAssets(r))
			},
		),
	}
}
