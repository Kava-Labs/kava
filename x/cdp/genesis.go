package cdp

import (
	"fmt"
	"math"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis sets initial genesis state for cdp module
func InitGenesis(ctx sdk.Context, k Keeper, pk PricefeedKeeper, data GenesisState) {
	// validate denoms - check that any collaterals in the CdpParams are in the pricefeed, pricefeed needs to initgenesis before cdp
	collateralMap := make(map[string]int)
	ap := pk.GetParams(ctx)
	for _, a := range ap.Markets {
		collateralMap[a.MarketID] = 1
	}

	for _, col := range data.Params.CollateralParams {
		_, found := collateralMap[col.MarketID]
		if !found {
			panic(fmt.Sprintf("%s collateral not found in pricefeed", col.Denom))
		}
	}

	k.SetParams(ctx, data.Params)
	for _, cp := range data.Params.CollateralParams {
		for _, dp := range data.Params.DebtParams {
			k.SetTotalPrincipal(ctx, cp.Denom, dp.Denom, sdk.ZeroInt())
		}
		feeFloat, err := strconv.ParseFloat(cp.StabilityFee.String(), 64)
		if err != nil {
			panic(err)
		}
		feePerSecond := math.Pow(feeFloat, (1 / 31536000.))
		k.SetFeeRate(ctx, cp.Denom, sdk.MustNewDecFromStr(fmt.Sprintf("%.18f", feePerSecond)))
	}

	for _, cdp := range data.CDPs {
		k.SetCDP(ctx, cdp)
		k.IndexCdpByOwner(ctx, cdp)
		ratio := k.CalculateCollateralToDebtRatio(ctx, cdp.Collateral, cdp.Principal.Add(cdp.AccumulatedFees))
		k.IndexCdpByCollateralRatio(ctx, cdp, ratio)
		k.IncrementTotalPrincipal(ctx, cdp.Collateral[0].Denom, cdp.Principal)
	}

	k.SetNextCdpID(ctx, data.StartingCdpID)
	k.SetDebtDenom(ctx, data.DebtDenom)
}

// ExportGenesis export genesis state
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	params := k.GetParams(ctx)
	cdps := k.GetAllCdps(ctx)
	cdpID := k.GetNextCdpID(ctx)

	return GenesisState{
		Params:        params,
		StartingCdpID: cdpID,
		CDPs:          cdps,
	}
}
