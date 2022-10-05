package app

import (
	_ "embed"
	"encoding/json"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	earntypes "github.com/kava-labs/kava/x/earn/types"
	evmutiltypes "github.com/kava-labs/kava/x/evmutil/types"
	pricefeedkeeper "github.com/kava-labs/kava/x/pricefeed/keeper"
	pricefeedtypes "github.com/kava-labs/kava/x/pricefeed/types"
	savingskeeper "github.com/kava-labs/kava/x/savings/keeper"
	savingstypes "github.com/kava-labs/kava/x/savings/types"
	etherminttypes "github.com/tharsis/ethermint/types"
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

			app.Logger().Info("converting all non-contract EthAccounts to BaseAccounts")
			ConvertEOAsToBaseAccount(ctx, app.accountKeeper)

			app.Logger().Info("updating x/pricefeed params with new markets")
			UpdatePricefeedParams(ctx, app.pricefeedKeeper)

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

//go:embed eth_eoa_addresses.json
var ethEOAAddresses []byte

func IterateEOAAddresses(f func(addr string)) {
	var addresses []string

	if err := json.Unmarshal(ethEOAAddresses, &addresses); err != nil {
		panic("failed to unmarshal embedded eth_eoa_addresses.json")
	}

	for _, addr := range addresses {
		f(addr)
	}
}

// ConvertEOAsToBaseAccount converts all non-contract EthAccounts to BaseAccounts
func ConvertEOAsToBaseAccount(ctx sdk.Context, accountKeeper authkeeper.AccountKeeper) {
	IterateEOAAddresses(func(addrStr string) {
		addr, err := sdk.AccAddressFromBech32(addrStr)
		if err != nil {
			panic("failed to parse address")
		}

		// Skip non-EthAccounts
		acc := accountKeeper.GetAccount(ctx, addr)
		ethAcc, isEthAcc := acc.(*etherminttypes.EthAccount)
		if !isEthAcc {
			return
		}

		// Skip contract accounts
		if ethAcc.Type() != etherminttypes.AccountTypeEOA {
			return
		}

		// Change to BaseAccount in store
		accountKeeper.SetAccount(ctx, ethAcc.BaseAccount)
	})
}

func UpdatePricefeedParams(ctx sdk.Context, pricefeedKeeper pricefeedkeeper.Keeper) {
	params := pricefeedKeeper.GetParams(ctx)
	oracles := params.Markets[0].Oracles
	newMarkets := pricefeedtypes.Markets{
		{
			MarketID:   "usdc:usd",
			BaseAsset:  "usdc",
			QuoteAsset: "usd",
			Oracles:    oracles,
			Active:     true,
		},
		{
			MarketID:   "usdc:usd:30",
			BaseAsset:  "usdc",
			QuoteAsset: "usd",
			Oracles:    oracles,
			Active:     true,
		},
		{
			MarketID:   "usdt:usd:30",
			BaseAsset:  "usdt",
			QuoteAsset: "usd",
			Oracles:    oracles,
			Active:     true,
		},
		{
			MarketID:   "usdt:usd:30",
			BaseAsset:  "usdt",
			QuoteAsset: "usd",
			Oracles:    oracles,
			Active:     true,
		},
		{
			MarketID:   "dai:usd:30",
			BaseAsset:  "dai",
			QuoteAsset: "usd",
			Oracles:    oracles,
			Active:     true,
		},
		{
			MarketID:   "dai:usd:30",
			BaseAsset:  "dai",
			QuoteAsset: "usd",
			Oracles:    oracles,
			Active:     true,
		},
	}
	params.Markets = append(params.Markets, newMarkets...)
	pricefeedKeeper.SetParams(ctx, params)
}
