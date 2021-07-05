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
	comtypes "github.com/kava-labs/kava/x/committee/types"
	v0_15committee "github.com/kava-labs/kava/x/committee/types"
	v0_15swap "github.com/kava-labs/kava/x/swap/types"
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

	cdc := app.MakeCodec()
	v0_15AppState[v0_15swap.ModuleName] = cdc.MustMarshalJSON(Swap())

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
			stabilityCom := comtypes.NewMemberCommittee(com.ID, com.Description, com.Members,
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
			newStabilityCom := stabilityCom.SetPermissions(newStabilityCommitteePermissions)
			committees = append(committees, newStabilityCom)
		case 2:
			safetyCom := comtypes.NewMemberCommittee(com.ID, com.Description, com.Members,
				[]v0_15committee.Permission{v0_15committee.SoftwareUpgradePermission{}},
				com.VoteThreshold, com.ProposalDuration, v0_15committee.FirstPastThePost)
			committees = append(committees, safetyCom)
		}
	}

	stabilityComMembers, err := loadStabilityComMembers()
	if err != nil {
		panic(err)
	}

	// ---------------------------- Initialize hard governance committee ----------------------------
	hardGovDuration := time.Duration(time.Hour * 24 * 7)
	hardGovThreshold := sdk.MustNewDecFromStr("0.5")
	hardGovQuorum := sdk.MustNewDecFromStr("0.33")

	hardGovCom := comtypes.NewTokenCommittee(3, "Hard Governance Committee", stabilityComMembers,
		[]v0_15committee.Permission{}, hardGovThreshold, hardGovDuration, v0_15committee.Deadline,
		hardGovQuorum, "hard")

	// Add hard money market committee permissions
	var newHardCommitteePermissions []v0_15committee.Permission
	var newHardSubParamPermissions v0_15committee.SubParamChangePermission

	// Allowed params permissions
	hardComAllowedParams := v0_15committee.AllowedParams{
		v0_15committee.AllowedParam{Subspace: "hard", Key: "MoneyMarkets"},
		v0_15committee.AllowedParam{Subspace: "hard", Key: "MinimumBorrowUSDValue"},
		v0_15committee.AllowedParam{Subspace: "incentive", Key: "HardSupplyRewardPeriods"},
		v0_15committee.AllowedParam{Subspace: "incentive", Key: "HardBorrowRewardPeriods"},
		v0_15committee.AllowedParam{Subspace: "incentive", Key: "HardDelegatorRewardPeriods"},
	}
	newHardSubParamPermissions.AllowedParams = hardComAllowedParams

	// Money market permissions
	var newMoneyMarketParams v0_15committee.AllowedMoneyMarkets
	hardMMDenoms := []string{"bnb", "busd", "btcb", "xrpb", "usdx", "ukava", "hard"}
	for _, mmDenom := range hardMMDenoms {
		newMoneyMarketParam := v0_15committee.NewAllowedMoneyMarket(mmDenom, true, true, false, true, true, true)
		newMoneyMarketParams = append(newMoneyMarketParams, newMoneyMarketParam)
	}
	newHardSubParamPermissions.AllowedMoneyMarkets = newMoneyMarketParams
	newHardCommitteePermissions = append(newHardCommitteePermissions, newHardSubParamPermissions)

	// Set hard governance committee permissions
	permissionedHardGovCom := hardGovCom.SetPermissions(newHardCommitteePermissions)
	committees = append(committees, permissionedHardGovCom)

	// ---------------------------- Initialize swp governance committee ----------------------------
	swpGovDuration := time.Duration(time.Hour * 24 * 7)
	swpGovThreshold := sdk.MustNewDecFromStr("0.5")
	swpGovQuorum := sdk.MustNewDecFromStr("0.33")

	swpGovCom := comtypes.NewTokenCommittee(4, "Swp Governance Committee", stabilityComMembers,
		[]v0_15committee.Permission{}, swpGovThreshold, swpGovDuration, v0_15committee.Deadline,
		swpGovQuorum, "swp")

	// Add swap money market committee permissions
	var newSwapCommitteePermissions []v0_15committee.Permission
	var newSwapSubParamPermissions v0_15committee.SubParamChangePermission

	// TODO: add additional incentive params that manage LP rewards
	swpAllowedParams := v0_15committee.AllowedParams{
		v0_15committee.AllowedParam{Subspace: "swap", Key: "AllowedPools"},
		v0_15committee.AllowedParam{Subspace: "swap", Key: "SwapFee"},
		v0_15committee.AllowedParam{Subspace: "incentive", Key: "HardDelegatorRewardPeriods"},
	}
	newSwapSubParamPermissions.AllowedParams = swpAllowedParams

	newSwpCommitteePermissions := append(newSwapCommitteePermissions, newSwapSubParamPermissions)
	permissionedSwapGovCom := swpGovCom.SetPermissions(newSwpCommitteePermissions)
	committees = append(committees, permissionedSwapGovCom)

	for _, v := range genesisState.Votes {
		newVote := v0_15committee.NewVote(v.ProposalID, v.Voter, comtypes.Yes)
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

// Swap introduces new v0.15 swap genesis state
func Swap() v0_15swap.GenesisState {
	return v0_15swap.NewGenesisState(v0_15swap.DefaultParams())
}

func loadStabilityComMembers() ([]sdk.AccAddress, error) {
	strAddrs := []string{
		"kava1gru35up50ql2wxhegr880qy6ynl63ujlv8gum2",
		"kava1sc3mh3pkas5e7xd269am4xm5mp6zweyzmhjagj",
		"kava1c9ye54e3pzwm3e0zpdlel6pnavrj9qqv6e8r4h",
		"kava1m7p6sjqrz6mylz776ct48wj6lpnpcd0z82209d",
		"kava1a9pmkzk570egv3sflu3uwdf3gejl7qfy9hghzl",
	}

	var addrs []sdk.AccAddress
	for _, strAddr := range strAddrs {
		addr, err := sdk.AccAddressFromBech32(strAddr)
		if err != nil {
			return addrs, err
		}
		addrs = append(addrs, addr)
	}

	return addrs, nil
}
