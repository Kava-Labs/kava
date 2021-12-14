package v0_16

import (
	v015incentive "github.com/kava-labs/kava/x/incentive/legacy/v0_15"
	v016incentive "github.com/kava-labs/kava/x/incentive/types"
)

func migrateMultiRewardPerids(oldPeriods v015incentive.MultiRewardPeriods) v016incentive.MultiRewardPeriods {
	newPeriods := make(v016incentive.MultiRewardPeriods, len(oldPeriods))
	for i, oldPeriod := range oldPeriods {
		newPeriods[i] = v016incentive.MultiRewardPeriod{
			Active:           oldPeriod.Active,
			CollateralType:   oldPeriod.CollateralType,
			Start:            oldPeriod.Start,
			End:              oldPeriod.End,
			RewardsPerSecond: oldPeriod.RewardsPerSecond,
		}
	}
	return newPeriods
}

func migrateRewardPeriods(oldPeriods v015incentive.RewardPeriods) v016incentive.RewardPeriods {
	newPeriods := make(v016incentive.RewardPeriods, len(oldPeriods))
	for i, oldPeriod := range oldPeriods {
		newPeriods[i] = v016incentive.RewardPeriod{
			Active:           oldPeriod.Active,
			CollateralType:   oldPeriod.CollateralType,
			Start:            oldPeriod.Start,
			End:              oldPeriod.End,
			RewardsPerSecond: oldPeriod.RewardsPerSecond,
		}
	}
	return newPeriods
}

func migrateMultipliersPerDenom(oldMpds v015incentive.MultipliersPerDenom) []v016incentive.MultipliersPerDenom {
	mpds := make([]v016incentive.MultipliersPerDenom, len(oldMpds))
	for i, oldMpd := range oldMpds {
		multipliers := make(v016incentive.Multipliers, len(oldMpd.Multipliers))
		for i, multiplier := range oldMpd.Multipliers {
			multipliers[i] = v016incentive.Multiplier{
				Name:         string(multiplier.Name),
				MonthsLockup: multiplier.MonthsLockup,
				Factor:       multiplier.Factor,
			}
		}
		mpds[i] = v016incentive.MultipliersPerDenom{
			Denom:       oldMpd.Denom,
			Multipliers: multipliers,
		}
	}
	return mpds
}

func migrateParams(params v015incentive.Params) v016incentive.Params {
	return v016incentive.Params{
		USDXMintingRewardPeriods: migrateRewardPeriods(params.USDXMintingRewardPeriods),
		HardSupplyRewardPeriods:  migrateMultiRewardPerids(params.HardSupplyRewardPeriods),
		HardBorrowRewardPeriods:  migrateMultiRewardPerids(params.HardBorrowRewardPeriods),
		DelegatorRewardPeriods:   migrateMultiRewardPerids(params.DelegatorRewardPeriods),
		SwapRewardPeriods:        migrateMultiRewardPerids(params.SwapRewardPeriods),
		ClaimMultipliers:         migrateMultipliersPerDenom(params.ClaimMultipliers),
		ClaimEnd:                 params.ClaimEnd,
	}
}

func migrateRewardState(oldRewardState v015incentive.GenesisRewardState) v016incentive.GenesisRewardState {
	allTimes := make(v016incentive.AccumulationTimes, len(oldRewardState.AccumulationTimes))
	for i, at := range oldRewardState.AccumulationTimes {
		allTimes[i] = v016incentive.AccumulationTime{
			CollateralType:           at.CollateralType,
			PreviousAccumulationTime: at.PreviousAccumulationTime,
		}
	}
	return v016incentive.GenesisRewardState{
		AccumulationTimes:  allTimes,
		MultiRewardIndexes: migrateMultiRewardIndexes(oldRewardState.MultiRewardIndexes),
	}
}

func migrateMultiRewardIndexes(oldMultiRewardIndexes v015incentive.MultiRewardIndexes) v016incentive.MultiRewardIndexes {
	multiRewardIndexes := make(v016incentive.MultiRewardIndexes, len(oldMultiRewardIndexes))
	for i, multiRewardIndex := range oldMultiRewardIndexes {
		multiRewardIndexes[i] = v016incentive.MultiRewardIndex{
			CollateralType: multiRewardIndex.CollateralType,
			RewardIndexes:  migrateRewadIndexes(multiRewardIndex.RewardIndexes),
		}
	}
	return multiRewardIndexes
}

func migrateRewadIndexes(oldRewardIndexes v015incentive.RewardIndexes) v016incentive.RewardIndexes {
	rewardIndexes := make(v016incentive.RewardIndexes, len(oldRewardIndexes))
	for j, rewardIndex := range oldRewardIndexes {
		rewardIndexes[j] = v016incentive.RewardIndex{
			CollateralType: rewardIndex.CollateralType,
			RewardFactor:   rewardIndex.RewardFactor,
		}
	}
	return rewardIndexes
}

func migrateUSDXMintingClaims(oldClaims v015incentive.USDXMintingClaims) v016incentive.USDXMintingClaims {
	claims := make(v016incentive.USDXMintingClaims, len(oldClaims))
	for i, oldClaim := range oldClaims {
		claims[i] = v016incentive.USDXMintingClaim{
			BaseClaim: v016incentive.BaseClaim{
				Owner:  oldClaim.BaseClaim.Owner,
				Reward: oldClaim.BaseClaim.Reward,
			},
			RewardIndexes: migrateRewadIndexes(oldClaim.RewardIndexes),
		}
	}
	return claims
}

func migrateHardLiquidityProviderClaims(oldClaims v015incentive.HardLiquidityProviderClaims) v016incentive.HardLiquidityProviderClaims {
	claims := make(v016incentive.HardLiquidityProviderClaims, len(oldClaims))
	for i, oldClaim := range oldClaims {
		claims[i] = v016incentive.HardLiquidityProviderClaim{
			BaseMultiClaim: v016incentive.BaseMultiClaim{
				Owner:  oldClaim.BaseMultiClaim.Owner,
				Reward: oldClaim.BaseMultiClaim.Reward,
			},
			SupplyRewardIndexes: migrateMultiRewardIndexes(oldClaim.SupplyRewardIndexes),
			BorrowRewardIndexes: migrateMultiRewardIndexes(oldClaim.BorrowRewardIndexes),
		}
	}
	return claims
}

func migrateDelegatorClaims(oldClaims v015incentive.DelegatorClaims) v016incentive.DelegatorClaims {
	claims := make(v016incentive.DelegatorClaims, len(oldClaims))
	for i, oldClaim := range oldClaims {
		claims[i] = v016incentive.DelegatorClaim{
			BaseMultiClaim: v016incentive.BaseMultiClaim{
				Owner:  oldClaim.BaseMultiClaim.Owner,
				Reward: oldClaim.BaseMultiClaim.Reward,
			},
			RewardIndexes: migrateMultiRewardIndexes(oldClaim.RewardIndexes),
		}
	}
	return claims
}

func migrateSwapClaims(oldClaims v015incentive.SwapClaims) v016incentive.SwapClaims {
	claims := make(v016incentive.SwapClaims, len(oldClaims))
	for i, oldClaim := range oldClaims {
		claims[i] = v016incentive.SwapClaim{
			BaseMultiClaim: v016incentive.BaseMultiClaim{
				Owner:  oldClaim.BaseMultiClaim.Owner,
				Reward: oldClaim.BaseMultiClaim.Reward,
			},
			RewardIndexes: migrateMultiRewardIndexes(oldClaim.RewardIndexes),
		}
	}
	return claims
}

// Migrate converts v0.15 incentive state and returns it in v0.16 format
func Migrate(oldState v015incentive.GenesisState) *v016incentive.GenesisState {
	return &v016incentive.GenesisState{
		Params:                      migrateParams(oldState.Params),
		USDXRewardState:             migrateRewardState(oldState.USDXRewardState),
		HardSupplyRewardState:       migrateRewardState(oldState.HardSupplyRewardState),
		HardBorrowRewardState:       migrateRewardState(oldState.HardBorrowRewardState),
		DelegatorRewardState:        migrateRewardState(oldState.DelegatorRewardState),
		SwapRewardState:             migrateRewardState(oldState.SwapRewardState),
		USDXMintingClaims:           migrateUSDXMintingClaims(oldState.USDXMintingClaims),
		HardLiquidityProviderClaims: migrateHardLiquidityProviderClaims(oldState.HardLiquidityProviderClaims),
		DelegatorClaims:             migrateDelegatorClaims(oldState.DelegatorClaims),
		SwapClaims:                  migrateSwapClaims(oldState.SwapClaims),
	}
}
