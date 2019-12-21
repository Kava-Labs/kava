package cdp

import (
	"fmt"

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
		k.SetFeeRate(ctx, cp.Denom, cp.StabilityFee)
	}

	for _, cdp := range data.CDPs {
		k.SetCDP(ctx, cdp)
		k.IndexCdpByOwner(ctx, cdp)
		ratio := k.CalculateCollateralToDebtRatio(ctx, cdp.Collateral, cdp.Principal.Add(cdp.AccumulatedFees))
		k.IndexCdpByCollateralRatio(ctx, cdp.Collateral[0].Denom, cdp.ID, ratio)
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
