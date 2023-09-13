package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PayRewards pays the rewards to the community pool
func (k Keeper) PayCommunityRewards(ctx sdk.Context) error {
	return nil
}

// DisableInflationAfterUpgrade disables inflation after the upgrade time
func (k Keeper) DisableInflationAfterUpgrade(ctx sdk.Context) {
	logger := k.Logger(ctx)

	// check if the upgrade time has passed
	params, found := k.GetParams(ctx)
	if !found {
		return
	}

	// skip if we have already upgraded - previousBlockTime is only set on upgrade.
	_, found = k.GetPreviousBlockTime(ctx)
	if found {
		return
	}

	blockTime := ctx.BlockTime()
	upgradeTime := params.UpgradeTimeDisableInflation

	// note: a vanilla kava chain should disable inflation after the first block, so we
	// only skip upgrade here if the upgradeTime both set and in the future.
	if blockTime.Before(upgradeTime) && !upgradeTime.IsZero() {
		return
	}

	logger.Info("disable inflation upgrade started")
	k.SetPreviousBlockTime(ctx, blockTime)

	// set x/min inflation to 0
	mintParams := k.mintKeeper.GetParams(ctx)
	mintParams.InflationMax = sdk.ZeroDec()
	mintParams.InflationMin = sdk.ZeroDec()
	k.mintKeeper.SetParams(ctx, mintParams)
	logger.Info("x/mint inflation set to 0")

	// disable kavadist inflation
	kavadistParams := k.kavadistKeeper.GetParams(ctx)
	kavadistParams.Active = false
	k.kavadistKeeper.SetParams(ctx, kavadistParams)

	logger.Info("x/kavadist inflation disabled")

	// todo: consolidate community funds (transfer from kavadist and community pool)

	logger.Info("disable inflation upgrade finished successfully!")

	return
}
