package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
)

// NewPeriod returns a new vesting period
func NewPeriod(amount sdk.Coins, length int64) vesting.Period {
	return vesting.Period{Amount: amount, Length: length}
}

// GetTotalVestingPeriodLength returns the summed length of all vesting periods
func GetTotalVestingPeriodLength(periods vesting.Periods) int64 {
	length := int64(0)
	for _, period := range periods {
		length += period.Length
	}
	return length
}

// MultiplyCoins multiplies each value in a set of coins by a single decimal value, rounding the result.
func MultiplyCoins(coins sdk.Coins, multiple sdk.Dec) sdk.Coins {
	var result sdk.Coins
	for _, coin := range coins {
		result = result.Add(
			sdk.NewCoin(coin.Denom, coin.Amount.ToDec().Mul(multiple).RoundInt()),
		)
	}
	return result
}

// FilterCoins returns a subset of the coins by denom. Specifying no denoms will return the original coins.
func FilterCoins(coins sdk.Coins, denoms []string) sdk.Coins {

	if len(denoms) == 0 {
		// with no filter, return all the coins
		return coins
	}
	// otherwise select denoms in filter
	var filteredCoins sdk.Coins
	for _, denom := range denoms {
		filteredCoins = filteredCoins.Add(sdk.NewCoin(denom, coins.AmountOf(denom)))
	}
	return filteredCoins
}
