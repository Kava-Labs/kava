package app

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	evmutilkeeper "github.com/kava-labs/kava/x/evmutil/keeper"
	evmutiltypes "github.com/kava-labs/kava/x/evmutil/types"
)

const (
	MainnetUpgradeName = "v0.24.0"
	TestnetUpgradeName = "v0.24.0-alpha.0"
)

var (
	MainnetAtomDenom = "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
	TestnetHardDenom = "hard"
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

	doUpgrade := upgradeInfo.Name == MainnetUpgradeName || upgradeInfo.Name == TestnetUpgradeName
	if doUpgrade && !app.upgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{}

		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
}

func MainnetUpgradeHandler(app App) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		app.Logger().Info("running mainnet upgrade handler")

		toVM, err := app.mm.RunMigrations(ctx, app.configurator, fromVM)
		if err != nil {
			return toVM, err
		}

		app.Logger().Info("initializing allowed_cosmos_denoms param of x/evmutil")
		allowedDenoms := []evmutiltypes.AllowedCosmosCoinERC20Token{
			{
				CosmosDenom: MainnetAtomDenom,
				// erc20 contract metadata
				Name:     "ATOM",
				Symbol:   "ATOM",
				Decimals: 6,
			},
		}
		InitializeEvmutilAllowedCosmosDenoms(ctx, app.evmutilKeeper, allowedDenoms)

		return toVM, nil
	}
}

func TestnetUpgradeHandler(app App) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		app.Logger().Info("running testnet upgrade handler")

		toVM, err := app.mm.RunMigrations(ctx, app.configurator, fromVM)
		if err != nil {
			return toVM, err
		}

		// on testnet, IBC is not enabled. we initialize HARD tokens for conversion to EVM.
		allowedDenoms := []evmutiltypes.AllowedCosmosCoinERC20Token{
			{
				CosmosDenom: TestnetHardDenom,
				// erc20 contract metadata
				Name:     "HARD",
				Symbol:   "HARD",
				Decimals: 6,
			},
		}
		InitializeEvmutilAllowedCosmosDenoms(ctx, app.evmutilKeeper, allowedDenoms)

		return toVM, nil
	}
}

// InitializeEvmutilAllowedCosmosDenoms sets the AllowedCosmosDenoms parameter of the x/evmutil module.
// This new parameter controls what cosmos denoms are allowed to be converted to ERC20 tokens.
func InitializeEvmutilAllowedCosmosDenoms(
	ctx sdk.Context,
	evmutilKeeper evmutilkeeper.Keeper,
	allowedCoins []evmutiltypes.AllowedCosmosCoinERC20Token,
) {
	params := evmutilKeeper.GetParams(ctx)
	params.AllowedCosmosDenoms = allowedCoins
	evmutilKeeper.SetParams(ctx, params)
}
