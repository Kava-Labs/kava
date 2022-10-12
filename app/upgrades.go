package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankKeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	evmutiltypes "github.com/kava-labs/kava/x/evmutil/types"
	kavadisttypes "github.com/kava-labs/kava/x/kavadist/types"
)

const UpgradeName = "v0.19.4-testnet"

func (app App) RegisterUpgradeHandlers() {
	app.upgradeKeeper.SetUpgradeHandler(UpgradeName,
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {

			return app.mm.RunMigrations(ctx, app.configurator, fromVM)
		},
	)
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

func AddKavadistFundAccount(ctx sdk.Context, accountKeeper authkeeper.AccountKeeper, bankKeeper bankKeeper.Keeper, distKeeper distrkeeper.Keeper) {
	maccAddr := accountKeeper.GetModuleAddress(kavadisttypes.FundModuleAccount)

	accountI := accountKeeper.GetAccount(ctx, maccAddr)
	// if account already exists and is a module account, return
	_, ok := accountI.(authtypes.ModuleAccountI)
	if ok {
		return
	}
	// if account exists and is not a module account, transfer funds to community pool
	if accountI != nil {
		// transfer balance if it exists
		coins := bankKeeper.GetAllBalances(ctx, maccAddr)
		if !coins.IsZero() {
			err := distKeeper.FundCommunityPool(ctx, coins, maccAddr)
			if err != nil {
				panic(err)
			}
		}
	}
	// instantiate new module account
	modBaseAcc := authtypes.NewBaseAccount(maccAddr, nil, 0, 0)
	modAcc := authtypes.NewModuleAccount(modBaseAcc, kavadisttypes.FundModuleAccount, []string{}...)
	accountKeeper.SetModuleAccount(ctx, modAcc)

}
