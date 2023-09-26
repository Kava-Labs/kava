package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) CheckAndDisableMintAndKavaDistInflation(ctx sdk.Context) {
	params, found := k.GetParams(ctx)
	if !found {
		// panic since this can only be reached if chain state is corrupted or method is ran at an invalid height
		panic("invalid state: module parameters not found")
	}

	// if upgrade time is in the future or zero there is nothing to do, so return
	if params.UpgradeTimeDisableInflation.IsZero() || params.UpgradeTimeDisableInflation.After(ctx.BlockTime()) {
		return
	}

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

	// reset disable inflation upgrade time
	params.UpgradeTimeDisableInflation = time.Time{}
	k.SetParams(ctx, params)
	logger.Info("disable inflation upgrade time reset")

	logger.Info("disable inflation upgrade finished successfully!")
}
