package testutil

import (
	crand "crypto/rand"
	"math/rand"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/kava-labs/kava/x/precisebank/types"
	"github.com/stretchr/testify/require"
)

// randRange returns a random number in the range [min, max)
// meaning max is never returned
func randRange(min, max int64) int64 {
	return rand.Int63n(max-min) + min
}

func randAccAddress() sdk.AccAddress {
	addrBytes := make([]byte, address.MaxAddrLen)
	_, err := crand.Read(addrBytes)
	if err != nil {
		panic(err)
	}

	addr := sdk.AccAddress(addrBytes)
	if addr.Empty() {
		panic("empty address")
	}

	return addr
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

	// 1 account is not valid, as the total amount needs to be a multiple of
	// conversionFactor. 0 < account balance < conversionFactor, so there must
	// be at least 2
	// NOTE: THIS IS ONLY TRUE WITH 0 REMAINDER
	// GenerateEqualFractionalBalancesWithRemainder repurposes the last balance
	// as the remainder, so this >= 2 requirement is not true in production code.
	require.GreaterOrEqual(t, count, 2, "count must be at least 2 to generate balances")

	fbs := make(types.FractionalBalances, count)
	sum := sdkmath.ZeroInt()

	// Random amounts for count - 1 FractionalBalances
	for i := 0; i < count-1; i++ {
		// Not just using sdk.AccAddress{byte(count)} since that has limited
		// range
		addr := randAccAddress().String()

		// Random 1 < amt < ConversionFactor
		// POSITIVE and less than ConversionFactor
		// If it's 0, Validate() will error.
		// Why start at 2 instead of 1? We want to make sure its divisible
		// for the last account, more details below.
		amt := randRange(2, types.ConversionFactor().Int64())
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
	addr := randAccAddress().String()

	// Why do we need to Mod(conversionFactor) again?
	// Edge case without: If sum == ConversionFactor, then lastAmt == 0 not ConversionFactor
	// 1_000_000_000_000 - (1_000_000_000_000 % 1_000_000_000_000)
	// = 1_000_000_000_000 - 0
	// = 1_000_000_000_000 (invalid!)

	// Note that we only have this issue in tests since we want to calculate a
	// new valid remainder, but we only validate in the actual code.
	amt := types.ConversionFactor().
		Sub(sum.Mod(types.ConversionFactor())).
		Mod(types.ConversionFactor())

	// We only want to generate VALID FractionalBalances - zero would not be
	// valid, so let's just borrow half of the previous amount. We generated
	// amounts from 2 to ConversionFactor, so we know the previous amount is
	// at least 2 and thus able to be split into two valid balances.
	if amt.IsZero() {
		fbs[count-2].Amount = fbs[count-2].Amount.QuoRaw(2)
		amt = fbs[count-2].Amount
	}

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

	countWithRemainder := count + 1

	// Generate 1 additional FractionalBalance so we can use one as remainder
	fbs := GenerateEqualFractionalBalances(t, countWithRemainder)

	// Use the last one as remainder
	remainder := fbs[countWithRemainder-1].Amount

	// Remove the balance used as remainder from the slice
	fbs = fbs[:countWithRemainder-1]

	require.Len(t, fbs, count)
	require.NotZero(t, remainder.Int64(), "remainder must be non-zero")

	return fbs, remainder
}
