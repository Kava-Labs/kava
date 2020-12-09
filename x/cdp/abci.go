package cdp

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/tendermint/tendermint/abci/types"

	pricefeedtypes "github.com/kava-labs/kava/x/pricefeed/types"
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
		ok := k.UpdatePricefeedStatus(ctx, cp.SpotMarketID)
		if !ok {
			continue
		}

		ok = k.UpdatePricefeedStatus(ctx, cp.LiquidationMarketID)
		if !ok {
			continue
		}

		err := k.AccumulateInterest(ctx, cp.Type)
		if err != nil {
			panic(err)
		}

		err = k.LiquidateCdps(ctx, cp.LiquidationMarketID, cp.Type, cp.LiquidationRatio)
		if err != nil && !errors.Is(err, pricefeedtypes.ErrNoValidPrice) {
			panic(err)
		}
	}

	err := k.RunSurplusAndDebtAuctions(ctx)
	if err != nil {
		panic(err)
	}

	distTimeElapsed := sdk.NewInt(ctx.BlockTime().Unix() - previousDistTime.Unix())
	if !distTimeElapsed.GTE(sdk.NewInt(int64(params.SavingsDistributionFrequency.Seconds()))) {
		return
	}

	err = k.DistributeSavingsRate(ctx, params.DebtParam.Denom)
	if err != nil {
		panic(err)
	}

	k.SetPreviousSavingsDistribution(ctx, ctx.BlockTime())
}
