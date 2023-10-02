package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
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

	// run disable inflation logic
	k.disableInflation(ctx)

	// reset disable inflation time to ensure next call is a no-op
	params.UpgradeTimeDisableInflation = time.Time{}
	// set staking rewards to provided intial value
	params.StakingRewardsPerSecond = params.UpgradeTimeSetStakingRewardsPerSecond
	k.SetParams(ctx, params)

}

// TODO: double check this is correct method for disabling inflation in kavadist without
// affecting rewards.  In addition, inflation periods in kavadist should be removed.
func (k Keeper) disableInflation(ctx sdk.Context) {
	logger := k.Logger(ctx)
	logger.Info("disable inflation upgrade started")

	// set x/min inflation to 0
	mintParams := k.mintKeeper.GetParams(ctx)
	mintParams.InflationMin = sdk.ZeroDec()
	mintParams.InflationMax = sdk.ZeroDec()
	k.mintKeeper.SetParams(ctx, mintParams)
	logger.Info("x/mint inflation set to 0")

	// disable kavadist inflation
	kavadistParams := k.kavadistKeeper.GetParams(ctx)
	kavadistParams.Active = false
	k.kavadistKeeper.SetParams(ctx, kavadistParams)
	logger.Info("x/kavadist inflation disabled")

	logger.Info("disable inflation upgrade finished successfully!")
}
