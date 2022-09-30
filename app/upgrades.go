package app

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	earntypes "github.com/kava-labs/kava/x/earn/types"
)

const UpgradeName = "v0.19.0"

func (app App) RegisterUpgradeHandlers() {
	app.upgradeKeeper.SetUpgradeHandler(UpgradeName,
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			// 100% inflation -> 75% inflation
			UpdateCosmosMintInflation(ctx, app.mintKeeper)

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
