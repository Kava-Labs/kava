package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	v2 "github.com/kava-labs/kava/x/evmutil/migrations/v2"
	v3 "github.com/kava-labs/kava/x/evmutil/migrations/v3"
	"github.com/kava-labs/kava/x/evmutil/types"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper            Keeper
	preciseBankKeeper types.PreciseBankKeeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(
	keeper Keeper,
	preciseBankKeeper types.PreciseBankKeeper,
) Migrator {
	return Migrator{
		keeper:            keeper,
		preciseBankKeeper: preciseBankKeeper,
	}
}

// Migrate1to2 migrates from version 1 to 2.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v2.MigrateStore(ctx, m.keeper.paramSubspace)
}

// Migrate1to2 migrates from version 2 to 3.
func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	return v3.MigrateStore(ctx, m.keeper.cdc, m.keeper.storeKey, m.preciseBankKeeper)
}
