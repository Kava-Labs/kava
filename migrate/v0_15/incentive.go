package v0_15

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	v0_14incentive "github.com/kava-labs/kava/x/incentive/legacy/v0_14"
	v0_15incentive "github.com/kava-labs/kava/x/incentive/types"
)

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
