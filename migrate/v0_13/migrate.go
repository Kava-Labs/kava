package v0_13

import (
	"time"

	v0_11cdp "github.com/kava-labs/kava/x/cdp/legacy/v0_11"
	v0_13cdp "github.com/kava-labs/kava/x/cdp/legacy/v0_13"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MigrateCDP migrates from a v0.9 (or v0.10) cdp genesis state to a v0.11 cdp genesis state
func MigrateCDP(oldGenState v0_11cdp.GenesisState) v0_13cdp.GenesisState {
	var newCDPs v0_13cdp.CDPs
	var newDeposits v0_13cdp.Deposits
	var newCollateralParams v0_13cdp.CollateralParams
	var newGenesisAccumulationTimes v0_13cdp.GenesisAccumulationTimes
	var previousAccumulationTime time.Time
	var totalPrincipals v0_13cdp.GenesisTotalPrincipals
	newStartingID := oldGenState.StartingCdpID

	totalPrincipalMap := make(map[string]sdk.Int)
	for _, cp := range oldGenState.Params.CollateralParams {
		newCollateralParam := v0_13cdp.NewCollateralParam(cp.Denom, cp.Type, cp.LiquidationRatio, cp.DebtLimit, cp.StabilityFee, cp.AuctionSize, cp.LiquidationPenalty, cp.Prefix, cp.SpotMarketID, cp.LiquidationMarketID, cp.ConversionFactor)
		newCollateralParams = append(newCollateralParams, newCollateralParam)
		newGenesisAccumulationTime := v0_13cdp.NewGenesisAccumulationTime(cp.Type, previousAccumulationTime, sdk.OneDec())
		newGenesisAccumulationTimes = append(newGenesisAccumulationTimes, newGenesisAccumulationTime)
		totalPrincipalMap[cp.Type] = sdk.ZeroInt()
	}

	for _, cdp := range oldGenState.CDPs {
		newCDP := v0_13cdp.NewCDPWithFees(cdp.ID, cdp.Owner, cdp.Collateral, cdp.Type, cdp.Principal, cdp.AccumulatedFees, cdp.FeesUpdated, sdk.OneDec())
		if previousAccumulationTime.Before(cdp.FeesUpdated) {
			previousAccumulationTime = cdp.FeesUpdated
		}
		totalPrincipalMap[cdp.Type] = totalPrincipalMap[cdp.Type].Add(newCDP.GetTotalPrincipal().Amount)

		newCDPs = append(newCDPs, newCDP)
	}

	for _, dep := range oldGenState.Deposits {
		newDep := v0_13cdp.NewDeposit(dep.CdpID, dep.Depositor, dep.Amount)
		newDeposits = append(newDeposits, newDep)
	}

	for ctype, tp := range totalPrincipalMap {
		totalPrincipal := v0_13cdp.NewGenesisTotalPrincipal(ctype, tp)
		totalPrincipals = append(totalPrincipals, totalPrincipal)
	}

	oldDebtParam := oldGenState.Params.DebtParam

	newDebtParam := v0_13cdp.NewDebtParam(oldDebtParam.Denom, oldDebtParam.ReferenceAsset, oldDebtParam.ConversionFactor, oldDebtParam.DebtFloor, oldDebtParam.SavingsRate)

	newGlobalDebtLimit := oldGenState.Params.GlobalDebtLimit

	newParams := v0_13cdp.NewParams(newGlobalDebtLimit, newCollateralParams, newDebtParam, oldGenState.Params.SurplusAuctionThreshold, oldGenState.Params.SurplusAuctionLot, oldGenState.Params.DebtAuctionThreshold, oldGenState.Params.DebtAuctionLot, oldGenState.Params.SavingsDistributionFrequency, false)

	return v0_13cdp.NewGenesisState(
		newParams,
		newCDPs,
		newDeposits,
		newStartingID,
		oldGenState.DebtDenom,
		oldGenState.GovDenom,
		oldGenState.PreviousDistributionTime,
		sdk.ZeroInt(),
		newGenesisAccumulationTimes,
		totalPrincipals,
	)
}
