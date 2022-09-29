package migrations

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/keeper"
	v2 "github.com/kava-labs/kava/x/incentive/migrations/v2"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper keeper.Keeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper keeper.Keeper) Migrator {
	return Migrator{keeper: keeper}
}

// Migrate1to2 migrates from version 1 to 2.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v2.MigrateParams(ctx, m.keeper)
}
