package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PayCommunityRewards pays rewards to the fee collector account
func (k Keeper) PayCommunityRewards(ctx sdk.Context) error {
	return nil
}

// ShouldStartDisableInflationUpgrade returns true if the disable inflation upgrade should be started
func (k Keeper) ShouldStartDisableInflationUpgrade(ctx sdk.Context) bool {
	params, found := k.GetParams(ctx)
	if !found {
		return false
	}

	// skip if we have already upgraded - previousBlockTime is first set on upgrade
	_, found = k.GetPreviousBlockTime(ctx)
	if found {
		return false
	}

	blockTime := ctx.BlockTime()
	upgradeTime := params.UpgradeTimeDisableInflation

	// a vanilla kava chain should disable inflation on the first block if `upgradeTime` is not set.
	// thus, we don't upgrade if `upgradeTime` is set and `blockTime` is before `upgradeTime`.
	if !upgradeTime.IsZero() && blockTime.Before(upgradeTime) {
		return false
	}

	return true
}

// StartDisableInflationUpgrade disables x/mint and x/kavadist inflation
func (k Keeper) StartDisableInflationUpgrade(ctx sdk.Context) {
	logger := k.Logger(ctx)

	blockTime := ctx.BlockTime()

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
}
