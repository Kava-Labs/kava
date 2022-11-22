package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// this is the same value used in the x/hard
const (
	SecondsPerYear = uint64(31536000)
)

// AccumulateInflation calculates the number of coins that should be minted to match a yearly `rate`
// for interest compounded each second of the year over `secondsSinceLastMint` seconds.
// `basis` is the base amount of coins that is inflated.
func (k Keeper) AccumulateInflation(
	ctx sdk.Context,
	rate sdk.Dec,
	basis sdk.Int,
	secondsSinceLastMint float64,
) (sdk.Coins, error) {
	bondDenom := k.BondDenom(ctx)

	// calculate the rate factor based on apy & seconds passed since last block
	inflationRate, err := CalculateInflationRate(rate, uint64(secondsSinceLastMint))
	if err != nil {
		return sdk.NewCoins(), err
	}

	amount := inflationRate.MulInt(basis).TruncateInt()

	return sdk.NewCoins(sdk.NewCoin(bondDenom, amount)), nil
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
