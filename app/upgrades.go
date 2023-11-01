package app

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	communitytypes "github.com/kava-labs/kava/x/community/types"
	kavadisttypes "github.com/kava-labs/kava/x/kavadist/types"
)

const (
	UpgradeName_Mainnet = "v0.25.0"
	UpgradeName_Testnet = "v0.25.0-alpha.0"
	UpgradeName_E2ETest = "v0.25.0-testing"
)

var (
	// KAVA to ukava - 6 decimals
	kavaConversionFactor = sdk.NewInt(1000_000)
	secondsPerYear       = sdk.NewInt(365 * 24 * 60 * 60)

	// 10 Million KAVA per year in staking rewards, inflation disable time 2024-01-01T00:00:00 UTC
	CommunityParams_Mainnet = communitytypes.NewParams(
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		// before switchover
		sdkmath.LegacyZeroDec(),
		// after switchover - 10M KAVA to ukava per year / seconds per year
		sdkmath.LegacyNewDec(10_000_000).
			MulInt(kavaConversionFactor).
			QuoInt(secondsPerYear),
	)

	// Testnet -- 15 Trillion KAVA per year in staking rewards, inflation disable time 2023-11-16T00:00:00 UTC
	CommunityParams_Testnet = communitytypes.NewParams(
		time.Date(2023, 11, 16, 0, 0, 0, 0, time.UTC),
		// before switchover
		sdkmath.LegacyZeroDec(),
		// after switchover
		sdkmath.LegacyNewDec(15_000_000).
			MulInt64(1_000_000). // 15M * 1M = 15T
			MulInt(kavaConversionFactor).
			QuoInt(secondsPerYear),
	)

	CommunityParams_E2E = communitytypes.NewParams(
		time.Now().Add(10*time.Second).UTC(), // relative time for testing
		sdkmath.LegacyNewDec(0),              // stakingRewardsPerSecond
		sdkmath.LegacyNewDec(1000),           // upgradeTimeSetstakingRewardsPerSecond
	)

	// ValidatorMinimumCommission is the new 5% minimum commission rate for validators
	ValidatorMinimumCommission = sdk.NewDecWithPrec(5, 2)
)

// RegisterUpgradeHandlers registers the upgrade handlers for the app.
func (app App) RegisterUpgradeHandlers() {
	app.upgradeKeeper.SetUpgradeHandler(
		UpgradeName_Mainnet,
		upgradeHandler(app, UpgradeName_Mainnet, CommunityParams_Mainnet),
	)
	app.upgradeKeeper.SetUpgradeHandler(
		UpgradeName_Testnet,
		upgradeHandler(app, UpgradeName_Testnet, CommunityParams_Testnet),
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

		//
		// Staking validator minimum commission
		//
		UpdateValidatorMinimumCommission(ctx, app)

		//
		// Community Params
		//
		app.communityKeeper.SetParams(ctx, communityParams)
		app.Logger().Info(
			"initialized x/community params",
			"UpgradeTimeDisableInflation", communityParams.UpgradeTimeDisableInflation,
			"StakingRewardsPerSecond", communityParams.StakingRewardsPerSecond,
			"UpgradeTimeSetStakingRewardsPerSecond", communityParams.UpgradeTimeSetStakingRewardsPerSecond,
		)

		//
		// Kavadist gov grant
		//
		msgGrant, err := authz.NewMsgGrant(
			app.accountKeeper.GetModuleAddress(kavadisttypes.ModuleName),        // granter
			app.accountKeeper.GetModuleAddress(govtypes.ModuleName),             // grantee
			authz.NewGenericAuthorization(sdk.MsgTypeURL(&banktypes.MsgSend{})), // authorization
			nil, // expiration
		)
		if err != nil {
			return toVM, err
		}
		_, err = app.authzKeeper.Grant(ctx, msgGrant)
		if err != nil {
			return toVM, err
		}
		app.Logger().Info("created gov grant for kavadist funds")

		//
		// Gov Quorum
		//
		govTallyParams := app.govKeeper.GetTallyParams(ctx)
		oldQuorum := govTallyParams.Quorum
		govTallyParams.Quorum = sdkmath.LegacyMustNewDecFromStr("0.2").String()
		app.govKeeper.SetTallyParams(ctx, govTallyParams)
		app.Logger().Info(fmt.Sprintf("updated tally quorum from %s to %s", oldQuorum, govTallyParams.Quorum))

		//
		// Incentive Params
		//
		UpdateIncentiveParams(ctx, app)

		return toVM, nil
	}
}

// UpdateValidatorMinimumCommission updates the commission rate for all
// validators to be at least the new min commission rate, and sets the minimum
// commission rate in the staking params.
func UpdateValidatorMinimumCommission(
	ctx sdk.Context,
	app App,
) {
	resultCount := make(map[stakingtypes.BondStatus]int)

	// Iterate over *all* validators including inactive
	app.stakingKeeper.IterateValidators(
		ctx,
		func(index int64, validator stakingtypes.ValidatorI) (stop bool) {
			// Skip if validator commission is already >= 5%
			if validator.GetCommission().GTE(ValidatorMinimumCommission) {
				return false
			}

			val, ok := validator.(stakingtypes.Validator)
			if !ok {
				panic("expected stakingtypes.Validator")
			}

			// Set minimum commission rate to 5%, when commission is < 5%
			val.Commission.Rate = ValidatorMinimumCommission
			val.Commission.UpdateTime = ctx.BlockTime()

			// Update MaxRate if necessary
			if val.Commission.MaxRate.LT(ValidatorMinimumCommission) {
				val.Commission.MaxRate = ValidatorMinimumCommission
			}

			if err := app.stakingKeeper.BeforeValidatorModified(ctx, val.GetOperator()); err != nil {
				panic(fmt.Sprintf("failed to call BeforeValidatorModified: %s", err))
			}
			app.stakingKeeper.SetValidator(ctx, val)

			// Keep track of counts just for logging purposes
			switch val.GetStatus() {
			case stakingtypes.Bonded:
				resultCount[stakingtypes.Bonded]++
			case stakingtypes.Unbonded:
				resultCount[stakingtypes.Unbonded]++
			case stakingtypes.Unbonding:
				resultCount[stakingtypes.Unbonding]++
			}

			return false
		},
	)

	app.Logger().Info(
		"updated validator minimum commission rate for all existing validators",
		stakingtypes.BondStatusBonded, resultCount[stakingtypes.Bonded],
		stakingtypes.BondStatusUnbonded, resultCount[stakingtypes.Unbonded],
		stakingtypes.BondStatusUnbonding, resultCount[stakingtypes.Unbonding],
	)

	stakingParams := app.stakingKeeper.GetParams(ctx)
	stakingParams.MinCommissionRate = ValidatorMinimumCommission
	app.stakingKeeper.SetParams(ctx, stakingParams)

	app.Logger().Info(
		"updated x/staking params minimum commission rate",
		"MinCommissionRate", stakingParams.MinCommissionRate,
	)
}

// UpdateIncentiveParams modifies the earn rewards period for bkava to be 600K KAVA per year.
func UpdateIncentiveParams(
	ctx sdk.Context,
	app App,
) {
	incentiveParams := app.incentiveKeeper.GetParams(ctx)

	// bkava annualized rewards: 600K KAVA
	newAmount := sdkmath.LegacyNewDec(600_000).
		MulInt(kavaConversionFactor).
		QuoInt(secondsPerYear).
		TruncateInt()

	for i := range incentiveParams.EarnRewardPeriods {
		if incentiveParams.EarnRewardPeriods[i].CollateralType != "bkava" {
			continue
		}

		// Update rewards per second via index
		incentiveParams.EarnRewardPeriods[i].RewardsPerSecond = sdk.NewCoins(
			sdk.NewCoin("ukava", newAmount),
		)
	}

	app.incentiveKeeper.SetParams(ctx, incentiveParams)
}
