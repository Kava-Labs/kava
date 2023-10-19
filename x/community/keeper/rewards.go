package keeper

import (
	sdkmath "cosmossdk.io/math"
)

const SecondsPerYear = 365 * 24 * 3600

// StakingRewardCalculator wraps data and calculation for determining staking rewards
// It assumes that staking comes from one of two sources depending on if inflation is enabled or not.
type StakingRewardCalculator struct {
	TotalSupply   sdkmath.Int
	TotalBonded   sdkmath.Int
	InflationRate sdkmath.LegacyDec
	CommunityTax  sdkmath.LegacyDec

	RewardsPerSecond sdkmath.LegacyDec
}

// GetAnnualizedRate returns the annualized staking reward rate
func (src StakingRewardCalculator) GetAnnualizedRate() sdkmath.LegacyDec {
	inflationEnabledRate := src.GetInflationRewardRate()
	inflationDisabledRate := src.GetRPSRewardRate()
	return inflationEnabledRate.Add(inflationDisabledRate)
}

// GetRPSRewardRate gets the rewards-per-second contribution of the staking reward rate.
// Will be zero if rewards per sec is zero.
func (src StakingRewardCalculator) GetRPSRewardRate() sdkmath.LegacyDec {
	if src.TotalBonded.IsZero() {
		return sdkmath.LegacyZeroDec()
	}
	return src.RewardsPerSecond.Mul(sdkmath.LegacyNewDec(SecondsPerYear)).QuoInt(src.TotalBonded)
}

// GetInflationRewardRate gets the inflationary contribution of the staking reward rate
// Will be zero if inflation is zero.
func (src StakingRewardCalculator) GetInflationRewardRate() sdkmath.LegacyDec {
	if src.TotalBonded.IsZero() {
		return sdkmath.LegacyZeroDec()
	}
	bondedRatio := sdkmath.LegacyNewDecFromInt(src.TotalBonded).QuoInt(src.TotalSupply)
	communityAdjustment := sdkmath.LegacyOneDec().Sub(src.CommunityTax)
	return src.InflationRate.Mul(communityAdjustment).Quo(bondedRatio)
}
