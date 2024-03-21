package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmkeeper "github.com/evmos/ethermint/x/evm/keeper"
	v2 "github.com/kava-labs/kava/x/evmutil/migrations/v2"
	v3 "github.com/kava-labs/kava/x/evmutil/migrations/v3"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper    Keeper
	evmKeeper *evmkeeper.Keeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper Keeper, evmKeeper *evmkeeper.Keeper) Migrator {
	return Migrator{
		keeper:    keeper,
		evmKeeper: evmKeeper,
	}
}

// Migrate1to2 migrates from version 1 to 2.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v2.MigrateStore(ctx, m.keeper.paramSubspace)
}

func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	return v3.Migrate(ctx, m.evmKeeper)
}
