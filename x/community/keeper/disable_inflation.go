package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/community/types"
)

// CheckAndDisableMintAndKavaDistInflation compares the disable inflation time and block time,
// and disables inflation if time is set and before block time.  Inflation time is reset,
// so this method is safe to call more than once.
func (k Keeper) CheckAndDisableMintAndKavaDistInflation(ctx sdk.Context) {
	// panic if params are not found  since this can only be reached if chain state is corrupted or method is ran at an invalid height
	params := k.mustGetParams(ctx)

	// if disable inflation time is in the future or zero there is nothing to do, so return
	if params.UpgradeTimeDisableInflation.IsZero() || params.UpgradeTimeDisableInflation.After(ctx.BlockTime()) {
		return
	}

	logger := k.Logger(ctx)
	logger.Info("disable inflation upgrade started")

	// run disable inflation logic
	k.disableInflation(ctx)
	k.disableCommunityTax(ctx)

	logger.Info("disable inflation upgrade finished successfully!")

	// reset disable inflation time to ensure next call is a no-op
	params.UpgradeTimeDisableInflation = time.Time{}
	// set staking rewards to provided initial value
	params.StakingRewardsPerSecond = params.UpgradeTimeSetStakingRewardsPerSecond
	k.SetParams(ctx, params)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeInflationStop,
			sdk.NewAttribute(
				types.AttributeKeyInflationDisableTime,
				ctx.BlockTime().Format(time.RFC3339),
			),
		),
	)

	if err := k.StartCommunityFundConsolidation(ctx); err != nil {
		panic(err)
	}
}

// TODO: double check this is correct method for disabling inflation in kavadist without
// affecting rewards.  In addition, inflation periods in kavadist should be removed.
func (k Keeper) disableInflation(ctx sdk.Context) {
	logger := k.Logger(ctx)

	// set x/min inflation to 0
	mintParams := k.mintKeeper.GetParams(ctx)
	mintParams.InflationMin = sdk.ZeroDec()
	mintParams.InflationMax = sdk.ZeroDec()
	if err := k.mintKeeper.SetParams(ctx, mintParams); err != nil {
		panic(err)
	}
	logger.Info("x/mint inflation set to 0")

	// disable kavadist inflation
	kavadistParams := k.kavadistKeeper.GetParams(ctx)
	kavadistParams.Active = false
	k.kavadistKeeper.SetParams(ctx, kavadistParams)
	logger.Info("x/kavadist inflation disabled")
}

// disableCommunityTax sets x/distribution Params.CommunityTax to 0
func (k Keeper) disableCommunityTax(ctx sdk.Context) {
	logger := k.Logger(ctx)

	distrParams := k.distrKeeper.GetParams(ctx)
	distrParams.CommunityTax = sdk.ZeroDec()
	if err := k.distrKeeper.SetParams(ctx, distrParams); err != nil {
		panic(err)
	}
	logger.Info("x/distribution community tax set to 0")
}
