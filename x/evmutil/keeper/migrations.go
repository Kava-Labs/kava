package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	v2 "github.com/kava-labs/kava/x/evmutil/migrations/v2"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	legacySubspace paramtypes.Subspace
}

// NewMigrator returns a new Migrator.
func NewMigrator(ss paramtypes.Subspace) Migrator {
	return Migrator{legacySubspace: ss}
}

// Migrate1to2 migrates from version 1 to 2.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v2.MigrateStore(ctx, m.legacySubspace)
}
