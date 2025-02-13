package app

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

const (
	UpgradeNameMainnet = "v0.28.0"
	UpgradeNameTestnet = "v0.28.0-alpha.0"
)

// RegisterUpgradeHandlers registers the upgrade handlers for the app.
func (app App) RegisterUpgradeHandlers() {
	app.upgradeKeeper.SetUpgradeHandler(
		UpgradeNameMainnet,
		upgradeHandler(app, UpgradeNameMainnet),
	)
	app.upgradeKeeper.SetUpgradeHandler(
		UpgradeNameTestnet,
		upgradeHandler(app, UpgradeNameTestnet),
	)

	upgradeInfo, err := app.upgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	doUpgrade := upgradeInfo.Name == UpgradeNameMainnet ||
		upgradeInfo.Name == UpgradeNameTestnet

	if doUpgrade && !app.upgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{}

		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
}

// upgradeHandler returns an UpgradeHandler for the given upgrade parameters.
func upgradeHandler(
	app App,
	name string,
) upgradetypes.UpgradeHandler {
	return func(
		ctx sdk.Context,
		plan upgradetypes.Plan,
		fromVM module.VersionMap,
	) (module.VersionMap, error) {
		logger := app.Logger()
		logger.Info(fmt.Sprintf("running %s upgrade handler", name))

		// run migrations for all modules and return new consensus version map
		versionMap, err := app.mm.RunMigrations(ctx, app.configurator, fromVM)
		if err != nil {
			return nil, errorsmod.Wrap(err, "failed to run migrations")
		}

		logger.Info("completed store migrations")

		return versionMap, nil
	}
}
