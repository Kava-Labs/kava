package cdp

import (
	"errors"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/x/cdp/keeper"
	"github.com/kava-labs/kava/x/cdp/types"
	pricefeedtypes "github.com/kava-labs/kava/x/pricefeed/types"
)

// BeginBlocker compounds the debt in outstanding cdps and liquidates cdps that are below the required collateralization ratio
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	params := k.GetParams(ctx)

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

		err = k.SynchronizeInterestForRiskyCDPs(ctx, cp.CheckCollateralizationIndexCount, sdk.MaxSortableDec, cp.Type)
		if err != nil {
			panic(err)
		}

		err = k.LiquidateCdps(ctx, cp.LiquidationMarketID, cp.Type, cp.LiquidationRatio, cp.CheckCollateralizationIndexCount)
		if err != nil && !errors.Is(err, pricefeedtypes.ErrNoValidPrice) {
			panic(err)
		}
	}

	err := k.RunSurplusAndDebtAuctions(ctx)
	if err != nil {
		panic(err)
	}
}
