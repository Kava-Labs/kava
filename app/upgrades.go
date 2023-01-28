package app

import (
	"fmt"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	communitykeeper "github.com/kava-labs/kava/x/community/keeper"
	communitytypes "github.com/kava-labs/kava/x/community/types"
)

const (
	MainnetUpgradeName = "v0.21.0"
	TestnetUpgradeName = "v0.21.0-alpha.0"
)

func (app App) RegisterUpgradeHandlers() {
	// register upgrade handler for mainnet
	app.upgradeKeeper.SetUpgradeHandler(MainnetUpgradeName, MainnetUpgradeHandler(app))
	// register upgrade handler for testnet
	app.upgradeKeeper.SetUpgradeHandler(TestnetUpgradeName, TestnetUpgradeHandler(app))

	upgradeInfo, err := app.upgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	// MAINNET STORE CHANGES
	// only the community module is added which has no store.
	// therefore, no store migration is necessary for mainnet.

	// TESTNET STORE CHANGES
	// we must undo the store changes performed in the v0.20.0-alpha.0 upgrade handler.
	if upgradeInfo.Name == TestnetUpgradeName && !app.upgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{
			Added: []string{
				minttypes.StoreKey,
			},
			Deleted: []string{
				"kavamint",
			},
		}
		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
}

// MainnetUpgradeHandler does nothing. No state changes are necessary on mainnet because v0.20.0 was
// never released and its upgrade handler was never run.
func MainnetUpgradeHandler(app App) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// no-op
		app.Logger().Info("running mainnet upgrade handler")
		return app.mm.RunMigrations(ctx, app.configurator, fromVM)
	}
}

// TestnetUpgradeHandler is the inverse of the v0.20.0-alpha.0 upgrade handler that was run on public
// testnet. It reverts the state changes to bring the chain back to its v0.19.0 state, which is expected
// in this upgrade.
// See original handler here: https://github.com/Kava-Labs/kava/blob/v0.20.0-alpha.0/app/upgrades.go#L65
func TestnetUpgradeHandler(app App) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		app.Logger().Info("running testnet upgrade handler")

		// move community pool funds back to community pool from community module.
		app.Logger().Info("migrating community pool funds")
		MigrateCommunityPoolFunds(ctx, app.accountKeeper, app.communityKeeper, app.distrKeeper)

		// reenable community tax
		app.Logger().Info("re-enabling community tax")
		ReenableCommunityTax(ctx, app.distrKeeper)

		// run migrations
		vm, err := app.mm.RunMigrations(ctx, app.configurator, fromVM)
		if err != nil {
			panic(err)
		}

		// initialize x/mint params. must be done after migrations so module exists.
		app.Logger().Info("initializing x/mint state")
		InitializeMintParams(ctx, app.mintKeeper, app.stakingKeeper)

		return vm, nil
	}
}

// MigrateCommunityPoolFunds takes the full balance of the x/community module account and transfers them
// back to the original community pool (the auth fee pool)
// In the v0.20.0-alpha.0 upgrade handler, community pool funds were moved to the x/community module
// account. This handler transfers them back.
func MigrateCommunityPoolFunds(
	ctx sdk.Context,
	accountKeeper authkeeper.AccountKeeper,
	communityKeeper communitykeeper.Keeper,
	distKeeper distrkeeper.Keeper,
) {
	// get total balance of x/community module account
	balance := communityKeeper.GetModuleAccountBalance(ctx)

	// transfer whole balance to the community pool (auth fee pool held by x/distribution)
	communityMaccAddress := accountKeeper.GetModuleAddress(communitytypes.ModuleAccountName)
	err := distKeeper.FundCommunityPool(ctx, balance, communityMaccAddress)
	if err != nil {
		panic(fmt.Sprintf("failed to move community pool funds: %s", err))
	}
}

// ReenableCommunityTax sets x/distribution's community_tax to the value currently on mainnet.
func ReenableCommunityTax(ctx sdk.Context, distrKeeper distrkeeper.Keeper) {
	params := distrKeeper.GetParams(ctx)
	params.CommunityTax = sdk.MustNewDecFromStr("0.925000000000000000") // community tax currently present on mainnet
	distrKeeper.SetParams(ctx, params)
}

// InitializeMintParams sets up the parameters and state of x/mint.
func InitializeMintParams(
	ctx sdk.Context,
	mintKeeper mintkeeper.Keeper,
	stakingKeeper stakingkeeper.Keeper,
) {
	// init params for x/mint with values from mainnet
	params := mintKeeper.GetParams(ctx)
	params.MintDenom = stakingKeeper.BondDenom(ctx)
	params.InflationMax = sdk.MustNewDecFromStr("0.750000000000000000")
	params.InflationMin = sdk.MustNewDecFromStr("0.750000000000000000")

	mintKeeper.SetParams(ctx, params)
}
