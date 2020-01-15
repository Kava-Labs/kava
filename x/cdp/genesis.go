package cdp

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis sets initial genesis state for cdp module
func InitGenesis(ctx sdk.Context, k Keeper, pk PricefeedKeeper, gs GenesisState) {
	if err := gs.Validate(); err != nil {
		panic(fmt.Sprintf("failed to validate %s genesis state: %s", ModuleName, err))
	}

	// validate denoms - check that any collaterals in the params are in the pricefeed,
	// pricefeed MUST call InitGenesis before cdp
	collateralMap := make(map[string]int)
	ap := pk.GetParams(ctx)
	for _, a := range ap.Markets {
		collateralMap[a.MarketID] = 1
	}

	for _, col := range gs.Params.CollateralParams {
		_, found := collateralMap[col.MarketID]
		if !found {
			panic(fmt.Sprintf("%s collateral not found in pricefeed", col.Denom))
		}
	}

	k.SetParams(ctx, gs.Params)

	// set the per second fee rate for each collateral type
	for _, cp := range gs.Params.CollateralParams {
		for _, dp := range gs.Params.DebtParams {
			k.SetTotalPrincipal(ctx, cp.Denom, dp.Denom, sdk.ZeroInt())
		}
		k.SetFeeRate(ctx, cp.Denom, cp.StabilityFee)
	}

	// add cdps
	for _, cdp := range gs.CDPs {
		if cdp.ID == gs.StartingCdpID {
			panic(fmt.Sprintf("starting cdp id is assigned to an existing cdp: %s", cdp))
		}
		k.SetCDP(ctx, cdp)
		k.IndexCdpByOwner(ctx, cdp)
		ratio := k.CalculateCollateralToDebtRatio(ctx, cdp.Collateral, cdp.Principal.Add(cdp.AccumulatedFees))
		k.IndexCdpByCollateralRatio(ctx, cdp.Collateral[0].Denom, cdp.ID, ratio)
		k.IncrementTotalPrincipal(ctx, cdp.Collateral[0].Denom, cdp.Principal)
	}

	k.SetNextCdpID(ctx, gs.StartingCdpID)
	k.SetDebtDenom(ctx, gs.DebtDenom)
	k.SetGovDenom(ctx, gs.GovDenom)

	for _, d := range gs.Deposits {
		k.SetDeposit(ctx, d)
	}
	// only set the previous block time if it's different than default
	if !gs.PreviousBlockTime.Equal(DefaultPreviousBlockTime) {
		k.SetPreviousBlockTime(ctx, gs.PreviousBlockTime)
	}
}

// ExportGenesis export genesis state for cdp module
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	params := k.GetParams(ctx)
	cdps := k.GetAllCdps(ctx)
	cdpID := k.GetNextCdpID(ctx)
	previousBlockTime, found := k.GetPreviousBlockTime(ctx)
	if !found {
		previousBlockTime = DefaultPreviousBlockTime
	}
	debtDenom := k.GetDebtDenom(ctx)

	return GenesisState{
		Params:            params,
		StartingCdpID:     cdpID,
		CDPs:              cdps,
		PreviousBlockTime: previousBlockTime,
		DebtDenom:         debtDenom,
	}
}
