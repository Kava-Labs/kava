package app

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	evmutiltypes "github.com/kava-labs/kava/x/evmutil/types"
	precisebankkeeper "github.com/kava-labs/kava/x/precisebank/keeper"
	precisebanktypes "github.com/kava-labs/kava/x/precisebank/types"
)

const (
	UpgradeName_Mainnet = "v0.27.0"
	UpgradeName_Testnet = "v0.27.0-alpha.0"
)

// RegisterUpgradeHandlers registers the upgrade handlers for the app.
func (app App) RegisterUpgradeHandlers() {
	app.upgradeKeeper.SetUpgradeHandler(
		UpgradeName_Mainnet,
		upgradeHandler(app, UpgradeName_Mainnet),
	)
	app.upgradeKeeper.SetUpgradeHandler(
		UpgradeName_Testnet,
		upgradeHandler(app, UpgradeName_Testnet),
	)

	upgradeInfo, err := app.upgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	doUpgrade := upgradeInfo.Name == UpgradeName_Mainnet ||
		upgradeInfo.Name == UpgradeName_Testnet

	if doUpgrade && !app.upgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{
			Added: []string{
				precisebanktypes.ModuleName,
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
) upgradetypes.UpgradeHandler {
	return func(
		ctx sdk.Context,
		plan upgradetypes.Plan,
		fromVM module.VersionMap,
	) (module.VersionMap, error) {
		app.Logger().Info(fmt.Sprintf("running %s upgrade handler", name))

		// Run migrations for all modules and return new consensus version map.
		versionMap, err := app.mm.RunMigrations(ctx, app.configurator, fromVM)
		if err != nil {
			return nil, err
		}

		// Migrate fractional balance reserve from x/evmutil to x/precisebank.
		// This should be done **after** store migrations are completed in
		// app.mm.RunMigrations, which migrates fractional balances to
		// x/precisebank.
		if err := MigrateFractionalBalanceReserve(
			ctx,
			app.Logger(),
			app.accountKeeper,
			app.bankKeeper,
			app.precisebankKeeper,
		); err != nil {
			return nil, err
		}

		return versionMap, nil
	}
}

func MigrateFractionalBalanceReserve(
	ctx sdk.Context,
	logger log.Logger,
	accountKeeper evmutiltypes.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	precisebankKeeper precisebankkeeper.Keeper,
) error {
	// Transfer x/evmutil reserve to x/precisebank.
	moduleAddr := accountKeeper.GetModuleAddress(evmutiltypes.ModuleName)
	reserveBalance := bankKeeper.GetBalance(ctx, moduleAddr, precisebanktypes.IntegerCoinDenom)
	if err := bankKeeper.SendCoinsFromModuleToModule(
		ctx,
		evmutiltypes.ModuleName,
		precisebanktypes.ModuleName,
		sdk.NewCoins(reserveBalance),
	); err != nil {
		return fmt.Errorf("failed to transfer x/evmutil reserve to x/precisebank: %w", err)
	}

	logger.Info(fmt.Sprintf("transferred reserve balance: %s", reserveBalance))

	// Ensure x/precisebank reserve fully backs all fractional balances.
	totalFractionalBalances := precisebankKeeper.GetTotalSumFractionalBalances(ctx)

	// Ensure state is correct:
	// - Non-zero total: Balances have been migrated to x/precisebank
	// - Multiple of conversion factor
	if totalFractionalBalances.IsZero() {
		return fmt.Errorf("invalid state, total fractional balances should not be zero")
	}

	if !totalFractionalBalances.Mod(precisebanktypes.ConversionFactor()).IsZero() {
		return fmt.Errorf(
			"invalid state, total fractional balances should be a multiple of the conversion factor but is %s",
			totalFractionalBalances,
		)
	}

	// Determine how much the reserve is off by, e.g. unbacked amount
	expectedReserveBalance := totalFractionalBalances.Quo(precisebanktypes.ConversionFactor())
	unbackedAmount := expectedReserveBalance.Sub(reserveBalance.Amount)
	logger.Info(fmt.Sprintf("total account fractional balances: %s", totalFractionalBalances))

	if unbackedAmount.IsPositive() {
		logger.Info(fmt.Sprintf("unbacked amount to be minted: %s", unbackedAmount))

		// Reserve is not enough, we can mint some
		// Mint the unbacked amount
		coins := sdk.NewCoins(sdk.NewCoin(precisebanktypes.IntegerCoinDenom, unbackedAmount))
		if err := bankKeeper.MintCoins(ctx, precisebanktypes.ModuleName, coins); err != nil {
			return fmt.Errorf("failed to mint extra reserve coins: %w", err)
		}
	} else if unbackedAmount.IsNegative() {
		logger.Info(fmt.Sprintf("extra reserve amount to be burned: %s", unbackedAmount.Neg()))

		// Reserve is too much, we can burn some
		// Burn the unbacked amount
		coins := sdk.NewCoins(sdk.NewCoin(precisebanktypes.IntegerCoinDenom, unbackedAmount.Neg()))
		if err := bankKeeper.BurnCoins(ctx, precisebanktypes.ModuleName, coins); err != nil {
			return fmt.Errorf("failed to burn extra reserve coins: %w", err)
		}
	} else {
		logger.Info("no extra reserve coins needed")
	}

	return nil
}
