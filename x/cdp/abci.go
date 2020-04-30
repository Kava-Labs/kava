package cdp

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/cdp/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// BeginBlocker compounds the debt in outstanding cdps and liquidates cdps that are below the required collateralization ratio
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k Keeper) {
	params := k.GetParams(ctx)

	previousDistTime, found := k.GetPreviousSavingsDistribution(ctx)
	if !found {
		previousDistTime = ctx.BlockTime()
		k.SetPreviousSavingsDistribution(ctx, previousDistTime)
	}

	for _, cp := range params.CollateralParams {

		err := k.UpdateFeesForAllCdps(ctx, cp.Denom)

		// handle if an error is returned then propagate up
		if err != nil {
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					EventTypeBeginBlockerFatal,
					sdk.NewAttribute(sdk.AttributeKeyModule, fmt.Sprintf("%s", ModuleName)),
					sdk.NewAttribute(types.AttributeKeyError, fmt.Sprintf("%s", err)),
				),
			)
		}

		err = k.LiquidateCdps(ctx, cp.MarketID, cp.Denom, cp.LiquidationRatio)
		if err != nil {
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					EventTypeBeginBlockerFatal,
					sdk.NewAttribute(sdk.AttributeKeyModule, fmt.Sprintf("%s", ModuleName)),
					sdk.NewAttribute(types.AttributeKeyError, fmt.Sprintf("%s", err)),
				),
			)
		}
	}
	err := k.RunSurplusAndDebtAuctions(ctx)
	if err != nil {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				EventTypeBeginBlockerFatal,
				sdk.NewAttribute(sdk.AttributeKeyModule, fmt.Sprintf("%s", ModuleName)),
				sdk.NewAttribute(types.AttributeKeyError, fmt.Sprintf("%s", err)),
			),
		)
	}
	distTimeElapsed := sdk.NewInt(ctx.BlockTime().Unix() - previousDistTime.Unix())
	if distTimeElapsed.GTE(sdk.NewInt(int64(params.SavingsDistributionFrequency.Seconds()))) {
		err := k.DistributeSavingsRate(ctx, params.DebtParam.Denom)
		if err != nil {
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					EventTypeBeginBlockerFatal,
					sdk.NewAttribute(sdk.AttributeKeyModule, fmt.Sprintf("%s", ModuleName)),
					sdk.NewAttribute(types.AttributeKeyError, fmt.Sprintf("%s", err)),
				),
			)
		}
		k.SetPreviousSavingsDistribution(ctx, ctx.BlockTime())
	}
	return
}
