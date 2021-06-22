package v0_15

import (
	v0_14committee "github.com/kava-labs/kava/x/committee/legacy/v0_14"
	"github.com/kava-labs/kava/x/committee/types"
	v0_15committee "github.com/kava-labs/kava/x/committee/types"
)

// Committee migrates from a v0.14 committee genesis state to a v0.15 committee genesis state
func Committee(genesisState v0_14committee.GenesisState) v0_15committee.GenesisState {

	committees := []v0_15committee.Committee{}
	votes := []v0_15committee.Vote{}
	proposals := []v0_15committee.Proposal{}

	for _, com := range genesisState.Committees {
		if com.ID == 1 {
			// Initialize member committee without permissions
			stabilityCom := types.NewMemberCommittee(com.ID, com.Description, com.Members,
				[]v0_15committee.Permission{}, com.VoteThreshold, com.ProposalDuration,
				v0_15committee.FirstPastThePost)

			// Build stability committee permissions
			var newStabilityCommitteePermissions []v0_15committee.Permission
			var newStabilitySubParamPermissions v0_15committee.SubParamChangePermission
			for _, perm := range com.Permissions {
				subPerm, ok := perm.(v0_14committee.SubParamChangePermission)
				if ok {
					// update AllowedParams
					var newAllowedParams v0_15committee.AllowedParams
					for _, ap := range subPerm.AllowedParams {
						newAP := v0_15committee.AllowedParam(ap)
						newAllowedParams = append(newAllowedParams, newAP)
					}
					newStabilitySubParamPermissions.AllowedParams = newAllowedParams

					// update AllowedCollateralParams
					var newCollateralParams v0_15committee.AllowedCollateralParams
					collateralTypes := []string{"bnb-a", "busd-a", "busd-b", "btcb-a", "xrpb-a", "ukava-a", "hard-a", "hbtc-a"}
					for _, cp := range subPerm.AllowedCollateralParams {
						newCP := v0_15committee.NewAllowedCollateralParam(
							cp.Type,
							cp.Denom,
							cp.LiquidationRatio,
							cp.DebtLimit,
							cp.StabilityFee,
							cp.AuctionSize,
							cp.LiquidationPenalty,
							cp.Prefix,
							cp.SpotMarketID,
							cp.LiquidationMarketID,
							cp.ConversionFactor,
							true,
							true,
						)
						newCollateralParams = append(newCollateralParams, newCP)
					}
					for _, cType := range collateralTypes {
						var foundCtype bool
						for _, cp := range newCollateralParams {
							if cType == cp.Type {
								foundCtype = true
							}
						}
						if !foundCtype {
							newCP := v0_15committee.NewAllowedCollateralParam(cType, false, false, true, true, true, false, false, false, false, false, true, true)
							newCollateralParams = append(newCollateralParams, newCP)
						}
					}
					newStabilitySubParamPermissions.AllowedCollateralParams = newCollateralParams

					// update AllowedDebtParam
					newDP := v0_15committee.AllowedDebtParam{
						Denom:            subPerm.AllowedDebtParam.Denom,
						ReferenceAsset:   subPerm.AllowedDebtParam.ReferenceAsset,
						ConversionFactor: subPerm.AllowedDebtParam.ConversionFactor,
						DebtFloor:        subPerm.AllowedDebtParam.DebtFloor,
					}
					newStabilitySubParamPermissions.AllowedDebtParam = newDP

					// update AllowedAssetParams
					var newAssetParams v0_15committee.AllowedAssetParams
					for _, ap := range subPerm.AllowedAssetParams {
						newAP := v0_15committee.AllowedAssetParam(ap)
						newAssetParams = append(newAssetParams, newAP)
					}
					newStabilitySubParamPermissions.AllowedAssetParams = newAssetParams

					// Update Allowed Markets
					var newMarketParams v0_15committee.AllowedMarkets
					for _, mp := range subPerm.AllowedMarkets {
						newMP := v0_15committee.AllowedMarket(mp)
						newMarketParams = append(newMarketParams, newMP)
					}
					newStabilitySubParamPermissions.AllowedMarkets = newMarketParams

					// Add hard money market committee permissions
					var newMoneyMarketParams v0_15committee.AllowedMoneyMarkets
					hardMMDenoms := []string{"bnb", "busd", "btcb", "xrpb", "usdx", "ukava", "hard"}
					for _, mmDenom := range hardMMDenoms {
						newMoneyMarketParam := v0_15committee.NewAllowedMoneyMarket(mmDenom, true, false, false, true, true, true)
						newMoneyMarketParams = append(newMoneyMarketParams, newMoneyMarketParam)
					}
					newStabilitySubParamPermissions.AllowedMoneyMarkets = newMoneyMarketParams
					newStabilityCommitteePermissions = append(newStabilityCommitteePermissions, newStabilitySubParamPermissions)
				}
			}
			newStabilityCommitteePermissions = append(newStabilityCommitteePermissions, v0_15committee.TextPermission{})

			// Set stability committee permissions
			baseStabilityCom := stabilityCom.SetPermissions(newStabilityCommitteePermissions)
			newStabilityCom := v0_15committee.MemberCommittee{BaseCommittee: baseStabilityCom}
			committees = append(committees, newStabilityCom)
		} else {
			safetyCom := types.NewMemberCommittee(com.ID, com.Description, com.Members,
				[]v0_15committee.Permission{v0_15committee.SoftwareUpgradePermission{}},
				com.VoteThreshold, com.ProposalDuration, v0_15committee.FirstPastThePost)
			committees = append(committees, safetyCom)
		}
	}

	for _, v := range genesisState.Votes {
		newVote := v0_15committee.NewVote(v.ProposalID, v.Voter, types.Yes)
		votes = append(votes, v0_15committee.Vote(newVote))
	}

	for _, p := range genesisState.Proposals {
		newPubProp := v0_15committee.PubProposal(p.PubProposal)
		newProp := v0_15committee.NewProposal(newPubProp, p.ID, p.CommitteeID, p.Deadline)
		proposals = append(proposals, newProp)
	}
	return v0_15committee.NewGenesisState(
		genesisState.NextProposalID, committees, proposals, votes)
}
