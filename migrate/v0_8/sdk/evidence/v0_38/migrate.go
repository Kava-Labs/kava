package v038

import (
	v038evidence "github.com/cosmos/cosmos-sdk/x/evidence"

	v18de63slashing "github.com/kava-labs/kava/migrate/v0_8/sdk/slashing/v18de63"
)

func Migrate(oldSlashingGenState v18de63slashing.GenesisState) v038evidence.GenesisState {
	// Need to use DefaultGenesisState as evidence doesn't export Params type (inside internal/ and missing from alias.go)

	newGenState := v038evidence.DefaultGenesisState()
	newGenState.Params.MaxEvidenceAge = oldSlashingGenState.Params.MaxEvidenceAge

	return newGenState
}
