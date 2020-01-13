package cdp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// BeginBlock compounds the debt in outstanding cdps and liquidates cdps that are below the required collateralization ratio
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k Keeper) {
	params := k.GetParams(ctx)
	previousBlockTime, found := k.GetPreviousBlockTime(ctx)
	if !found {
		previousBlockTime = ctx.BlockTime()
	}
	timeElapsed := sdk.NewInt(ctx.BlockTime().Unix() - previousBlockTime.Unix())
	for _, cp := range params.CollateralParams {
		for _, dp := range params.DebtParams {
			k.HandleNewDebt(ctx, cp.Denom, dp.Denom, timeElapsed)
		}

		k.LiquidateCdps(ctx, cp.MarketID, cp.Denom, cp.LiquidationRatio)
	}
	k.HandleSurplusAndDebtAuctions(ctx)
	k.SetPreviousBlockTime(ctx, ctx.BlockTime())
	return
}
