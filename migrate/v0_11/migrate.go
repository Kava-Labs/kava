package v0_11

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	v0_11bep3 "github.com/kava-labs/kava/x/bep3/legacy/v0_11"
	v0_9bep3 "github.com/kava-labs/kava/x/bep3/legacy/v0_9"
	v0_11cdp "github.com/kava-labs/kava/x/cdp/legacy/v0_11"
	v0_9cdp "github.com/kava-labs/kava/x/cdp/legacy/v0_9"
	v0_11committee "github.com/kava-labs/kava/x/committee/legacy/v0_11"
	v0_9committee "github.com/kava-labs/kava/x/committee/legacy/v0_9"
	v0_11incentive "github.com/kava-labs/kava/x/incentive/legacy/v0_11"
	v0_9incentive "github.com/kava-labs/kava/x/incentive/legacy/v0_9"
)

// MigrateBep3 migrates from a v0.9 (or v0.10) bep3 genesis state to a v0.11 bep3 genesis state
func MigrateBep3(oldGenState v0_9bep3.GenesisState) v0_11bep3.GenesisState {
	var assetParams v0_11bep3.AssetParams
	var assetSupplies v0_11bep3.AssetSupplies
	v0_9Params := oldGenState.Params

	for _, asset := range v0_9Params.SupportedAssets {
		v10AssetParam := v0_11bep3.AssetParam{
			Active:        asset.Active,
			Denom:         asset.Denom,
			CoinID:        asset.CoinID,
			DeputyAddress: v0_9Params.BnbDeputyAddress,
			FixedFee:      v0_9Params.BnbDeputyFixedFee,
			MinSwapAmount: sdk.OneInt(), // set min swap to one - prevents accounts that hold zero bnb from creating spam txs
			MaxSwapAmount: v0_9Params.MaxAmount,
			MinBlockLock:  v0_9Params.MinBlockLock,
			MaxBlockLock:  v0_9Params.MaxBlockLock,
			SupplyLimit: v0_11bep3.SupplyLimit{
				Limit:          asset.Limit,
				TimeLimited:    false,
				TimePeriod:     time.Duration(0),
				TimeBasedLimit: sdk.ZeroInt(),
			},
		}
		assetParams = append(assetParams, v10AssetParam)
	}
	for _, supply := range oldGenState.AssetSupplies {
		newSupply := v0_11bep3.NewAssetSupply(supply.IncomingSupply, supply.OutgoingSupply, supply.CurrentSupply, sdk.NewCoin(supply.CurrentSupply.Denom, sdk.ZeroInt()), time.Duration(0))
		assetSupplies = append(assetSupplies, newSupply)
	}
	var swaps v0_11bep3.AtomicSwaps
	for _, oldSwap := range oldGenState.AtomicSwaps {
		newSwap := v0_11bep3.AtomicSwap{
			Amount:              oldSwap.Amount,
			RandomNumberHash:    oldSwap.RandomNumberHash,
			ExpireHeight:        oldSwap.ExpireHeight,
			Timestamp:           oldSwap.Timestamp,
			Sender:              oldSwap.Sender,
			Recipient:           oldSwap.Recipient,
			SenderOtherChain:    oldSwap.SenderOtherChain,
			RecipientOtherChain: oldSwap.RecipientOtherChain,
			ClosedBlock:         oldSwap.ClosedBlock,
			Status:              v0_11bep3.SwapStatus(oldSwap.Status),
			CrossChain:          oldSwap.CrossChain,
			Direction:           v0_11bep3.SwapDirection(oldSwap.Direction),
		}
		swaps = append(swaps, newSwap)
	}
	return v0_11bep3.GenesisState{
		Params: v0_11bep3.Params{
			AssetParams: assetParams},
		AtomicSwaps:       swaps,
		Supplies:          assetSupplies,
		PreviousBlockTime: v0_11bep3.DefaultPreviousBlockTime,
	}
}

// MigrateCDP migrates from a v0.9 (or v0.10) cdp genesis state to a v0.11 cdp genesis state
func MigrateCDP(oldGenState v0_9cdp.GenesisState) v0_11cdp.GenesisState {
	var newCDPs v0_11cdp.CDPs
	var newDeposits v0_11cdp.Deposits
	var newCollateralParams v0_11cdp.CollateralParams
	newStartingID := uint64(0)

	for _, cdp := range oldGenState.CDPs {
		newCDP := v0_11cdp.NewCDP(cdp.ID, cdp.Owner, cdp.Collateral, "bnb-a", cdp.Principal, cdp.AccumulatedFees, cdp.FeesUpdated)
		newCDPs = append(newCDPs, newCDP)
		if cdp.ID >= newStartingID {
			newStartingID = cdp.ID + 1
		}
	}

	for _, dep := range oldGenState.Deposits {
		newDep := v0_11cdp.NewDeposit(dep.CdpID, dep.Depositor, dep.Amount)
		newDeposits = append(newDeposits, newDep)
	}

	for _, cp := range oldGenState.Params.CollateralParams {
		newCollateralParam := v0_11cdp.NewCollateralParam(cp.Denom, "bnb-a", cp.LiquidationRatio, cp.DebtLimit, cp.StabilityFee, cp.AuctionSize, cp.LiquidationPenalty, cp.Prefix, cp.SpotMarketID, cp.LiquidationMarketID, cp.ConversionFactor)
		newCollateralParams = append(newCollateralParams, newCollateralParam)
	}
	oldDebtParam := oldGenState.Params.DebtParam

	newDebtParam := v0_11cdp.NewDebtParam(oldDebtParam.Denom, oldDebtParam.ReferenceAsset, oldDebtParam.ConversionFactor, oldDebtParam.DebtFloor, oldDebtParam.SavingsRate)

	newParams := v0_11cdp.NewParams(oldGenState.Params.GlobalDebtLimit, newCollateralParams, newDebtParam, oldGenState.Params.SurplusAuctionThreshold, oldGenState.Params.SurplusAuctionLot, oldGenState.Params.DebtAuctionThreshold, oldGenState.Params.DebtAuctionLot, oldGenState.Params.SavingsDistributionFrequency, false)

	return v0_11cdp.GenesisState{
		Params:                   newParams,
		CDPs:                     newCDPs,
		Deposits:                 newDeposits,
		StartingCdpID:            newStartingID,
		DebtDenom:                oldGenState.DebtDenom,
		GovDenom:                 oldGenState.GovDenom,
		PreviousDistributionTime: oldGenState.PreviousDistributionTime,
	}

}

// MigrateIncentive migrates from a v0.9 (or v0.10) incentive genesis state to a v0.11 incentive genesis state
func MigrateIncentive(oldGenState v0_9incentive.GenesisState) v0_11incentive.GenesisState {
	var newRewards v0_11incentive.Rewards
	var newRewardPeriods v0_11incentive.RewardPeriods
	var newClaimPeriods v0_11incentive.ClaimPeriods
	var newClaims v0_11incentive.Claims
	var newClaimPeriodIds v0_11incentive.GenesisClaimPeriodIDs

	for _, oldReward := range oldGenState.Params.Rewards {
		newReward := v0_11incentive.NewReward(oldReward.Active, oldReward.Denom, oldReward.AvailableRewards, oldReward.Duration, oldReward.TimeLock, oldReward.ClaimDuration)
		newRewards = append(newRewards, newReward)
	}
	newParams := v0_11incentive.NewParams(true, newRewards)

	for _, oldRewardPeriod := range oldGenState.RewardPeriods {
		newRewardPeriod := v0_11incentive.NewRewardPeriod(oldRewardPeriod.Denom, oldRewardPeriod.Start, oldRewardPeriod.End, oldRewardPeriod.Reward, oldRewardPeriod.ClaimEnd, oldRewardPeriod.ClaimTimeLock)
		newRewardPeriods = append(newRewardPeriods, newRewardPeriod)
	}

	for _, oldClaimPeriod := range oldGenState.ClaimPeriods {
		newClaimPeriod := v0_11incentive.NewClaimPeriod(oldClaimPeriod.Denom, oldClaimPeriod.ID, oldClaimPeriod.End, oldClaimPeriod.TimeLock)
		newClaimPeriods = append(newClaimPeriods, newClaimPeriod)
	}

	for _, oldClaim := range oldGenState.Claims {
		newClaim := v0_11incentive.NewClaim(oldClaim.Owner, oldClaim.Reward, oldClaim.Denom, oldClaim.ClaimPeriodID)
		newClaims = append(newClaims, newClaim)
	}

	for _, oldClaimPeriodID := range oldGenState.NextClaimPeriodIDs {
		newClaimPeriodID := v0_11incentive.GenesisClaimPeriodID{
			CollateralType: oldClaimPeriodID.Denom,
			ID:             oldClaimPeriodID.ID,
		}
		newClaimPeriodIds = append(newClaimPeriodIds, newClaimPeriodID)
	}

	return v0_11incentive.NewGenesisState(newParams, oldGenState.PreviousBlockTime, newRewardPeriods, newClaimPeriods, newClaims, newClaimPeriodIds)
}

func MigrateCommittee(oldGenState v0_9committee.GenesisState) v0_11committee.GenesisState {
	var newCommittees []v0_11committee.Committee
	var newStabilityCommittee v0_11committee.Committee
	var newSafetyCommittee v0_11committee.Committee
	var newProposals []v0_11committee.Proposal
	var newVotes []v0_11committee.Vote

	for _, committee := range oldGenState.Committees {
		if committee.ID == 1 {
			newStabilityCommittee.Description = committee.Description
			newStabilityCommittee.ID = committee.ID
			newStabilityCommittee.Members = committee.Members
			newStabilityCommittee.VoteThreshold = committee.VoteThreshold
			newStabilityCommittee.ProposalDuration = committee.ProposalDuration
			var newStabilityPermissions []v0_11committee.Permission
			var newStabilitySubParamPermissions v0_11committee.SubParamChangePermission
			for _, permission := range committee.Permissions {
				subPermission, ok := permission.(v0_9committee.SubParamChangePermission)
				if ok {
					oldCollateralParam := subPermission.AllowedCollateralParams[0]
					newCollateralParam := v0_11committee.AllowedCollateralParam{
						Type:                "bnb-a",
						Denom:               false,
						AuctionSize:         oldCollateralParam.AuctionSize,
						ConversionFactor:    oldCollateralParam.ConversionFactor,
						DebtLimit:           oldCollateralParam.DebtLimit,
						LiquidationMarketID: oldCollateralParam.LiquidationMarketID,
						SpotMarketID:        oldCollateralParam.SpotMarketID,
						LiquidationPenalty:  oldCollateralParam.LiquidationPenalty,
						LiquidationRatio:    oldCollateralParam.LiquidationRatio,
						Prefix:              oldCollateralParam.Prefix,
						StabilityFee:        oldCollateralParam.StabilityFee,
					}
					oldDebtParam := subPermission.AllowedDebtParam
					newDebtParam := v0_11committee.AllowedDebtParam{
						ConversionFactor: oldDebtParam.ConversionFactor,
						DebtFloor:        oldDebtParam.DebtFloor,
						Denom:            oldDebtParam.Denom,
						ReferenceAsset:   oldDebtParam.ReferenceAsset,
						SavingsRate:      oldDebtParam.SavingsRate,
					}
					oldAssetParam := subPermission.AllowedAssetParams[0]
					newAssetParam := v0_11committee.AllowedAssetParam{
						Active: oldAssetParam.Active,
						CoinID: oldAssetParam.CoinID,
						Denom:  oldAssetParam.Denom,
						Limit:  oldAssetParam.Limit,
					}
					oldMarketParams := subPermission.AllowedMarkets
					var newMarketParams v0_11committee.AllowedMarkets
					for _, oldMarketParam := range oldMarketParams {
						newMarketParam := v0_11committee.AllowedMarket(oldMarketParam)
						newMarketParams = append(newMarketParams, newMarketParam)
					}
					oldAllowedParams := subPermission.AllowedParams
					var newAllowedParams v0_11committee.AllowedParams
					for _, oldAllowedParam := range oldAllowedParams {
						newAllowedParam := v0_11committee.AllowedParam(oldAllowedParam)
						if oldAllowedParam.Subspace == "bep3" && oldAllowedParam.Key == "SupportedAssets" {
							newAllowedParam.Key = "AssetParams"
						}

						newAllowedParams = append(newAllowedParams, newAllowedParam)
					}
					newStabilitySubParamPermissions.AllowedAssetParams = v0_11committee.AllowedAssetParams{newAssetParam}
					newStabilitySubParamPermissions.AllowedCollateralParams = v0_11committee.AllowedCollateralParams{newCollateralParam}
					newStabilitySubParamPermissions.AllowedDebtParam = newDebtParam
					newStabilitySubParamPermissions.AllowedMarkets = newMarketParams
					newStabilitySubParamPermissions.AllowedParams = newAllowedParams
					newStabilityPermissions = append(newStabilityPermissions, newStabilitySubParamPermissions)
				}
			}
			newStabilityPermissions = append(newStabilityPermissions, v0_11committee.TextPermission{})
			newStabilityCommittee.Permissions = newStabilityPermissions
			newCommittees = append(newCommittees, newStabilityCommittee)
		} else {
			newSafetyCommittee.ID = committee.ID
			newSafetyCommittee.Description = committee.Description
			newSafetyCommittee.Members = committee.Members
			newSafetyCommittee.Permissions = []v0_11committee.Permission{v0_11committee.SoftwareUpgradePermission{}}
			newSafetyCommittee.VoteThreshold = committee.VoteThreshold
			newSafetyCommittee.ProposalDuration = committee.ProposalDuration
			newCommittees = append(newCommittees, newSafetyCommittee)
		}
	}
	for _, oldProp := range oldGenState.Proposals {
		newPubProposal := v0_11committee.PubProposal(oldProp.PubProposal)
		newProp := v0_11committee.NewProposal(newPubProposal, oldProp.ID, oldProp.CommitteeID, oldProp.Deadline)
		newProposals = append(newProposals, newProp)
	}

	for _, oldVote := range oldGenState.Votes {
		newVote := v0_11committee.NewVote(oldVote.ProposalID, oldVote.Voter)
		newVotes = append(newVotes, newVote)
	}

	return v0_11committee.GenesisState{
		NextProposalID: oldGenState.NextProposalID,
		Committees:     newCommittees,
		Proposals:      newProposals,
		Votes:          newVotes,
	}
}
