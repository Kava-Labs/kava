package app

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	evmutilkeeper "github.com/kava-labs/kava/x/evmutil/keeper"
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
		logger := app.Logger()
		logger.Info(fmt.Sprintf("running %s upgrade handler", name))

		// Run migrations for all modules and return new consensus version map.
		versionMap, err := app.mm.RunMigrations(ctx, app.configurator, fromVM)
		if err != nil {
			return nil, err
		}

		logger.Info("completed store migrations")

		// Migration of fractional balances from x/evmutil to x/precisebank
		if err := MigrateEvmutilToPrecisebank(
			ctx,
			app.accountKeeper,
			app.bankKeeper,
			app.evmutilKeeper,
			app.precisebankKeeper,
		); err != nil {
			return nil, err
		}

		return versionMap, nil
	}
}

// MigrateEvmutilToPrecisebank migrates all required state from x/evmutil to
// x/precisebank and ensures the resulting state is correct.
// This migrates the following state:
// - Fractional balances
// - Fractional balance reserve
// Initializes the following state in x/precisebank:
// - Remainder amount
func MigrateEvmutilToPrecisebank(
	ctx sdk.Context,
	accountKeeper evmutiltypes.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	evmutilKeeper evmutilkeeper.Keeper,
	precisebankKeeper precisebankkeeper.Keeper,
) error {
	logger := ctx.Logger()

	aggregateSum, err := TransferFractionalBalances(
		ctx,
		evmutilKeeper,
		precisebankKeeper,
	)
	if err != nil {
		return fmt.Errorf("fractional balances transfer: %w", err)
	}
	logger.Info(
		"fractional balances transferred from x/evmutil to x/precisebank",
		"aggregate sum", aggregateSum,
	)

	remainder := InitializeRemainder(ctx, precisebankKeeper, aggregateSum)
	logger.Info("remainder amount initialized in x/precisebank", "remainder", remainder)

	// Migrate fractional balance reserve from x/evmutil to x/precisebank.
	// This should be done **after** store migrations are completed in
	// app.mm.RunMigrations, which migrates fractional balances to
	// x/precisebank.
	if err := TransferFractionalBalanceReserve(
		ctx,
		accountKeeper,
		bankKeeper,
		precisebankKeeper,
	); err != nil {
		return fmt.Errorf("reserve transfer: %w", err)
	}

	return nil
}

// TransferFractionalBalances migrates fractional balances from x/evmutil to
// x/precisebank. It sets the fractional balance in x/precisebank and deletes
// the account from x/evmutil. Returns the aggregate sum of all fractional
// balances.
func TransferFractionalBalances(
	ctx sdk.Context,
	evmutilKeeper evmutilkeeper.Keeper,
	precisebankKeeper precisebankkeeper.Keeper,
) (sdkmath.Int, error) {
	aggregateSum := sdkmath.ZeroInt()

	var iterErr error

	evmutilKeeper.IterateAllAccounts(ctx, func(acc evmutiltypes.Account) bool {
		// Set account balance in x/precisebank
		precisebankKeeper.SetFractionalBalance(ctx, acc.Address, acc.Balance)

		// Delete account from x/evmutil
		iterErr := evmutilKeeper.SetAccount(ctx, evmutiltypes.Account{
			Address: acc.Address,
			// Set balance to 0 to delete it
			Balance: sdkmath.ZeroInt(),
		})

		// Halt iteration if there was an error
		if iterErr != nil {
			return true
		}

		// Aggregate sum of all fractional balances
		aggregateSum = aggregateSum.Add(acc.Balance)

		// Continue iterating
		return false
	})

	return aggregateSum, iterErr
}

// InitializeRemainder initializes the remainder amount in x/precisebank. It
// calculates the remainder amount that is needed to ensure that the sum of all
// fractional balances is a multiple of the conversion factor. The remainder
// amount is stored in the store and returned.
func InitializeRemainder(
	ctx sdk.Context,
	precisebankKeeper precisebankkeeper.Keeper,
	aggregateSum sdkmath.Int,
) sdkmath.Int {
	// Extra fractional coins that exceed the conversion factor.
	// This extra + remainder should equal the conversion factor to ensure
	// (sum(fBalances) + remainder) % conversionFactor = 0
	extraFractionalAmount := aggregateSum.Mod(precisebanktypes.ConversionFactor())
	remainder := precisebanktypes.ConversionFactor().
		Sub(extraFractionalAmount).
		// Mod conversion factor to ensure remainder is valid.
		// If extraFractionalAmount is a multiple of conversion factor, the
		// remainder is 0.
		Mod(precisebanktypes.ConversionFactor())

	// Panics if the remainder is invalid. In a correct chain state and only
	// mint/burns due to transfers, this would be 0.
	precisebankKeeper.SetRemainderAmount(ctx, remainder)

	return remainder
}

// TransferFractionalBalanceReserve migrates the fractional balance reserve from
// x/evmutil to x/precisebank. It transfers the reserve balance from x/evmutil
// to x/precisebank and ensures that the reserve fully backs all fractional
// balances. It mints or burns coins to back the fractional balances exactly.
func TransferFractionalBalanceReserve(
	ctx sdk.Context,
	accountKeeper evmutiltypes.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	precisebankKeeper precisebankkeeper.Keeper,
) error {
	logger := ctx.Logger()

	// Transfer x/evmutil reserve to x/precisebank.
	evmutilAddr := accountKeeper.GetModuleAddress(evmutiltypes.ModuleName)
	reserveBalance := bankKeeper.GetBalance(ctx, evmutilAddr, precisebanktypes.IntegerCoinDenom)

	if err := bankKeeper.SendCoinsFromModuleToModule(
		ctx,
		evmutiltypes.ModuleName,     // from x/evmutil
		precisebanktypes.ModuleName, // to x/precisebank
		sdk.NewCoins(reserveBalance),
	); err != nil {
		return fmt.Errorf("failed to transfer reserve from x/evmutil to x/precisebank: %w", err)
	}

	logger.Info(fmt.Sprintf("transferred reserve balance: %s", reserveBalance))

	// Ensure x/precisebank reserve fully backs all fractional balances.
	totalFractionalBalances := precisebankKeeper.GetTotalSumFractionalBalances(ctx)

	// Does NOT ensure state is correct, total fractional balances should be a
	// multiple of conversion factor but is not guaranteed due to the remainder.
	// Remainder initialization is handled by InitializeRemainder.

	// Determine how much the reserve is off by, e.g. unbacked amount
	expectedReserveBalance := totalFractionalBalances.Quo(precisebanktypes.ConversionFactor())

	// If there is a remainder (totalFractionalBalances % conversionFactor != 0),
	// then expectedReserveBalance is rounded up to the nearest integer.
	if totalFractionalBalances.Mod(precisebanktypes.ConversionFactor()).IsPositive() {
		expectedReserveBalance = expectedReserveBalance.Add(sdkmath.OneInt())
	}

	unbackedAmount := expectedReserveBalance.Sub(reserveBalance.Amount)
	logger.Info(fmt.Sprintf("total account fractional balances: %s", totalFractionalBalances))

	// Three possible cases:
	// 1. Reserve is not enough, mint coins to back the fractional balances
	// 2. Reserve is too much, burn coins to back the fractional balances exactly
	// 3. Reserve is exactly enough, no action needed
	if unbackedAmount.IsPositive() {
		coins := sdk.NewCoins(sdk.NewCoin(precisebanktypes.IntegerCoinDenom, unbackedAmount))
		if err := bankKeeper.MintCoins(ctx, precisebanktypes.ModuleName, coins); err != nil {
			return fmt.Errorf("failed to mint extra reserve coins: %w", err)
		}

		logger.Info(fmt.Sprintf("unbacked amount minted to reserve: %s", unbackedAmount))
	} else if unbackedAmount.IsNegative() {
		coins := sdk.NewCoins(sdk.NewCoin(precisebanktypes.IntegerCoinDenom, unbackedAmount.Neg()))
		if err := bankKeeper.BurnCoins(ctx, precisebanktypes.ModuleName, coins); err != nil {
			return fmt.Errorf("failed to burn extra reserve coins: %w", err)
		}

		logger.Info(fmt.Sprintf("extra reserve amount burned: %s", unbackedAmount.Neg()))
	} else {
		logger.Info("reserve exactly backs fractional balances, no mint/burn needed")
	}

	return nil
}
