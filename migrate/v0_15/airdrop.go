package v0_15

import (
	"fmt"
	"io/ioutil"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/hard"
)

func MakeSwpAirdropMap(hardDepositSnapshotFile string, swpTokens sdk.Int) map[string]sdk.Coin {
	hardDepositSnapshot := getHardDepositsFromFile(hardDepositSnapshotFile)
	var usdxDepositWeights []sdk.Int
	var usdxDepositors []sdk.AccAddress

	for _, dep := range hardDepositSnapshot {
		if dep.Amount.AmountOf("usdx").IsPositive() {
			usdxDepositWeights = append(usdxDepositWeights, dep.Amount.AmountOf("usdx"))
			usdxDepositors = append(usdxDepositors, dep.Depositor)
		}
	}

	tokenBuckets := splitIntIntoWeightedBuckets(swpTokens, usdxDepositWeights)
	if len(tokenBuckets) != len(usdxDepositWeights) {
		panic("swp airdrop accounts not equal to number of USDX depositors")
	}
	if !totalInts(tokenBuckets...).Equal(swpTokens) {
		panic(fmt.Sprintf("expected swp airdrop tokens: %s, %s", swpTokens, totalInts(tokenBuckets...)))
	}

	swpAirdropMap := make(map[string]sdk.Coin)

	for idx, swpTokenAmount := range tokenBuckets {
		swpAirdropMap[usdxDepositors[idx].String()] = sdk.NewCoin("swp", swpTokenAmount)
	}

	return swpAirdropMap
}

func getHardDepositsFromFile(f string) hard.Deposits {
	var deposits hard.Deposits
	cdc := app.MakeCodec()
	bz, err := ioutil.ReadFile(f)
	if err != nil {
		panic(fmt.Sprintf("Couldn't open hard deposit snapshot file: %v", err))
	}
	cdc.MustUnmarshalJSON(bz, &deposits)
	return deposits
}

// splitIntIntoWeightedBuckets divides an initial +ve integer among several buckets in proportion to the buckets' weights
// It uses the largest remainder method: https://en.wikipedia.org/wiki/Largest_remainder_method
// See also: https://stackoverflow.com/questions/13483430/how-to-make-rounded-percentages-add-up-to-100
func splitIntIntoWeightedBuckets(amount sdk.Int, buckets []sdk.Int) []sdk.Int {
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
