package keeper

import (
	sdkmath "cosmossdk.io/math"
)

const SecondsPerYear = 365 * 24 * 3600

// CalculateStakingAnnualPercentage returns the annualized staking reward rate.
// It assumes that staking comes from one of two sources depending on if inflation is enabled or not.
func CalculateStakingAnnualPercentage(totalSupply, totalBonded sdkmath.Int, inflationRate, communityTax, rewardsPerSecond sdkmath.LegacyDec) sdkmath.LegacyDec {
	// no rewards are given if no tokens are bonded, in addition avoid division by zero
	if totalBonded.IsZero() {
		return sdkmath.LegacyZeroDec()
	}

	// the percent of inflationRate * totalSupply tokens that are distributed to stakers
	percentInflationDistributedToStakers := sdkmath.LegacyOneDec().Sub(communityTax)

	// the total amount of tokens distributed to stakers in a year
	amountGivenPerYear := inflationRate.
		MulInt(totalSupply).Mul(percentInflationDistributedToStakers).  // portion provided by inflation via mint & distribution modules
		Add(rewardsPerSecond.Mul(sdkmath.LegacyNewDec(SecondsPerYear))) // portion provided by community module

	// divide by total bonded tokens to get the percent return
	return amountGivenPerYear.QuoInt(totalBonded)
}
