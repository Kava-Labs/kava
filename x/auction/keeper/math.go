package keeper

import (
	"sort"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// splitIntIntoWeightedBuckets divides an initial +ve integer among several buckets in proportion to the buckets' weights
// It uses the largest remainder method: https://en.wikipedia.org/wiki/Largest_remainder_method
// See also: https://stackoverflow.com/questions/13483430/how-to-make-rounded-percentages-add-up-to-100
func splitIntIntoWeightedBuckets(amount sdkmath.Int, buckets []sdkmath.Int) []sdkmath.Int {
	// Limit input to +ve numbers as algorithm hasn't been scoped to work with -ve numbers.
	if amount.IsNegative() {
		panic("negative amount")
	}
	if len(buckets) < 1 {
		panic("no buckets")
	}
	for _, bucket := range buckets {
		if bucket.IsNegative() {
			panic("negative bucket")
		}
	}

	// 1) Split the amount by weights, recording whole number part and remainder

	totalWeights := totalInts(buckets...)
	if !totalWeights.IsPositive() {
		panic("total weights must sum to > 0")
	}

	quotients := make([]quoRem, len(buckets))
	for i := range buckets {
		// amount * ( weight/total_weight )
		q := amount.Mul(buckets[i]).Quo(totalWeights)
		r := amount.Mul(buckets[i]).Mod(totalWeights)
		quotients[i] = quoRem{index: i, quo: q, rem: r}
	}

	// 2) Calculate total left over from remainders, and apportion it to buckets with the highest remainder (to minimize error)

	// sort by decreasing remainder order
	sort.Slice(quotients, func(i, j int) bool {
		return quotients[i].rem.GT(quotients[j].rem)
	})

	// calculate total left over from remainders
	allocated := sdk.ZeroInt()
	for _, qr := range quotients {
		allocated = allocated.Add(qr.quo)
	}
	leftToAllocate := amount.Sub(allocated)

	// apportion according to largest remainder
	results := make([]sdkmath.Int, len(quotients))
	for _, qr := range quotients {
		results[qr.index] = qr.quo
		if !leftToAllocate.IsZero() {
			results[qr.index] = results[qr.index].Add(sdk.OneInt())
			leftToAllocate = leftToAllocate.Sub(sdk.OneInt())
		}
	}
	return results
}

type quoRem struct {
	index int
	quo   sdkmath.Int
	rem   sdkmath.Int
}

// totalInts adds together sdk.Ints
func totalInts(is ...sdkmath.Int) sdkmath.Int {
	total := sdk.ZeroInt()
	for _, i := range is {
		total = total.Add(i)
	}
	return total
}
