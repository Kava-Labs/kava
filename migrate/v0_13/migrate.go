package v0_13

import (
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/supply"

	"github.com/kava-labs/kava/x/bep3"
	v0_13cdp "github.com/kava-labs/kava/x/cdp"
	v0_11cdp "github.com/kava-labs/kava/x/cdp/legacy/v0_11"
)

// MigrateCDP migrates from a v0.11 cdp genesis state to a v0.13 cdp genesis state
func MigrateCDP(oldGenState v0_11cdp.GenesisState) v0_13cdp.GenesisState {
	var newCDPs v0_13cdp.CDPs
	var newDeposits v0_13cdp.Deposits
	var newCollateralParams v0_13cdp.CollateralParams
	var newGenesisAccumulationTimes v0_13cdp.GenesisAccumulationTimes
	var previousAccumulationTime time.Time
	var totalPrincipals v0_13cdp.GenesisTotalPrincipals
	newStartingID := oldGenState.StartingCdpID

	totalPrincipalMap := make(map[string]sdk.Int)

	for _, cdp := range oldGenState.CDPs {
		newCDP := v0_13cdp.NewCDPWithFees(cdp.ID, cdp.Owner, cdp.Collateral, cdp.Type, cdp.Principal, cdp.AccumulatedFees, cdp.FeesUpdated, sdk.OneDec())
		if previousAccumulationTime.Before(cdp.FeesUpdated) {
			previousAccumulationTime = cdp.FeesUpdated
		}
		_, found := totalPrincipalMap[cdp.Type]
		if !found {
			totalPrincipalMap[cdp.Type] = sdk.ZeroInt()
		}
		totalPrincipalMap[cdp.Type] = totalPrincipalMap[cdp.Type].Add(newCDP.GetTotalPrincipal().Amount)

		newCDPs = append(newCDPs, newCDP)
	}

	for _, cp := range oldGenState.Params.CollateralParams {
		newCollateralParam := v0_13cdp.NewCollateralParam(cp.Denom, cp.Type, cp.LiquidationRatio, cp.DebtLimit, cp.StabilityFee, cp.AuctionSize, cp.LiquidationPenalty, cp.Prefix, cp.SpotMarketID, cp.LiquidationMarketID, sdk.MustNewDecFromStr("0.01"), sdk.NewInt(10), cp.ConversionFactor)
		newCollateralParams = append(newCollateralParams, newCollateralParam)
		newGenesisAccumulationTime := v0_13cdp.NewGenesisAccumulationTime(cp.Type, previousAccumulationTime, sdk.OneDec())
		newGenesisAccumulationTimes = append(newGenesisAccumulationTimes, newGenesisAccumulationTime)
	}

	for _, dep := range oldGenState.Deposits {
		newDep := v0_13cdp.NewDeposit(dep.CdpID, dep.Depositor, dep.Amount)
		newDeposits = append(newDeposits, newDep)
	}

	for ctype, tp := range totalPrincipalMap {
		totalPrincipal := v0_13cdp.NewGenesisTotalPrincipal(ctype, tp)
		totalPrincipals = append(totalPrincipals, totalPrincipal)
	}

	sort.Slice(totalPrincipals, func(i, j int) bool { return totalPrincipals[i].CollateralType < totalPrincipals[j].CollateralType })

	oldDebtParam := oldGenState.Params.DebtParam

	newDebtParam := v0_13cdp.NewDebtParam(oldDebtParam.Denom, oldDebtParam.ReferenceAsset, oldDebtParam.ConversionFactor, oldDebtParam.DebtFloor)

	newGlobalDebtLimit := oldGenState.Params.GlobalDebtLimit

	newParams := v0_13cdp.NewParams(newGlobalDebtLimit, newCollateralParams, newDebtParam, oldGenState.Params.SurplusAuctionThreshold, oldGenState.Params.SurplusAuctionLot, oldGenState.Params.DebtAuctionThreshold, oldGenState.Params.DebtAuctionLot, false)

	return v0_13cdp.NewGenesisState(
		newParams,
		newCDPs,
		newDeposits,
		newStartingID,
		oldGenState.DebtDenom,
		oldGenState.GovDenom,
		newGenesisAccumulationTimes,
		totalPrincipals,
	)
}

// MigrateAuth migrates from a v0.11 auth genesis state to a v0.13
func MigrateAuth(genesisState auth.GenesisState) auth.GenesisState {
	savingsRateMaccCoins := sdk.NewCoins()
	savingsMaccAddr := supply.NewModuleAddress(v0_11cdp.SavingsRateMacc)
	savingsRateMaccIndex := 0
	liquidatorMaccIndex := 0
	for idx, acc := range genesisState.Accounts {
		if acc.GetAddress().Equals(savingsMaccAddr) {
			savingsRateMaccCoins = acc.GetCoins()
			savingsRateMaccIndex = idx
			err := acc.SetCoins(acc.GetCoins().Sub(acc.GetCoins()))
			if err != nil {
				panic(err)
			}
		}
		if acc.GetAddress().Equals(supply.NewModuleAddress(v0_13cdp.LiquidatorMacc)) {
			liquidatorMaccIndex = idx
		}
	}
	liquidatorAcc := genesisState.Accounts[liquidatorMaccIndex]
	err := liquidatorAcc.SetCoins(liquidatorAcc.GetCoins().Add(savingsRateMaccCoins...))
	if err != nil {
		panic(err)
	}
	genesisState.Accounts[liquidatorMaccIndex] = liquidatorAcc

	genesisState.Accounts = removeIndex(genesisState.Accounts, savingsRateMaccIndex)
	return genesisState
}

// Bep3 migrates a v0.11 bep3 genesis state to a v0.13 genesis state
func Bep3(genesisState bep3.GenesisState) bep3.GenesisState {
	var newSupplies bep3.AssetSupplies
	for _, supply := range genesisState.Supplies {
		if supply.GetDenom() == "bnb" {
			supply.CurrentSupply = supply.CurrentSupply.Sub(sdk.NewCoin("bnb", sdk.NewInt(1000000000000)))
		}
		newSupplies = append(newSupplies, supply)
	}
	var newSwaps bep3.AtomicSwaps
	for _, swap := range genesisState.AtomicSwaps {
		if swap.Status == bep3.Completed || swap.Status == bep3.Expired {
			swap.ClosedBlock = 1              // reset closed block to one so expired swaps are removed from long term storage properly
			newSwaps = append(newSwaps, swap) // don't migrate open swaps
		}
	}
	return bep3.NewGenesisState(genesisState.Params, newSwaps, newSupplies, genesisState.PreviousBlockTime)
}

func removeIndex(accs authexported.GenesisAccounts, index int) authexported.GenesisAccounts {
	ret := make(authexported.GenesisAccounts, 0)
	ret = append(ret, accs[:index]...)
	return append(ret, accs[index+1:]...)
}
