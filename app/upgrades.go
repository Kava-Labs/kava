package app

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	earntypes "github.com/kava-labs/kava/x/earn/types"
	evmutiltypes "github.com/kava-labs/kava/x/evmutil/types"
	savingskeeper "github.com/kava-labs/kava/x/savings/keeper"
	savingstypes "github.com/kava-labs/kava/x/savings/types"
)

const UpgradeName = "v0.19.0"

func (app App) RegisterUpgradeHandlers() {
	app.upgradeKeeper.SetUpgradeHandler(UpgradeName,
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			// 100% inflation -> 75% inflation
			app.Logger().Info("updating x/mint params with inflation 100% -> 75%")
			UpdateCosmosMintInflation(ctx, app.mintKeeper)

			app.Logger().Info("updating x/savings params with new supported denoms")
			UpdateSavingsParams(ctx, app.savingsKeeper)

			app.Logger().Info("updating x/evmutil module account with new permissions")
			UpdateEvmutilPermissions(ctx, app.accountKeeper)

			return app.mm.RunMigrations(ctx, app.configurator, fromVM)
		},
	)

	upgradeInfo, err := app.upgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	if upgradeInfo.Name == UpgradeName && !app.upgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{
			Added: []string{
				earntypes.StoreKey,
			},
			Deleted: []string{
				"bridge",
			},
		}

		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
}

func UpdateCosmosMintInflation(ctx sdk.Context, mintKeeper mintkeeper.Keeper) {
	mintParams := mintKeeper.GetParams(ctx)
	// 0.75
	mintParams.InflationMin = sdk.NewDecWithPrec(75, 2)
	mintParams.InflationMax = sdk.NewDecWithPrec(75, 2)

	mintKeeper.SetParams(ctx, mintParams)
}

func UpdateSavingsParams(ctx sdk.Context, savingsKeeper savingskeeper.Keeper) {
	savingsParams := savingstypes.NewParams([]string{
		"ukava",
		"bkava",
		"erc20/multichain/usdc",
	})

	savingsKeeper.SetParams(ctx, savingsParams)
}

func UpdateEvmutilPermissions(ctx sdk.Context, accountKeeper authkeeper.AccountKeeper) {
	// add minter and burner permissions to evmutil
	evmutilAcc, ok := accountKeeper.GetModuleAccount(ctx, evmutiltypes.ModuleName).(*authtypes.ModuleAccount)
	if !ok {
		panic("unable to fetch evmutil module account")
	}
	evmutilAcc.Permissions = []string{
		authtypes.Minter, authtypes.Burner,
	}
	accountKeeper.SetModuleAccount(ctx, evmutilAcc)
}
