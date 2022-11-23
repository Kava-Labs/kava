package app

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	communitytypes "github.com/kava-labs/kava/x/community/types"
	kavamintkeeper "github.com/kava-labs/kava/x/kavamint/keeper"
	kavaminttypes "github.com/kava-labs/kava/x/kavamint/types"
)

const UpgradeName = "v0.20.0"

var (
	CommunityPoolInflation = sdk.NewDecWithPrec(70, 2)
	StakingRewardsApy      = sdk.NewDecWithPrec(10, 2)
)

func (app App) RegisterUpgradeHandlers() {
	app.upgradeKeeper.SetUpgradeHandler(UpgradeName,
		func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			app.Logger().Info("transferring original community pool funds to new community pool")
			MoveCommunityPoolFunds(ctx, app.distrKeeper, app.bankKeeper)

			// min & max inflation -> 0%
			app.Logger().Info("disabling x/mint inflation")
			DisableMintInflation(ctx, app.mintKeeper)

			// community_tax -> 0%
			app.Logger().Info("disabling x/distribution community tax")
			DisableCommunityTax(ctx, app.distrKeeper)

			vm, err := app.mm.RunMigrations(ctx, app.configurator, fromVM)
			if err != nil {
				return vm, err
			}

			// initialize kavamint state after running migrations so that store exists & persists
			app.Logger().Info("initializing x/kavamint state")
			InitializeKavamintState(ctx, app.kavamintKeeper)

			return vm, nil
		},
	)

	upgradeInfo, err := app.upgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	if upgradeInfo.Name == UpgradeName && !app.upgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{
			Added: []string{
				kavaminttypes.StoreKey,
			},
		}

		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
}

// MoveCommunityPoolFunds takes the full balance of the original community pool (the auth fee pool)
// and transfers them to the new community pool (the x/community module account)
func MoveCommunityPoolFunds(
	ctx sdk.Context,
	distKeeper distrkeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
) {
	// get balance of original community pool
	balance, leftoverDust := distKeeper.GetFeePoolCommunityCoins(ctx).TruncateDecimal()

	// the balance of the community fee pool is held by the distribution module.
	// transfer whole pool balance from distribution module to new community pool module account
	err := bankKeeper.SendCoinsFromModuleToModule(
		ctx,
		distrtypes.ModuleName,
		communitytypes.ModuleAccountName,
		balance,
	)
	if err != nil {
		panic(err)
	}

	// make sure x/distribution knows that there're no funds in the community pool.
	// we keep the leftover decimal change in the account to ensure all funds are accounted for.
	feePool := distKeeper.GetFeePool(ctx)
	feePool.CommunityPool = leftoverDust
	distKeeper.SetFeePool(ctx, feePool)
}

func DisableMintInflation(ctx sdk.Context, mintKeeper mintkeeper.Keeper) {
	// set params to have min & max inflation of 0%
	params := mintKeeper.GetParams(ctx)
	params.InflationMax = sdk.ZeroDec()
	params.InflationMin = sdk.ZeroDec()
	mintKeeper.SetParams(ctx, params)

	// set minter state to reflect 0% inflation
	mintKeeper.SetMinter(ctx, minttypes.NewMinter(sdk.ZeroDec(), sdk.ZeroDec()))
}

func DisableCommunityTax(ctx sdk.Context, distrKeeper distrkeeper.Keeper) {
	params := distrKeeper.GetParams(ctx)
	params.CommunityTax = sdk.ZeroDec()
	distrKeeper.SetParams(ctx, params)
}

func InitializeKavamintState(ctx sdk.Context, kavamintKeeper kavamintkeeper.Keeper) {
	// init inflationary params for x/kavamint
	params := kavaminttypes.NewParams(CommunityPoolInflation, StakingRewardsApy)
	kavamintKeeper.SetParams(ctx, params)
	// set previous block time to current block's time
	kavamintKeeper.SetPreviousBlockTime(ctx, ctx.BlockTime())
}
