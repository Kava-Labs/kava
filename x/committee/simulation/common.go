package simulation

import (
	"fmt"
	"math/big"
	"math/rand"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

func RandomAddresses(r *rand.Rand, accs []simulation.Account) []sdk.AccAddress {
	r.Shuffle(len(accs), func(i, j int) {
		accs[i], accs[j] = accs[j], accs[i]
	})

	var addresses []sdk.AccAddress
	numAddresses := r.Intn(len(accs) + 1)
	for i := 0; i < numAddresses; i++ {
		addresses = append(addresses, accs[i].Address)
	}
	return addresses
}

func RandomPositiveDuration(r *rand.Rand, inclusiveMin, exclusiveMax time.Duration) (time.Duration, error) {
	min := int64(inclusiveMin)
	max := int64(exclusiveMax)
	if min < 0 || max < 0 {
		return 0, fmt.Errorf("min and max must be positive")
	}
	if min >= max {
		return 0, fmt.Errorf("max must be < min")
	}
	randPositiveInt64 := r.Int63n(max-min) + min
	return time.Duration(randPositiveInt64), nil
}

func RandomTime(r *rand.Rand, inclusiveMin, exclusiveMax time.Time) (time.Time, error) {
	if exclusiveMax.Before(inclusiveMin) {
		return time.Time{}, fmt.Errorf("max must be > min")
	}
	period := exclusiveMax.Sub(inclusiveMin)
	subPeriod, err := RandomPositiveDuration(r, 0, period)
	if err != nil {
		return time.Time{}, err
	}
	return inclusiveMin.Add(subPeriod), nil
}

// RandInt randomly generates an sdkmath.Int in the range [inclusiveMin, inclusiveMax]. It works for negative and positive integers.
func RandIntInclusive(r *rand.Rand, inclusiveMin, inclusiveMax sdkmath.Int) (sdkmath.Int, error) {
	if inclusiveMin.GT(inclusiveMax) {
		return sdkmath.Int{}, fmt.Errorf("min larger than max")
	}
	return RandInt(r, inclusiveMin, inclusiveMax.Add(sdk.OneInt()))
}

// RandInt randomly generates an sdkmath.Int in the range [inclusiveMin, exclusiveMax). It works for negative and positive integers.
func RandInt(r *rand.Rand, inclusiveMin, exclusiveMax sdkmath.Int) (sdkmath.Int, error) {
	// validate input
	if inclusiveMin.GTE(exclusiveMax) {
		return sdkmath.Int{}, fmt.Errorf("min larger or equal to max")
	}
	// shift the range to start at 0
	shiftedRange := exclusiveMax.Sub(inclusiveMin) // should always be positive given the check above
	// randomly pick from the shifted range
	shiftedRandInt := sdkmath.NewIntFromBigInt(new(big.Int).Rand(r, shiftedRange.BigInt()))
	// shift back to the original range
	return shiftedRandInt.Add(inclusiveMin), nil
}
