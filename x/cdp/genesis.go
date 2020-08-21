package cdp

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/cdp/types"
)

// InitGenesis sets initial genesis state for cdp module
func InitGenesis(ctx sdk.Context, k Keeper, pk types.PricefeedKeeper, sk types.SupplyKeeper, gs GenesisState) {

	if err := gs.Validate(); err != nil {
		panic(fmt.Sprintf("failed to validate %s genesis state: %s", ModuleName, err))
	}

	// check if the module accounts exists
	cdpModuleAcc := sk.GetModuleAccount(ctx, ModuleName)
	if cdpModuleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", ModuleName))
	}
	liqModuleAcc := sk.GetModuleAccount(ctx, LiquidatorMacc)
	if liqModuleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", LiquidatorMacc))
	}
	savingsRateMacc := sk.GetModuleAccount(ctx, SavingsRateMacc)
	if savingsRateMacc == nil {
		panic(fmt.Sprintf("%s module account has not been set", SavingsRateMacc))
	}

	// validate denoms - check that any collaterals in the params are in the pricefeed,
	// pricefeed MUST call InitGenesis before cdp
	collateralMap := make(map[string]int)
	ap := pk.GetParams(ctx)
	for _, a := range ap.Markets {
		collateralMap[a.MarketID] = 1
	}

	for _, col := range gs.Params.CollateralParams {
		_, found := collateralMap[col.SpotMarketID]
		if !found {
			panic(fmt.Sprintf("%s collateral not found in pricefeed", col.Denom))
		}
		// sets the status of the pricefeed in the store
		// if pricefeed not active, debt operations are paused
		_ = k.UpdatePricefeedStatus(ctx, col.SpotMarketID)

		_, found = collateralMap[col.LiquidationMarketID]
		if !found {
			panic(fmt.Sprintf("%s collateral not found in pricefeed", col.Denom))
		}
		// sets the status of the pricefeed in the store
		// if pricefeed not active, debt operations are paused
		_ = k.UpdatePricefeedStatus(ctx, col.LiquidationMarketID)
	}

	k.SetParams(ctx, gs.Params)

	// set the per second fee rate for each collateral type
	for _, cp := range gs.Params.CollateralParams {
		k.SetTotalPrincipal(ctx, cp.Type, gs.Params.DebtParam.Denom, sdk.ZeroInt())
	}

	// add cdps
	for _, cdp := range gs.CDPs {
		if cdp.ID == gs.StartingCdpID {
			panic(fmt.Sprintf("starting cdp id is assigned to an existing cdp: %s", cdp))
		}
		err := k.SetCDP(ctx, cdp)
		if err != nil {
			panic(fmt.Sprintf("error setting cdp: %v", err))
		}
		k.IndexCdpByOwner(ctx, cdp)
		ratio := k.CalculateCollateralToDebtRatio(ctx, cdp.Collateral, cdp.Type, cdp.GetTotalPrincipal())
		k.IndexCdpByCollateralRatio(ctx, cdp.Type, cdp.ID, ratio)
		k.IncrementTotalPrincipal(ctx, cdp.Type, cdp.GetTotalPrincipal())
	}

	k.SetNextCdpID(ctx, gs.StartingCdpID)
	k.SetDebtDenom(ctx, gs.DebtDenom)
	k.SetGovDenom(ctx, gs.GovDenom)
	// only set the previous block time if it's different than default
	if !gs.PreviousDistributionTime.Equal(types.DefaultPreviousDistributionTime) {
		k.SetPreviousSavingsDistribution(ctx, gs.PreviousDistributionTime)
	}

	for _, d := range gs.Deposits {
		k.SetDeposit(ctx, d)
	}
}

// ExportGenesis export genesis state for cdp module
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	params := k.GetParams(ctx)

	cdps := CDPs{}
	deposits := Deposits{}
	k.IterateAllCdps(ctx, func(cdp CDP) (stop bool) {
		cdps = append(cdps, cdp)
		k.IterateDeposits(ctx, cdp.ID, func(deposit Deposit) (stop bool) {
			deposits = append(deposits, deposit)
			return false
		})
		return false
	})

	cdpID := k.GetNextCdpID(ctx)
	debtDenom := k.GetDebtDenom(ctx)
	govDenom := k.GetGovDenom(ctx)

	previousDistributionTime, found := k.GetPreviousSavingsDistribution(ctx)
	if !found {
		previousDistributionTime = DefaultPreviousDistributionTime
	}

	return NewGenesisState(params, cdps, deposits, cdpID, debtDenom, govDenom, previousDistributionTime)
}
