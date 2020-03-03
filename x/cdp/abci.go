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
	previousBlockTime, found := k.GetPreviousBlockTime(ctx)
	if !found {
		previousBlockTime = ctx.BlockTime()
	}
	previousDistTime, found := k.GetPreviousSavingsDistribution(ctx)
	if !found {
		previousDistTime = ctx.BlockTime()
		k.SetPreviousSavingsDistribution(ctx, previousDistTime)
	}
	blockTimeElapsed := sdk.NewInt(ctx.BlockTime().Unix() - previousBlockTime.Unix())
	for _, cp := range params.CollateralParams {
		for _, dp := range params.DebtParams {
			k.HandleNewDebt(ctx, cp.Denom, dp.Denom, blockTimeElapsed)
		}

		err := k.LiquidateCdps(ctx, cp.MarketID, cp.Denom, cp.LiquidationRatio)
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
		for _, dp := range params.DebtParams {
			err := k.DistributeSavingsRate(ctx, dp.Denom)
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
		k.SetPreviousSavingsDistribution(ctx, ctx.BlockTime())
	}
	k.SetPreviousBlockTime(ctx, ctx.BlockTime())
	return
}
