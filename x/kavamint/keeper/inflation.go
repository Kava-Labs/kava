package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// this is the same value used in the x/hard
const (
	SecondsPerYear = uint64(31536000)
)

// AccumulateStakingRewards calculates the number of coins that should be minted for staking rewards
// given the staking rewards APY and the time of last accumulation.
// The amount is the total_bonded_tokens * spy * seconds_since_last_accumulation
// where spy is the staking_rewards_apy converted to a compound-per-second rate.
func (k Keeper) AccumulateStakingRewards(ctx sdk.Context, since time.Time) (sdk.Coins, error) {
	params := k.GetParams(ctx)
	bondDenom := k.BondDenom(ctx)

	// determine seconds passed since this block time
	// truncate the float with uint64(). remaining fraction of second will be picked up in next block.
	secondsSinceLastBlock := ctx.BlockTime().Sub(since).Seconds()

	// calculate the rate factor based on apy & seconds passed since last block
	stakingRewardsRate, err := CalculateInflationRate(params.StakingRewardsApy, uint64(secondsSinceLastBlock))
	if err != nil {
		return sdk.NewCoins(), err
	}

	totalBonded := k.TotalBondedTokens(ctx)
	stakingRewardsAmount := stakingRewardsRate.MulInt(totalBonded).TruncateInt()

	return sdk.NewCoins(sdk.NewCoin(bondDenom, stakingRewardsAmount)), nil
}

// CalculateInflationRate converts an APY into the factor corresponding with that APY's accumulation
// over a period of secondsPassed seconds.
func CalculateInflationRate(apy sdk.Dec, secondsPassed uint64) (sdk.Dec, error) {
	perSecondInterestRate, err := apyToSpy(apy.Add(sdk.OneDec()))
	if err != nil {
		return sdk.ZeroDec(), err
	}
	rate := perSecondInterestRate.Power(secondsPassed)
	return rate.Sub(sdk.OneDec()), nil
}

// apyToSpy converts the input annual interest rate. For example, 10% apy would be passed as 1.10.
// SPY = Per second compounded interest rate is how cosmos mathematically represents APY.
func apyToSpy(apy sdk.Dec) (sdk.Dec, error) {
	// Note: any APY greater than 176.5 will cause an out-of-bounds error
	root, err := apy.ApproxRoot(SecondsPerYear)
	if err != nil {
		return sdk.ZeroDec(), err
	}
	return root, nil
}
