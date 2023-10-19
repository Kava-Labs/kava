package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const SecondsPerYear = 365 * 24 * 3600

// StakingRewardCalculator wraps data and calculation for determining staking rewards
// It assumes that staking comes from one of two sources depending on if inflation is enabled or not.
type StakingRewardCalculator struct {
	TotalSupply   sdkmath.Int
	TotalBonded   sdkmath.Int
	InflationRate sdk.Dec
	CommunityTax  sdk.Dec

	RewardsPerSecond sdk.Dec
}

// GetAnnualizedRate returns the annualized staking reward rate
func (src StakingRewardCalculator) GetAnnualizedRate() sdk.Dec {
	inflationEnabledRate := src.GetInflationRewardRate()
	inflationDisabledRate := src.GetRPSRewardRate()
	return inflationEnabledRate.Add(inflationDisabledRate)
}

// GetRPSRewardRate gets the rewards-per-second contribution of the staking reward rate.
// Will be zero if rewards per sec is zero.
func (src StakingRewardCalculator) GetRPSRewardRate() sdk.Dec {
	if src.TotalBonded.IsZero() {
		return sdk.ZeroDec()
	}
	return src.RewardsPerSecond.Mul(sdk.NewDec(SecondsPerYear)).QuoInt(src.TotalBonded)
}

// GetInflationRewardRate gets the inflationary contribution of the staking reward rate
// Will be zero if inflation is zero.
func (src StakingRewardCalculator) GetInflationRewardRate() sdk.Dec {
	if src.TotalBonded.IsZero() {
		return sdk.ZeroDec()
	}
	bondedRatio := sdk.NewDecFromInt(src.TotalBonded).QuoInt(src.TotalSupply)
	communityAdjustment := sdk.OneDec().Sub(src.CommunityTax)
	return src.InflationRate.Mul(communityAdjustment).Quo(bondedRatio)
}
