package app

// This is purely an example to illustrate how to e2e test an upgrade handler.
// An upgrade handler named "example-upgrade" is registered that when run does the following:
// - Arbitrarily moves 1234ukava from the community pool to the community module account
// - Arbitrarily sets a parameter. In this case it makes ukava the only AllowedVault in x/earn.
// In the e2e tests, with the properly configured env variables, the

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	communitytypes "github.com/kava-labs/kava/x/community/types"
	earnkeeper "github.com/kava-labs/kava/x/earn/keeper"
	earntypes "github.com/kava-labs/kava/x/earn/types"
	kavadisttypes "github.com/kava-labs/kava/x/kavadist/types"
)

const UpgradeName = "example-upgrade"

var (
	MovedCommunityPoolFunds = sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1234)))
)

func (app App) RegisterUpgradeHandlers() {
	app.upgradeKeeper.SetUpgradeHandler(UpgradeName, ExampleUpgradeHandler(app))
	upgradeInfo, err := app.upgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	if upgradeInfo.Name == UpgradeName && !app.upgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storetypes.StoreUpgrades{}))
	}
}

// NOTE: this should never be run on a live network. It's purely an example of doing random things
// that we can check in an e2e test of an upgraded network.
func ExampleUpgradeHandler(app App) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// some example operation we can look for to prove the upgrade happened...
		// move an arbitrary amount of funds from the community pool to the community module account
		app.Logger().Info("move some community pool funds to community module account")
		MoveCommunityPoolFunds(ctx, MovedCommunityPoolFunds, app.distrKeeper, app.bankKeeper)

		// set an arbitrary param that we can check pre/post upgrade
		app.Logger().Info("updating earn parameters")
		UpdateEarnParams(ctx, app.earnKeeper)

		vm, err := app.mm.RunMigrations(ctx, app.configurator, fromVM)
		if err != nil {
			panic(err)
		}

		return vm, nil
	}
}

// arbitrary balance manipulation
func MoveCommunityPoolFunds(
	ctx sdk.Context,
	amount sdk.Coins,
	distKeeper distrkeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
) {
	// get balance of original community pool
	balance, leftoverDust := distKeeper.GetFeePoolCommunityCoins(ctx).TruncateDecimal()
	leftover := balance.Sub(amount)
	if !leftover.IsValid() {
		panic("moving too much. community pool lacks funds.")
	}

	// the balance of the community fee pool is held by the distribution module.
	// transfer the desired amount to the community module account.
	// post-upgrade this should be the only balance of the community module and we should be able to
	// check that before the upgrade height the funds were in the community pool.
	err := bankKeeper.SendCoinsFromModuleToModule(
		ctx,
		distrtypes.ModuleName,
		communitytypes.ModuleAccountName,
		amount,
	)
	if err != nil {
		panic(err)
	}

	// update the community pool's balance of the fee pool in x/distribution.
	feePool := distKeeper.GetFeePool(ctx)
	feePool.CommunityPool = sdk.NewDecCoinsFromCoins(leftover...).Add(leftoverDust...)
	distKeeper.SetFeePool(ctx, feePool)
}

// arbitrary param change: removes all the earn vaults except for ukava
// should be able to check the difference before/after upgrade in an e2e test.
func UpdateEarnParams(ctx sdk.Context, earnKeeper earnkeeper.Keeper) {
	earnParams := earntypes.NewParams(
		earntypes.AllowedVaults{
			// ukava - Community Pool
			earntypes.NewAllowedVault(
				"ukava",
				earntypes.StrategyTypes{earntypes.STRATEGY_TYPE_SAVINGS},
				true,
				[]sdk.AccAddress{authtypes.NewModuleAddress(kavadisttypes.FundModuleAccount)},
			),
		})

	if err := earnParams.Validate(); err != nil {
		panic(err)
	}

	earnKeeper.SetParams(ctx, earnParams)
}
