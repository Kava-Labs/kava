package app

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	communitykeeper "github.com/kava-labs/kava/x/community/keeper"
	communitytypes "github.com/kava-labs/kava/x/community/types"
)

const (
	UpgradeName_Mainnet = "v0.25.0"
	UpgradeName_Testnet = "v0.25.0-alpha.0"
	UpgradeName_E2ETest = "v0.25.0-testing"
)

var CommunityParams_E2E = communitytypes.NewParams(
	time.Now().Add(10*time.Second).UTC(), // relative time for testing
	sdkmath.LegacyNewDec(0),              // stakingRewardsPerSecond
	sdkmath.LegacyNewDec(1000),           // upgradeTimeSetstakingRewardsPerSecond
)

// RegisterUpgradeHandlers registers the upgrade handlers for the app.
func (app App) RegisterUpgradeHandlers() {
	app.upgradeKeeper.SetUpgradeHandler(
		UpgradeName_Mainnet,
		upgradeHandler(app, UpgradeName_Mainnet, communitytypes.DefaultParams()),
	)
	app.upgradeKeeper.SetUpgradeHandler(
		UpgradeName_Testnet,
		upgradeHandler(app, UpgradeName_Testnet, communitytypes.DefaultParams()),
	)
	app.upgradeKeeper.SetUpgradeHandler(
		UpgradeName_E2ETest,
		upgradeHandler(app, UpgradeName_Testnet, CommunityParams_E2E),
	)

	upgradeInfo, err := app.upgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	doUpgrade := upgradeInfo.Name == UpgradeName_Mainnet ||
		upgradeInfo.Name == UpgradeName_Testnet ||
		upgradeInfo.Name == UpgradeName_E2ETest

	if doUpgrade && !app.upgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{
			Added: []string{
				// x/community added store
				communitytypes.ModuleName,
			},
		}

		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
}

// upgradeHandler returns an UpgradeHandler for the given upgrade parameters.
func upgradeHandler(
	app App,
	name string,
	communityParams communitytypes.Params,
) upgradetypes.UpgradeHandler {
	return func(
		ctx sdk.Context,
		plan upgradetypes.Plan,
		fromVM module.VersionMap,
	) (module.VersionMap, error) {
		app.Logger().Info(fmt.Sprintf("running %s upgrade handler", name))

		toVM, err := app.mm.RunMigrations(ctx, app.configurator, fromVM)
		if err != nil {
			return toVM, err
		}

		app.Logger().Info("initializing x/community params")
		InitializeCommunityParams(ctx, app.communityKeeper, communityParams)

		return toVM, nil
	}
}

// InitializeCommunityParams sets the community params in the store, first
// checking that they are not already set.
func InitializeCommunityParams(
	ctx sdk.Context,
	communityKeeper communitykeeper.Keeper,
	params communitytypes.Params,
) {
	_, found := communityKeeper.GetParams(ctx)
	if found {
		panic("x/community params already set")
	}

	// SetParams validates the params
	communityKeeper.SetParams(ctx, params)
}
