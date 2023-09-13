package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PayCommunityRewards pays rewards to the fee collector account
func (k Keeper) PayCommunityRewards(ctx sdk.Context) error {
	return nil
}

// DisableInflationAfterUpgrade disables inflation on or after the upgrade time
func (k Keeper) DisableInflationAfterUpgrade(ctx sdk.Context) {
	logger := k.Logger(ctx)

	params, found := k.GetParams(ctx)
	if !found {
		return
	}

	// skip if we have already upgraded - previousBlockTime is first set on upgrade
	_, found = k.GetPreviousBlockTime(ctx)
	if found {
		return
	}

	blockTime := ctx.BlockTime()
	upgradeTime := params.UpgradeTimeDisableInflation

	// a vanilla kava chain should disable inflation on the first block if `upgradeTime` is not set.
	// thus, we only skip upgrade here if `upgradeTime` is set and `blockTime` is before `upgradeTime`.
	if !upgradeTime.IsZero() && blockTime.Before(upgradeTime) {
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
