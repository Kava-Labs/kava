package app

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	communitytypes "github.com/kava-labs/kava/x/community/types"
	kavamintkeeper "github.com/kava-labs/kava/x/kavamint/keeper"
	kavaminttypes "github.com/kava-labs/kava/x/kavamint/types"
)

const (
	MainnetUpgradeName = "v0.20.0"
	TestnetUpgradeName = "v0.20.0-alpha.0"
)

var (
	MainnetCommunityPoolInflation = sdk.NewDecWithPrec(75, 2) // 75%
	MainnetStakingRewardsApy      = sdk.NewDecWithPrec(15, 2) // 15%

	TestnetCommunityPoolInflation = sdk.OneDec()             // 100%
	TestnetStakingRewardsApy      = sdk.NewDecWithPrec(5, 2) // 5%
)

func (app App) RegisterUpgradeHandlers() {
	// register upgrade handler for mainnet
	app.upgradeKeeper.SetUpgradeHandler(MainnetUpgradeName,
		CommunityPoolAndInflationUpgradeHandler(app, MainnetCommunityPoolInflation, MainnetStakingRewardsApy),
	)
	// register upgrade handler for testnet
	app.upgradeKeeper.SetUpgradeHandler(TestnetUpgradeName,
		CommunityPoolAndInflationUpgradeHandler(app, TestnetCommunityPoolInflation, TestnetStakingRewardsApy),
	)

	upgradeInfo, err := app.upgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	doUpgrade := upgradeInfo.Name == MainnetUpgradeName || upgradeInfo.Name == TestnetUpgradeName
	if doUpgrade && !app.upgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{
			Added: []string{
				kavaminttypes.StoreKey,
			},
			Deleted: []string{
				minttypes.StoreKey,
			},
		}

		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
}

// CommunityPoolAndInflationUpgradeHandler returns an upgrade handler for migrating the community
// pool to the new x/community module account and replaces x/mint with x/kavamint
// x/kavamint is initialized with the parameters `communityPoolInflation` and `stakingRewardsApy`
func CommunityPoolAndInflationUpgradeHandler(app App, communityPoolInflation, stakingRewardsApy sdk.Dec) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// community pool goes from fee pool sub-account to x/community module account
		app.Logger().Info("transferring original community pool funds to new community pool")
		MoveCommunityPoolFunds(ctx, app.distrKeeper, app.bankKeeper)

		// community_tax -> 0%
		app.Logger().Info("disabling x/distribution community tax")
		DisableCommunityTax(ctx, app.distrKeeper)

		vm, err := app.mm.RunMigrations(ctx, app.configurator, fromVM)
		if err != nil {
			return vm, err
		}

		// initialize kavamint state after running migrations so that store exists & persists
		app.Logger().Info("initializing x/kavamint state")
		InitializeKavamintState(ctx, app.kavamintKeeper, communityPoolInflation, stakingRewardsApy)

		return vm, nil
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

// DisableCommunityTax sets x/distribution's community_tax parameter to zero.
// The vanilla community tax is no longer used because community pool inflation is set separately
// from staking rewards. No portion of the defined staking rewards is taken as a community tax.
func DisableCommunityTax(ctx sdk.Context, distrKeeper distrkeeper.Keeper) {
	params := distrKeeper.GetParams(ctx)
	params.CommunityTax = sdk.ZeroDec()
	distrKeeper.SetParams(ctx, params)
}

// InitializeKavamintState sets up the parameters and state of x/kavamint.
// The inflationary parameters are set from the args for communityPoolInflation & stakingRewardsApy
func InitializeKavamintState(
	ctx sdk.Context,
	kavamintKeeper kavamintkeeper.Keeper,
	communityPoolInflation, stakingRewardsApy sdk.Dec,
) {
	// init inflationary params for x/kavamint
	params := kavaminttypes.NewParams(communityPoolInflation, stakingRewardsApy)
	kavamintKeeper.SetParams(ctx, params)
	// set previous block time to current block's time
	kavamintKeeper.SetPreviousBlockTime(ctx, ctx.BlockTime())
}
