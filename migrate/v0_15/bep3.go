package v0_15

import (
	v0_15bep3 "github.com/kava-labs/kava/x/bep3/types"
)

// Bep3 migrates from a v0.14 bep3 genesis state to a v0.15 incentive genesis state
func Bep3(bep3GS v0_15bep3.GenesisState) v0_15bep3.GenesisState {
	return bep3GS
}
