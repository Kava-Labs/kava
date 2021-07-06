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
	v0_15committee "github.com/kava-labs/kava/x/committee/types"
	v0_14incentive "github.com/kava-labs/kava/x/incentive/legacy/v0_14"
	v0_15incentive "github.com/kava-labs/kava/x/incentive/types"
)

var (
	// TODO: update GenesisTime and chain-id for kava-8 launch
	GenesisTime = time.Date(2021, 4, 8, 15, 0, 0, 0, time.UTC)
	ChainID     = "kava-8"
	// TODO: update SWP reward per second amount before production
	SwpRewardsPerSecond = sdk.NewCoin("swp", sdk.OneInt())
)

// Migrate translates a genesis file from kava v0.14 format to kava v0.15 format
func Migrate(genDoc tmtypes.GenesisDoc) tmtypes.GenesisDoc {
	// migrate app state
	var appStateMap genutil.AppMap
	cdc := codec.New()
	cryptoAmino.RegisterAmino(cdc)
	tmtypes.RegisterEvidences(cdc)

	// Old codec does not need all old modules registered on it to correctly decode at this stage
	// as it only decodes the app state into a map of module names to json encoded bytes.
	if err := cdc.UnmarshalJSON(genDoc.AppState, &appStateMap); err != nil {
		panic(err)
	}

	MigrateAppState(appStateMap)

	v0_15Codec := app.MakeCodec()
	marshaledNewAppState, err := v0_15Codec.MarshalJSON(appStateMap)
	if err != nil {
		panic(err)
	}
	genDoc.AppState = marshaledNewAppState
	genDoc.GenesisTime = GenesisTime
	genDoc.ChainID = ChainID
	return genDoc
}

// MigrateAppState migrates application state from v0.14 format to a kava v0.15 format
// It modifies the provided genesis state in place.
func MigrateAppState(v0_14AppState genutil.AppMap) {
	v0_14Codec := makeV014Codec()
	v0_15Codec := app.MakeCodec()

	// Migrate incentive app state
	if v0_14AppState[v0_14incentive.ModuleName] != nil {
		var incentiveGenState v0_14incentive.GenesisState
		v0_14Codec.MustUnmarshalJSON(v0_14AppState[v0_14incentive.ModuleName], &incentiveGenState)
		delete(v0_14AppState, v0_14incentive.ModuleName)
		v0_14AppState[v0_15incentive.ModuleName] = v0_15Codec.MustMarshalJSON(Incentive(incentiveGenState))
	}

	// Migrate commmittee app state
	if v0_14AppState[v0_14committee.ModuleName] != nil {
		// Unmarshal v14 committee genesis state and delete it
		var committeeGS v0_14committee.GenesisState
		v0_14Codec.MustUnmarshalJSON(v0_14AppState[v0_14committee.ModuleName], &committeeGS)
		delete(v0_14AppState, v0_14committee.ModuleName)
		// Marshal v15 committee genesis state
		v0_14AppState[v0_15committee.ModuleName] = v0_15Codec.MustMarshalJSON(Committee(committeeGS))
	}
}

func makeV014Codec() *codec.Codec {
	cdc := codec.New()
	sdk.RegisterCodec(cdc)
	v0_14committee.RegisterCodec(cdc)
	v0_14incentive.RegisterCodec(cdc)
	return cdc
}

// Committee migrates from a v0.14 committee genesis state to a v0.15 committee genesis state
func Committee(genesisState v0_14committee.GenesisState) v0_15committee.GenesisState {

	committees := []v0_15committee.Committee{}
	votes := []v0_15committee.Vote{}
	proposals := []v0_15committee.Proposal{}

	for _, com := range genesisState.Committees {
		if com.ID == 1 {
			// Initialize member committee without permissions
			stabilityCom := v0_15committee.NewMemberCommittee(com.ID, com.Description, com.Members,
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
			safetyCom := v0_15committee.NewMemberCommittee(com.ID, com.Description, com.Members,
				[]v0_15committee.Permission{v0_15committee.SoftwareUpgradePermission{}},
				com.VoteThreshold, com.ProposalDuration, v0_15committee.FirstPastThePost)
			committees = append(committees, safetyCom)
		}
	}

	for _, v := range genesisState.Votes {
		newVote := v0_15committee.NewVote(v.ProposalID, v.Voter, v0_15committee.Yes)
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

// Incentive migrates from a v0.14 incentive genesis state to a v0.15 incentive genesis state
func Incentive(incentiveGS v0_14incentive.GenesisState) v0_15incentive.GenesisState {
	// Migrate params
	var claimMultipliers v0_15incentive.Multipliers
	for _, m := range incentiveGS.Params.ClaimMultipliers {
		newMultiplier := v0_15incentive.NewMultiplier(v0_15incentive.MultiplierName(m.Name), m.MonthsLockup, m.Factor)
		claimMultipliers = append(claimMultipliers, newMultiplier)
	}

	var usdxMintingRewardPeriods v0_15incentive.RewardPeriods
	for _, rp := range incentiveGS.Params.USDXMintingRewardPeriods {
		usdxMintingRewardPeriod := v0_15incentive.NewRewardPeriod(rp.Active,
			rp.CollateralType, rp.Start, rp.End, rp.RewardsPerSecond)
		usdxMintingRewardPeriods = append(usdxMintingRewardPeriods, usdxMintingRewardPeriod)
	}

	var hardSupplyRewardPeriods v0_15incentive.MultiRewardPeriods
	for _, rp := range incentiveGS.Params.HardSupplyRewardPeriods {
		hardSupplyRewardPeriod := v0_15incentive.NewMultiRewardPeriod(rp.Active,
			rp.CollateralType, rp.Start, rp.End, rp.RewardsPerSecond)
		hardSupplyRewardPeriods = append(hardSupplyRewardPeriods, hardSupplyRewardPeriod)
	}

	var hardBorrowRewardPeriods v0_15incentive.MultiRewardPeriods
	for _, rp := range incentiveGS.Params.HardBorrowRewardPeriods {
		hardBorrowRewardPeriod := v0_15incentive.NewMultiRewardPeriod(rp.Active,
			rp.CollateralType, rp.Start, rp.End, rp.RewardsPerSecond)
		hardBorrowRewardPeriods = append(hardBorrowRewardPeriods, hardBorrowRewardPeriod)
	}

	var hardDelegatorRewardPeriods v0_15incentive.MultiRewardPeriods
	for _, rp := range incentiveGS.Params.HardDelegatorRewardPeriods {
		rewardsPerSecond := sdk.NewCoins(rp.RewardsPerSecond, SwpRewardsPerSecond)
		hardDelegatorRewardPeriod := v0_15incentive.NewMultiRewardPeriod(rp.Active,
			rp.CollateralType, rp.Start, rp.End, rewardsPerSecond)
		hardDelegatorRewardPeriods = append(hardDelegatorRewardPeriods, hardDelegatorRewardPeriod)
	}

	// Build new params from migrated values
	params := v0_15incentive.NewParams(
		usdxMintingRewardPeriods,
		hardSupplyRewardPeriods,
		hardBorrowRewardPeriods,
		hardDelegatorRewardPeriods,
		v0_15incentive.DefaultMultiRewardPeriods, // TODO add expected swap reward periods
		claimMultipliers,
		incentiveGS.Params.ClaimEnd,
	)

	// Migrate accumulation times
	var usdxAccumulationTimes v0_15incentive.GenesisAccumulationTimes
	for _, t := range incentiveGS.USDXAccumulationTimes {
		newAccumulationTime := v0_15incentive.NewGenesisAccumulationTime(t.CollateralType, t.PreviousAccumulationTime)
		usdxAccumulationTimes = append(usdxAccumulationTimes, newAccumulationTime)
	}

	var hardSupplyAccumulationTimes v0_15incentive.GenesisAccumulationTimes
	for _, t := range incentiveGS.HardSupplyAccumulationTimes {
		newAccumulationTime := v0_15incentive.NewGenesisAccumulationTime(t.CollateralType, t.PreviousAccumulationTime)
		hardSupplyAccumulationTimes = append(hardSupplyAccumulationTimes, newAccumulationTime)
	}

	var hardBorrowAccumulationTimes v0_15incentive.GenesisAccumulationTimes
	for _, t := range incentiveGS.HardBorrowAccumulationTimes {
		newAccumulationTime := v0_15incentive.NewGenesisAccumulationTime(t.CollateralType, t.PreviousAccumulationTime)
		hardBorrowAccumulationTimes = append(hardBorrowAccumulationTimes, newAccumulationTime)
	}

	var hardDelegatorAccumulationTimes v0_15incentive.GenesisAccumulationTimes
	for _, t := range incentiveGS.HardDelegatorAccumulationTimes {
		newAccumulationTime := v0_15incentive.NewGenesisAccumulationTime(t.CollateralType, t.PreviousAccumulationTime)
		hardDelegatorAccumulationTimes = append(hardDelegatorAccumulationTimes, newAccumulationTime)
	}

	// Migrate USDX minting claims
	var usdxMintingClaims v0_15incentive.USDXMintingClaims
	for _, claim := range incentiveGS.USDXMintingClaims {
		var rewardIndexes v0_15incentive.RewardIndexes
		for _, ri := range claim.RewardIndexes {
			rewardIndex := v0_15incentive.NewRewardIndex(ri.CollateralType, ri.RewardFactor)
			rewardIndexes = append(rewardIndexes, rewardIndex)
		}
		usdxMintingClaim := v0_15incentive.NewUSDXMintingClaim(claim.Owner, claim.Reward, rewardIndexes)
		usdxMintingClaims = append(usdxMintingClaims, usdxMintingClaim)
	}

	// Migrate Hard protocol claims (includes creating new Delegator claims)
	var hardClaims v0_15incentive.HardLiquidityProviderClaims
	var delegatorClaims v0_15incentive.DelegatorClaims
	for _, claim := range incentiveGS.HardLiquidityProviderClaims {
		// Migrate supply multi reward indexes
		var supplyMultiRewardIndexes v0_15incentive.MultiRewardIndexes
		for _, sri := range claim.SupplyRewardIndexes {
			var rewardIndexes v0_15incentive.RewardIndexes
			for _, ri := range sri.RewardIndexes {
				rewardIndex := v0_15incentive.NewRewardIndex(ri.CollateralType, ri.RewardFactor)
				rewardIndexes = append(rewardIndexes, rewardIndex)
			}
			supplyMultiRewardIndex := v0_15incentive.NewMultiRewardIndex(sri.CollateralType, rewardIndexes)
			supplyMultiRewardIndexes = append(supplyMultiRewardIndexes, supplyMultiRewardIndex)
		}

		// Migrate borrow multi reward indexes
		var borrowMultiRewardIndexes v0_15incentive.MultiRewardIndexes
		for _, bri := range claim.BorrowRewardIndexes {
			var rewardIndexes v0_15incentive.RewardIndexes
			for _, ri := range bri.RewardIndexes {
				rewardIndex := v0_15incentive.NewRewardIndex(ri.CollateralType, ri.RewardFactor)
				rewardIndexes = append(rewardIndexes, rewardIndex)
			}
			borrowMultiRewardIndex := v0_15incentive.NewMultiRewardIndex(bri.CollateralType, rewardIndexes)
			borrowMultiRewardIndexes = append(borrowMultiRewardIndexes, borrowMultiRewardIndex)
		}

		// Migrate delegator reward indexes to multi reward indexes inside DelegatorClaims
		var delegatorMultiRewardIndexes v0_15incentive.MultiRewardIndexes
		var delegatorRewardIndexes v0_15incentive.RewardIndexes
		for _, ri := range claim.DelegatorRewardIndexes {
			delegatorRewardIndex := v0_15incentive.NewRewardIndex(ri.CollateralType, ri.RewardFactor)
			delegatorRewardIndexes = append(delegatorRewardIndexes, delegatorRewardIndex)
		}
		delegatorMultiRewardIndex := v0_15incentive.NewMultiRewardIndex(v0_15incentive.BondDenom, delegatorRewardIndexes)
		delegatorMultiRewardIndexes = append(delegatorMultiRewardIndexes, delegatorMultiRewardIndex)

		// TODO: It's impossible to distinguish between rewards from delegation vs. liquidity providing
		//		 as they're all combined inside claim.Reward, so I'm just putting them all inside
		// 		 the hard claim to avoid duplicating rewards.
		delegatorClaim := v0_15incentive.NewDelegatorClaim(claim.Owner, sdk.NewCoins(), delegatorMultiRewardIndexes)
		delegatorClaims = append(delegatorClaims, delegatorClaim)

		hardClaim := v0_15incentive.NewHardLiquidityProviderClaim(claim.Owner, claim.Reward,
			supplyMultiRewardIndexes, borrowMultiRewardIndexes)
		hardClaims = append(hardClaims, hardClaim)
	}

	return v0_15incentive.NewGenesisState(
		params,
		usdxAccumulationTimes,
		hardSupplyAccumulationTimes,
		hardBorrowAccumulationTimes,
		hardDelegatorAccumulationTimes,
		v0_15incentive.DefaultGenesisAccumulationTimes, // There is no previous swap rewards so accumulation starts at genesis time.
		usdxMintingClaims,
		hardClaims,
		delegatorClaims,
		v0_15incentive.DefaultSwapClaims,
	)
}
