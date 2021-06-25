package v0_15

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"

	cryptoAmino "github.com/tendermint/tendermint/crypto/encoding/amino"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/kava-labs/kava/app"
	v0_14committee "github.com/kava-labs/kava/x/committee/legacy/v0_14"
	"github.com/kava-labs/kava/x/committee/types"
	v0_15committee "github.com/kava-labs/kava/x/committee/types"
)

var (
	// TODO: update GenesisTime for kava-8 launch
	GenesisTime = time.Date(2021, 4, 8, 15, 0, 0, 0, time.UTC)
)

// Migrate translates a genesis file from kava v0.14 format to kava v0.15 format
func Migrate(genDoc tmtypes.GenesisDoc) tmtypes.GenesisDoc {
	// migrate app state
	var appStateMap genutil.AppMap
	cdc := codec.New()
	cryptoAmino.RegisterAmino(cdc)
	tmtypes.RegisterEvidences(cdc)

	if err := cdc.UnmarshalJSON(genDoc.AppState, &appStateMap); err != nil {
		panic(err)
	}
	newAppState := MigrateAppState(appStateMap)
	v0_15Codec := app.MakeCodec()
	marshaledNewAppState, err := v0_15Codec.MarshalJSON(newAppState)
	if err != nil {
		panic(err)
	}
	genDoc.AppState = marshaledNewAppState
	genDoc.GenesisTime = GenesisTime
	genDoc.ChainID = "kava-8"
	return genDoc
}

// MigrateAppState migrates application state from v0.14 format to a kava v0.15 format
func MigrateAppState(v0_14AppState genutil.AppMap) genutil.AppMap {
	v0_15AppState := v0_14AppState

	// Migrate commmittee app state
	if v0_14AppState[v0_14committee.ModuleName] != nil {
		// Unmarshal v14 committee genesis state and delete it
		var committeeGS v0_14committee.GenesisState
		cdc := codec.New()
		sdk.RegisterCodec(cdc)
		v0_14committee.RegisterCodec(cdc)
		cdc.MustUnmarshalJSON(v0_14AppState[v0_14committee.ModuleName], &committeeGS)
		delete(v0_14AppState, v0_14committee.ModuleName)
		// Marshal v15 committee genesis state
		cdc = app.MakeCodec()
		v0_15AppState[v0_15committee.ModuleName] = cdc.MustMarshalJSON(Committee(committeeGS))
	}

	return v0_15AppState
}

// Committee migrates from a v0.14 committee genesis state to a v0.15 committee genesis state
func Committee(genesisState v0_14committee.GenesisState) v0_15committee.GenesisState {

	committees := []v0_15committee.Committee{}
	votes := []v0_15committee.Vote{}
	proposals := []v0_15committee.Proposal{}

	for _, com := range genesisState.Committees {
		switch com.ID {
		case 1:
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
		case 2:
			safetyCom := types.NewMemberCommittee(com.ID, com.Description, com.Members,
				[]v0_15committee.Permission{v0_15committee.SoftwareUpgradePermission{}},
				com.VoteThreshold, com.ProposalDuration, v0_15committee.FirstPastThePost)
			committees = append(committees, safetyCom)
		case 3:
			// Initialize hard governance committee without permissions
			quorum := sdk.MustNewDecFromStr("0.33")
			tallyDenom := "hard"
			hardGovCom := types.NewTokenCommittee(com.ID, com.Description, com.Members,
				[]v0_15committee.Permission{}, com.VoteThreshold, com.ProposalDuration,
				v0_15committee.FirstPastThePost, quorum, tallyDenom)

			// Build hard governance committee permissions
			var newHardCommitteePermissions []v0_15committee.Permission
			var newHardSubParamPermissions v0_15committee.SubParamChangePermission
			for _, perm := range com.Permissions {
				subPerm, ok := perm.(v0_14committee.SubParamChangePermission)
				if ok {
					// Update AllowedParams
					var newAllowedParams v0_15committee.AllowedParams
					for _, ap := range subPerm.AllowedParams {
						newAP := v0_15committee.AllowedParam(ap)
						newAllowedParams = append(newAllowedParams, newAP)
					}
					newHardSubParamPermissions.AllowedParams = newAllowedParams

					// Add hard money market committee permissions
					var newMoneyMarketParams v0_15committee.AllowedMoneyMarkets
					for _, mm := range subPerm.AllowedMoneyMarkets {
						newMoneyMarketParam := v0_15committee.NewAllowedMoneyMarket(
							mm.Denom, mm.BorrowLimit, mm.SpotMarketID, mm.ConversionFactor,
							mm.InterestRateModel, mm.ReserveFactor, mm.KeeperRewardPercentage,
						)
						newMoneyMarketParams = append(newMoneyMarketParams, newMoneyMarketParam)
					}
					newHardSubParamPermissions.AllowedMoneyMarkets = newMoneyMarketParams
					newHardCommitteePermissions = append(newHardCommitteePermissions, newHardSubParamPermissions)
				}
			}
			// Set hard governance committee permissions
			permissionedHardGovCom := hardGovCom.SetPermissions(newHardCommitteePermissions)
			committees = append(committees, permissionedHardGovCom)
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
