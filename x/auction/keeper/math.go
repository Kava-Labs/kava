package keeper

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// splitIntIntoWeightedBuckets divides an initial +ve integer among several buckets in proportion to the buckets' weights
// It uses the largest remainder method:
// https://en.wikipedia.org/wiki/Largest_remainder_method
// see also: https://stackoverflow.com/questions/13483430/how-to-make-rounded-percentages-add-up-to-100
func splitIntIntoWeightedBuckets(amount sdk.Int, buckets []sdk.Int) []sdk.Int {
	// TODO ideally change algorithm to work with -ve numbers. Limiting to +ve numbers until them
	if amount.IsNegative() {
		panic("negative amount")
	}
	for _, bucket := range buckets {
		if bucket.IsNegative() {
			panic("negative bucket")
		}
	}

	totalWeights := totalInts(buckets...)

	// split amount by weights, recording whole number part and remainder
	quotients := make([]quoRem, len(buckets))
	for i := range buckets {
		q := amount.Mul(buckets[i]).Quo(totalWeights)
		r := amount.Mul(buckets[i]).Mod(totalWeights)
		quotients[i] = quoRem{index: i, quo: q, rem: r}
	}

	// apportion left over to buckets with the highest remainder (to minimize error)
	sort.Slice(quotients, func(i, j int) bool {
		return quotients[i].rem.GT(quotients[j].rem) // decreasing remainder order
	})

	allocated := sdk.ZeroInt()
	for _, qr := range quotients {
		allocated = allocated.Add(qr.quo)
	}
	leftToAllocate := amount.Sub(allocated)

	results := make([]sdk.Int, len(quotients))
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
	quo   sdk.Int
	rem   sdk.Int
}

// totalInts adds together sdk.Ints
func totalInts(is ...sdk.Int) sdk.Int {
	total := sdk.ZeroInt()
	for _, i := range is {
		total = total.Add(i)
	}
	return total
}
