package v0_13

import (
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/supply"

	v0_13cdp "github.com/kava-labs/kava/x/cdp"
	v0_11cdp "github.com/kava-labs/kava/x/cdp/legacy/v0_11"
	v0_13committee "github.com/kava-labs/kava/x/committee"
	v0_11committee "github.com/kava-labs/kava/x/committee/legacy/v0_11"
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

func MigrateCommitte(genesisState v0_11committee.GenesisState) v0_13committee.GenesisState {
	committees := []v0_13committee.Committee{}
	votes := []v0_13committee.Vote{}
	proposals := []v0_13committee.Proposal{}

	var newStabilityCommittee v0_13committee.Committee

	for _, com := range genesisState.Committees {
		if com.ID == 1 {
			newStabilityCommittee.Description = com.Description
			newStabilityCommittee.ID = com.ID
			newStabilityCommittee.Members = com.Members
			newStabilityCommittee.VoteThreshold = com.VoteThreshold
			newStabilityCommittee.ProposalDuration = com.ProposalDuration
			var newStabilityCommitteePermissions []v0_13committee.Permission
			var newStabilitySubParamPermissions v0_13committee.SubParamChangePermission

			for _, perm := range com.Permissions {
				subPerm, ok := perm.(v0_11committee.SubParamChangePermission)
				if ok {
					v0_13committee.AllowedCollateralParams(subPerm.AllowedCollateralParams)
					var newCollateralParams v0_13committee.AllowedCollateralParams
					for _, cp := range subPerm.AllowedCollateralParams {
						newCP := v0_13committee.AllowedCollateralParam(cp)
						newCollateralParams = append(newCollateralParams, newCP)
					}
					newStabilitySubParamPermissions.AllowedCollateralParams = newCollateralParams

					var newAssetParams v0_13committee.AllowedAssetParams
					for _, ap := range subPerm.AllowedAssetParams {
						newAP := v0_13committee.AllowedAssetParam{}
						newAssetParams = append(newAssetParams, newAP)
					}
					newStabilitySubParamPermissions.AllowedAssetParams = newAssetParams

					var newMarketParams v0_13committee.AllowedMarkets
					for _, mp := range subPerm.AllowedMarkets {
						newMP := v0_13committee.AllowedMarket(mp)
						newMarketParams = append(newMarketParams, newMP)
					}
					newStabilitySubParamPermissions.AllowedMarkets = newMarketParams
					newDP := v0_13committee.AllowedDebtParam{
						Denom:            subPerm.AllowedDebtParam.Denom,
						ReferenceAsset:   subPerm.AllowedDebtParam.ReferenceAsset,
						ConversionFactor: subPerm.AllowedDebtParam.ConversionFactor,
						DebtFloor:        subPerm.AllowedDebtParam.DebtFloor,
					}
					newStabilitySubParamPermissions.AllowedDebtParam = newDP

				}

			}

		}
	}

	for _, v := range genesisState.Votes {
		votes = append(votes, v0_13committee.Vote(v))
	}

	for _, p := range genesisState.Proposals {
		proposals = append(proposals, v0_13committee.Proposal(p))
	}
	return v0_13committee.NewGenesisState(
		genesisState.NextProposalID, v0_13committee.DefaultGenesisState().Committees, v0_13committee.DefaultGenesisState().Proposals, v0_13committee.DefaultGenesisState().Votes)
}

func removeIndex(accs authexported.GenesisAccounts, index int) authexported.GenesisAccounts {
	ret := make(authexported.GenesisAccounts, 0)
	ret = append(ret, accs[:index]...)
	return append(ret, accs[index+1:]...)
}
