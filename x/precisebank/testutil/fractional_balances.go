package testutil

import (
	"math/rand"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/precisebank/types"
	"github.com/stretchr/testify/require"
)

// randRange returns a random number in the range [min, max)
// meaning max is never returned
func randRange(min, max int64) int64 {
	return rand.Int63n(max-min) + min
}

// GenerateEqualFractionalBalances generates count number of FractionalBalances
// with randomly generated amounts such that the sum of all amounts is a
// multiple of types.CONVERSION_FACTOR. If a remainder is desired, any single
// FractionalBalance can be removed from the returned slice and used as the
// remainder.
func GenerateEqualFractionalBalances(
	t *testing.T,
	count int,
) types.FractionalBalances {
	t.Helper()

	fbs := make(types.FractionalBalances, count)
	sum := sdkmath.ZeroInt()

	// Random amounts for count - 1 FractionalBalances
	for i := 0; i < count-1; i++ {
		addr := sdk.AccAddress{byte(i)}.String()

		// Random 0 < amt < CONVERSION_FACTOR
		// POSITIVE and less than CONVERSION_FACTOR
		// If it's 0, Validate() will error
		amt := randRange(1, types.ConversionFactor().Int64())
		amtInt := sdkmath.NewInt(amt)

		fb := types.NewFractionalBalance(addr, amtInt)
		require.NoError(t, fb.Validate())

		fbs[i] = fb

		sum = sum.Add(amtInt)
	}

	// Last FractionalBalance must make sum of all balances equal to have 0
	// fractional remainder. Effectively the amount needed to round up to the
	// nearest integer amount to make this true.
	// (sum + lastAmt) % CONVERSION_FACTOR = 0
	// aka
	// CONVERSION_FACTOR - (sum % CONVERSION_FACTOR) = lastAmt
	addr := sdk.AccAddress{byte(count - 1)}.String()
	amt := types.ConversionFactor().
		Sub(sum.Mod(types.ConversionFactor()))

	fb := types.NewFractionalBalance(addr, amt)
	require.NoError(t, fb.Validate())

	fbs[count-1] = fb

	// Lets double check this before returning
	verificationSum := sdkmath.ZeroInt()
	for _, fb := range fbs {
		verificationSum = verificationSum.Add(fb.Amount)
	}
	require.True(t, verificationSum.Mod(types.ConversionFactor()).IsZero())

	// Also make sure no duplicate addresses
	require.NoError(t, fbs.Validate())

	return fbs
}

// GenerateEqualFractionalBalancesWithRemainder generates count number of
// FractionalBalances with randomly generated amounts as well as a non-zero
// remainder.
// 0 == (sum(FractionalBalances) + remainder) % conversionFactor
// Where remainder > 0
func GenerateEqualFractionalBalancesWithRemainder(
	t *testing.T,
	count int,
) (types.FractionalBalances, sdkmath.Int) {
	t.Helper()

	require.GreaterOrEqual(t, count, 2, "count must be at least 2 to generate both balances and remainder")

	// Generate 1 additional FractionalBalance so we can use one as remainder
	fbs := GenerateEqualFractionalBalances(t, count+1)

	// Use the last one as remainder
	remainder := fbs[count-1].Amount
	fbs = fbs[:count-1]

	require.NotZero(t, remainder.Int64(), "remainder must be non-zero")

	return fbs, remainder
}
