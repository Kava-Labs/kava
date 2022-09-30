package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	evmutiltypes "github.com/kava-labs/kava/x/evmutil/types"
)

const UpgradeName = "v0.19.2-testnet"

func (app App) RegisterUpgradeHandlers() {
	app.upgradeKeeper.SetUpgradeHandler(UpgradeName,
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {

			// add minter and burner permissions to evmutil
			evmutilAcc, ok := app.accountKeeper.GetModuleAccount(ctx, evmutiltypes.ModuleName).(*authtypes.ModuleAccount)
			if !ok {
				panic("unable to fetch evmutil module account")
			}
			evmutilAcc.Permissions = []string{
				authtypes.Minter, authtypes.Burner,
			}
			app.accountKeeper.SetModuleAccount(ctx, evmutilAcc)

			return app.mm.RunMigrations(ctx, app.configurator, fromVM)
		},
	)
}
