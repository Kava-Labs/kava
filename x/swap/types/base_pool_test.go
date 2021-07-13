package types_test

import (
	"fmt"
	"math/big"
	"testing"

	types "github.com/kava-labs/kava/x/swap/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// i creates a new sdk.Int from int64
func i(n int64) sdk.Int {
	return sdk.NewInt(n)
}

// s returns a new sdk.Int from a string
func s(str string) sdk.Int {
	num, ok := sdk.NewIntFromString(str)
	if !ok {
		panic(fmt.Sprintf("overflow creating Int from %s", str))
	}
	return num
}

// d creates a new sdk.Dec from a string
func d(str string) sdk.Dec {
	return sdk.MustNewDecFromStr(str)
}

// exp takes a sdk.Int and computes the power
// helper to generate large numbers
func exp(n sdk.Int, power int64) sdk.Int {
	b := n.BigInt()
	b.Exp(b, big.NewInt(power), nil)
	return sdk.NewIntFromBigInt(b)
}

func TestBasePool_NewPool_Validation(t *testing.T) {
	testCases := []struct {
		reservesA   sdk.Int
		reservesB   sdk.Int
		expectedErr string
	}{
		{i(0), i(1e6), "invalid pool: reserves must be greater than zero"},
		{i(0), i(0), "invalid pool: reserves must be greater than zero"},
		{i(-1), i(1e6), "invalid pool: reserves must be greater than zero"},
		{i(1e6), i(-1), "invalid pool: reserves must be greater than zero"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("reservesA=%s reservesB=%s", tc.reservesA, tc.reservesB), func(t *testing.T) {
			pool, err := types.NewBasePool(tc.reservesA, tc.reservesB)
			require.EqualError(t, err, tc.expectedErr)
			assert.Nil(t, pool)
		})
	}
}

func TestBasePool_NewPoolWithExistingShares_Validation(t *testing.T) {
	testCases := []struct {
		reservesA   sdk.Int
		reservesB   sdk.Int
		totalShares sdk.Int
		expectedErr string
	}{
		{i(0), i(1e6), i(1), "invalid pool: reserves must be greater than zero"},
		{i(0), i(0), i(1), "invalid pool: reserves must be greater than zero"},
		{i(-1), i(1e6), i(3), "invalid pool: reserves must be greater than zero"},
		{i(1e6), i(-1), i(100), "invalid pool: reserves must be greater than zero"},
		{i(1e6), i(-1), i(3), "invalid pool: reserves must be greater than zero"},
		{i(1e6), i(1e6), i(0), "invalid pool: total shares must be greater than zero"},
		{i(1e6), i(1e6), i(-1), "invalid pool: total shares must be greater than zero"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("reservesA=%s reservesB=%s shares=%s", tc.reservesA, tc.reservesB, tc.totalShares), func(t *testing.T) {
			pool, err := types.NewBasePoolWithExistingShares(tc.reservesA, tc.reservesB, tc.totalShares)
			require.EqualError(t, err, tc.expectedErr)
			assert.Nil(t, pool)
		})
	}
}

func TestBasePool_InitialState(t *testing.T) {
	testCases := []struct {
		reservesA      sdk.Int
		reservesB      sdk.Int
		expectedShares sdk.Int
	}{
		{i(1), i(1), i(1)},
		{i(100), i(100), i(100)},
		{i(100), i(10000000), i(31622)},
		{i(1e5), i(5e6), i(707106)},
		{i(1e6), i(5e6), i(2236067)},
		{i(1e15), i(7e15), i(2645751311064590)},
		{i(1), i(6e18), i(2449489742)},
		{i(1.345678e18), i(4.313456e18), i(2409257736973775913)},
		// handle sqrt of large numbers, sdk.Int.ApproxSqrt() doesn't converge in 100 iterations
		{i(145345664).Mul(exp(i(10), 26)), i(6432294561).Mul(exp(i(10), 20)), s("96690543695447979624812468142651")},
		{i(465432423).Mul(exp(i(10), 50)), i(4565432).Mul(exp(i(10), 50)), s("4609663846531258725944608083913166083991595286362304230475")},
		{exp(i(2), 253), exp(i(2), 253), s("14474011154664524427946373126085988481658748083205070504932198000989141204992")},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("reservesA=%s reservesB=%s", tc.reservesA, tc.reservesB), func(t *testing.T) {
			pool, err := types.NewBasePool(tc.reservesA, tc.reservesB)
			require.Nil(t, err)
			assert.Equal(t, tc.reservesA, pool.ReservesA())
			assert.Equal(t, tc.reservesB, pool.ReservesB())
			assert.Equal(t, tc.expectedShares, pool.TotalShares())
		})
	}
}

func TestBasePool_ExistingState(t *testing.T) {
	testCases := []struct {
		reservesA   sdk.Int
		reservesB   sdk.Int
		totalShares sdk.Int
	}{
		{i(1), i(1), i(1)},
		{i(100), i(100), i(100)},
		{i(1e5), i(5e6), i(707106)},
		{i(1e15), i(7e15), i(2645751311064590)},
		{i(465432423).Mul(exp(i(10), 50)), i(4565432).Mul(exp(i(10), 50)), s("4609663846531258725944608083913166083991595286362304230475")},
		{exp(i(2), 253), exp(i(2), 253), s("14474011154664524427946373126085988481658748083205070504932198000989141204992")},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("reservesA=%s reservesB=%s shares=%s", tc.reservesA, tc.reservesB, tc.totalShares), func(t *testing.T) {
			pool, err := types.NewBasePoolWithExistingShares(tc.reservesA, tc.reservesB, tc.totalShares)
			require.Nil(t, err)
			assert.Equal(t, tc.reservesA, pool.ReservesA())
			assert.Equal(t, tc.reservesB, pool.ReservesB())
			assert.Equal(t, tc.totalShares, pool.TotalShares())
		})
	}
}

func TestBasePool_ShareValue_PoolCreator(t *testing.T) {
	testCases := []struct {
		reservesA sdk.Int
		reservesB sdk.Int
	}{
		{i(1), i(1)},
		{i(100), i(100)},
		{i(100), i(10000000)},
		{i(1e5), i(5e6)},
		{i(1e15), i(7e15)},
		{i(1), i(6e18)},
		{i(1.345678e18), i(4.313456e18)},
		// ensure no overflows in intermediate values
		{exp(i(2), 253), exp(i(2), 253)},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("reservesA=%s reservesB=%s", tc.reservesA, tc.reservesB), func(t *testing.T) {
			pool, err := types.NewBasePool(tc.reservesA, tc.reservesB)
			assert.NoError(t, err)

			a, b := pool.ShareValue(pool.TotalShares())
			// pool creators experience zero truncation error and always
			// and always receive their original balance on a 100% withdraw
			// when there are no other deposits that result in a fractional share ownership
			assert.Equal(t, tc.reservesA, a, "share value of reserves A not equal")
			assert.Equal(t, tc.reservesB, b, "share value of reserves B not equal")
		})
	}
}

func TestBasePool_AddLiquidity(t *testing.T) {
	testCases := []struct {
		initialA       sdk.Int
		initialB       sdk.Int
		desiredA       sdk.Int
		desiredB       sdk.Int
		expectedA      sdk.Int
		expectedB      sdk.Int
		expectedShares sdk.Int
	}{
		{i(1), i(1), i(1), i(1), i(1), i(1), i(1)},   // small pool, i(100)% deposit
		{i(10), i(10), i(5), i(5), i(5), i(5), i(5)}, // i(50)% deposit
		{i(10), i(10), i(3), i(3), i(3), i(3), i(3)}, // i(30)% deposit
		{i(10), i(10), i(1), i(1), i(1), i(1), i(1)}, // i(10)% deposit

		// small pools, unequal deposit ratios
		{i(11), i(10), i(5), i(6), i(5), i(4), i(4)},
		{i(11), i(10), i(5), i(5), i(5), i(4), i(4)},
		// this test case fails if we don't use min share ratio
		{i(11), i(10), i(5), i(4), i(4), i(4), i(3)},

		// small pools, unequal deposit ratios, reversed
		{i(10), i(11), i(6), i(5), i(4), i(5), i(4)},
		{i(10), i(11), i(5), i(5), i(4), i(5), i(4)},
		// this test case fails if we don't use min share ratio
		{i(10), i(11), i(4), i(5), i(4), i(4), i(3)},

		{i(10e6), i(11e6), i(5e6), i(5e6), i(4545454), i(5e6), i(4767312)},
		{i(11e6), i(10e6), i(5e6), i(5e6), i(5e6), i(4545454), i(4767312)},

		// pool size near max of sdk.Int, ensure intermidiate calculations do not overflow
		{exp(i(10), 70), exp(i(10), 70), i(1e18), i(1e18), i(1e18), i(1e18), i(1e18)},
	}

	for _, tc := range testCases {
		name := fmt.Sprintf("initialA=%s initialB=%s desiredA=%s desiredB=%s", tc.initialA, tc.initialB, tc.desiredA, tc.desiredB)
		t.Run(name, func(t *testing.T) {
			pool, err := types.NewBasePool(tc.initialA, tc.initialB)
			require.NoError(t, err)
			initialShares := pool.TotalShares()

			actualA, actualB, actualShares := pool.AddLiquidity(tc.desiredA, tc.desiredB)

			// assert correct values are retruned
			assert.Equal(t, tc.expectedA, actualA, "deposited A liquidity not equal")
			assert.Equal(t, tc.expectedB, actualB, "deposited B liquidity not equal")
			assert.Equal(t, tc.expectedShares, actualShares, "calculated shares not equal")

			// assert pool liquidity and shares are updated
			assert.Equal(t, tc.initialA.Add(actualA), pool.ReservesA(), "total reserves A not equal")
			assert.Equal(t, tc.initialB.Add(actualB), pool.ReservesB(), "total reserves B not equal")
			assert.Equal(t, initialShares.Add(actualShares), pool.TotalShares(), "total shares not equal")

			leftA := actualShares.BigInt()
			leftA.Mul(leftA, tc.initialA.BigInt())
			rightA := initialShares.BigInt()
			rightA.Mul(rightA, actualA.BigInt())

			leftB := actualShares.BigInt()
			leftB.Mul(leftB, tc.initialB.BigInt())
			rightB := initialShares.BigInt()
			rightB.Mul(rightB, actualB.BigInt())

			// assert that the share ratio is less than or equal to the deposit ratio
			// actualShares / initialShares <= actualA / initialA
			assert.True(t, leftA.Cmp(rightA) <= 0, "share ratio is greater than deposit A ratio")
			// actualShares / initialShares <= actualB / initialB
			assert.True(t, leftB.Cmp(rightB) <= 0, "share ratio is greater than deposit B ratio")

			// assert that share value of returned shares is not greater than the deposited amount
			shareValueA, shareValueB := pool.ShareValue(actualShares)
			assert.True(t, shareValueA.LTE(actualA), "share value A greater than deposited A")
			assert.True(t, shareValueB.LTE(actualB), "share value B greater than deposited B")
		})
	}
}

func TestBasePool_RemoveLiquidity(t *testing.T) {
	testCases := []struct {
		reservesA sdk.Int
		reservesB sdk.Int
		shares    sdk.Int
		expectedA sdk.Int
		expectedB sdk.Int
	}{
		{i(1), i(1), i(1), i(1), i(1)},
		{i(100), i(100), i(50), i(50), i(50)},
		{i(100), i(10000000), i(10435), i(32), i(3299917)},
		{i(10000000), i(100), i(10435), i(3299917), i(32)},
		{i(1.345678e18), i(4.313456e18), i(3.134541e17), i(175078108044025869), i(561197935621412888)},
		// ensure no overflows in intermediate values
		{exp(i(10), 70), exp(i(10), 70), i(1e18), i(1e18), i(1e18)},
		{exp(i(2), 253), exp(i(2), 253), exp(i(2), 253), exp(i(2), 253), exp(i(2), 253)},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("reservesA=%s reservesB=%s shares=%s", tc.reservesA, tc.reservesB, tc.shares), func(t *testing.T) {
			pool, err := types.NewBasePool(tc.reservesA, tc.reservesB)
			assert.NoError(t, err)
			initialShares := pool.TotalShares()

			a, b := pool.RemoveLiquidity(tc.shares)

			// pool creators experience zero truncation error and always
			// and always receive their original balance on a 100% withdraw
			// when there are no other deposits that result in a fractional share ownership
			assert.Equal(t, tc.expectedA, a, "withdrawn A not equal")
			assert.Equal(t, tc.expectedB, b, "withdrawn B not equal")

			// asset that pool state is updated
			assert.Equal(t, tc.reservesA.Sub(a), pool.ReservesA(), "reserves A after withdraw not equal")
			assert.Equal(t, tc.reservesB.Sub(b), pool.ReservesB(), "reserves B after withdraw not equal")
			assert.Equal(t, initialShares.Sub(tc.shares), pool.TotalShares(), "total shares after withdraw not equal")
		})
	}
}

func TestBasePool_Panic_OutOfBounds(t *testing.T) {
	pool, err := types.NewBasePool(sdk.NewInt(100), sdk.NewInt(100))
	require.NoError(t, err)

	assert.Panics(t, func() { pool.ShareValue(pool.TotalShares().Add(sdk.NewInt(1))) }, "ShareValue did not panic when shares > totalShares")
	assert.Panics(t, func() { pool.RemoveLiquidity(pool.TotalShares().Add(sdk.NewInt(1))) }, "RemoveLiquidity did not panic when shares > totalShares")
}

func TestBasePool_EmptyAndRefill(t *testing.T) {
	testCases := []struct {
		reservesA sdk.Int
		reservesB sdk.Int
	}{
		{i(1), i(1)},
		{i(100), i(100)},
		{i(100), i(10000000)},
		{i(1e5), i(5e6)},
		{i(1e6), i(5e6)},
		{i(1e15), i(7e15)},
		{i(1), i(6e18)},
		{i(1.345678e18), i(4.313456e18)},
		{i(145345664).Mul(exp(i(10), 26)), i(6432294561).Mul(exp(i(10), 20))},
		{i(465432423).Mul(exp(i(10), 50)), i(4565432).Mul(exp(i(10), 50))},
		{exp(i(2), 253), exp(i(2), 253)},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("reservesA=%s reservesB=%s", tc.reservesA, tc.reservesB), func(t *testing.T) {
			pool, err := types.NewBasePool(tc.reservesA, tc.reservesB)
			require.NoError(t, err)

			initialShares := pool.TotalShares()
			pool.RemoveLiquidity(initialShares)

			assert.True(t, pool.IsEmpty())
			assert.True(t, pool.TotalShares().IsZero(), "total shares are not depleted")

			pool.AddLiquidity(tc.reservesA, tc.reservesB)
			assert.Equal(t, initialShares, pool.TotalShares(), "total shares not equal")
		})
	}
}

func TestBasePool_Panics_AddLiquidity(t *testing.T) {
	assert.Panics(t, func() {
		pool, err := types.NewBasePool(i(1e6), i(1e6))
		require.NoError(t, err)

		pool.AddLiquidity(i(0), i(1e6))
	}, "did not panic when reserve A is zero")

	assert.Panics(t, func() {
		pool, err := types.NewBasePool(i(1e6), i(1e6))
		require.NoError(t, err)

		pool.AddLiquidity(i(-1), i(1e6))
	}, "did not panic when reserve A is negative")

	assert.Panics(t, func() {
		pool, err := types.NewBasePool(i(1e6), i(1e6))
		require.NoError(t, err)

		pool.AddLiquidity(i(1e6), i(0))
	}, "did not panic when reserve B is zero")

	assert.Panics(t, func() {
		pool, err := types.NewBasePool(i(1e6), i(1e6))
		require.NoError(t, err)

		pool.AddLiquidity(i(1e6), i(0))
	}, "did not panic when reserve B is zero")
}

func TestBasePool_Panics_RemoveLiquidity(t *testing.T) {
	assert.Panics(t, func() {
		pool, err := types.NewBasePool(i(1e6), i(1e6))
		require.NoError(t, err)

		pool.RemoveLiquidity(i(0))
	}, "did not panic when shares are zero")

	assert.Panics(t, func() {
		pool, err := types.NewBasePool(i(1e6), i(1e6))
		require.NoError(t, err)

		pool.RemoveLiquidity(i(-1))
	}, "did not panic when shares are negative")
}

func TestBasePool_ReservesOnlyDepletedWithLastShare(t *testing.T) {
	testCases := []struct {
		reservesA sdk.Int
		reservesB sdk.Int
	}{
		{i(5), i(5)},
		{i(100), i(100)},
		{i(100), i(10000000)},
		{i(1e5), i(5e6)}, {i(1e6), i(5e6)},
		{i(1e15), i(7e15)},
		{i(1), i(6e18)},
		{i(1.345678e18), i(4.313456e18)},
		{i(145345664).Mul(exp(i(10), 26)), i(6432294561).Mul(exp(i(10), 20))},
		{i(465432423).Mul(exp(i(10), 50)), i(4565432).Mul(exp(i(10), 50))},
		{exp(i(2), 253), exp(i(2), 253)},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("reservesA=%s reservesB=%s", tc.reservesA, tc.reservesB), func(t *testing.T) {
			pool, err := types.NewBasePool(tc.reservesA, tc.reservesB)
			require.NoError(t, err)

			initialShares := pool.TotalShares()
			pool.RemoveLiquidity(initialShares.Sub(i(1)))

			assert.False(t, pool.ReservesA().IsZero(), "reserves A equal to zero")
			assert.False(t, pool.ReservesB().IsZero(), "reserves B equal to zero")

			pool.RemoveLiquidity(i(1))
			assert.True(t, pool.IsEmpty())
		})
	}
}

func TestBasePool_Swap_ExactInput(t *testing.T) {
	testCases := []struct {
		reservesA      sdk.Int
		reservesB      sdk.Int
		exactInput     sdk.Int
		fee            sdk.Dec
		expectedOutput sdk.Int
		expectedFee    sdk.Int
	}{
		// test small pools
		{i(10), i(10), i(1), d("0.003"), i(0), i(1)},
		{i(10), i(10), i(3), d("0.003"), i(1), i(1)},
		{i(10), i(10), i(10), d("0.003"), i(4), i(1)},
		{i(10), i(10), i(91), d("0.003"), i(9), i(1)},
		// test fee values and ceil
		{i(1e6), i(1e6), i(1000), d("0.003"), i(996), i(3)},
		{i(1e6), i(1e6), i(1000), d("0.0031"), i(995), i(4)},
		{i(1e6), i(1e6), i(1000), d("0.0039"), i(995), i(4)},
		{i(1e6), i(1e6), i(1000), d("0.001"), i(998), i(1)},
		{i(1e6), i(1e6), i(1000), d("0.025"), i(974), i(25)},
		{i(1e6), i(1e6), i(1000), d("0.1"), i(899), i(100)},
		{i(1e6), i(1e6), i(1000), d("0.5"), i(499), i(500)},
		// test various random pools and swaps
		{i(10e6), i(500e6), i(1e6), d("0.0025"), i(45351216), i(2500)},
		{i(10e6), i(500e6), i(8e6), d("0.003456"), i(221794899), i(27648)},
		// test very large pools and swaps
		{exp(i(2), 250), exp(i(2), 250), exp(i(2), 249), d("0.003"), s("601876423139828614225164081027182620796370196819963934493551943901658899790"), s("2713877091499598330239944961141122840311015265600950719674787125185463976")},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("reservesA=%s reservesB=%s exactInput=%s fee=%s", tc.reservesA, tc.reservesB, tc.exactInput, tc.fee), func(t *testing.T) {
			poolA, err := types.NewBasePool(tc.reservesA, tc.reservesB)
			require.NoError(t, err)
			swapA, feeA := poolA.SwapExactAForB(tc.exactInput, tc.fee)

			poolB, err := types.NewBasePool(tc.reservesB, tc.reservesA)
			require.NoError(t, err)
			swapB, feeB := poolB.SwapExactBForA(tc.exactInput, tc.fee)

			// pool must be symmetric - if we swap reserves, then swap opposite direction
			// then the results should be equal
			require.Equal(t, swapA, swapB, "expected swap methods to have equal swap results")
			require.Equal(t, feeA, feeB, "expected swap methods to have equal fee results")
			require.Equal(t, poolA.ReservesA(), poolB.ReservesB(), "expected reserves A to be equal")
			require.Equal(t, poolA.ReservesB(), poolB.ReservesA(), "expected reserves B to be equal")

			assert.Equal(t, tc.expectedOutput, swapA, "returned swap not equal")
			assert.Equal(t, tc.expectedFee, feeA, "returned fee not equal")

			expectedReservesA := tc.reservesA.Add(tc.exactInput)
			expectedReservesB := tc.reservesB.Sub(tc.expectedOutput)

			assert.Equal(t, expectedReservesA, poolA.ReservesA(), "expected new reserves A not equal")
			assert.Equal(t, expectedReservesB, poolA.ReservesB(), "expected new reserves B not equal")
		})
	}
}

func TestBasePool_Swap_ExactOutput(t *testing.T) {
	testCases := []struct {
		reservesA     sdk.Int
		reservesB     sdk.Int
		exactOutput   sdk.Int
		fee           sdk.Dec
		expectedInput sdk.Int
		expectedFee   sdk.Int
	}{
		// test small pools
		{i(10), i(10), i(1), d("0.003"), i(3), i(1)},
		{i(10), i(10), i(9), d("0.003"), i(91), i(1)},
		// test fee values and ceil
		{i(1e6), i(1e6), i(996), d("0.003"), i(1000), i(3)},
		{i(1e6), i(1e6), i(995), d("0.0031"), i(1000), i(4)},
		{i(1e6), i(1e6), i(995), d("0.0039"), i(1000), i(4)},
		{i(1e6), i(1e6), i(998), d("0.001"), i(1000), i(1)},
		{i(1e6), i(1e6), i(974), d("0.025"), i(1000), i(25)},
		{i(1e6), i(1e6), i(899), d("0.1"), i(1000), i(100)},
		{i(1e6), i(1e6), i(499), d("0.5"), i(1000), i(500)},
		// test various random pools and swaps
		{i(10e6), i(500e6), i(45351216), d("0.0025"), i(1e6), i(2500)},
		{i(10e6), i(500e6), i(221794899), d("0.003456"), i(8e6), i(27648)},
		// test very large pools and swaps
		{exp(i(2), 250), exp(i(2), 250), s("601876423139828614225164081027182620796370196819963934493551943901658899790"), d("0.003"), s("904625697166532776746648320380374280103671755200316906558262375061821325311"), s("2713877091499598330239944961141122840311015265600950719674787125185463976")},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("reservesA=%s reservesB=%s exactOutput=%s fee=%s", tc.reservesA, tc.reservesB, tc.exactOutput, tc.fee), func(t *testing.T) {
			poolA, err := types.NewBasePool(tc.reservesA, tc.reservesB)
			require.NoError(t, err)
			swapA, feeA := poolA.SwapAForExactB(tc.exactOutput, tc.fee)

			poolB, err := types.NewBasePool(tc.reservesB, tc.reservesA)
			require.NoError(t, err)
			swapB, feeB := poolB.SwapBForExactA(tc.exactOutput, tc.fee)

			// pool must be symmetric - if we swap reserves, then swap opposite direction
			// then the results should be equal
			require.Equal(t, swapA, swapB, "expected swap methods to have equal swap results")
			require.Equal(t, feeA, feeB, "expected swap methods to have equal fee results")
			require.Equal(t, poolA.ReservesA(), poolB.ReservesB(), "expected reserves A to be equal")
			require.Equal(t, poolA.ReservesB(), poolB.ReservesA(), "expected reserves B to be equal")

			assert.Equal(t, tc.expectedInput.String(), swapA.String(), "returned swap not equal")
			assert.Equal(t, tc.expectedFee, feeA, "returned fee not equal")

			expectedReservesA := tc.reservesA.Add(tc.expectedInput)
			expectedReservesB := tc.reservesB.Sub(tc.exactOutput)

			assert.Equal(t, expectedReservesA, poolA.ReservesA(), "expected new reserves A not equal")
			assert.Equal(t, expectedReservesB, poolA.ReservesB(), "expected new reserves B not equal")
		})
	}
}

func TestBasePool_Panics_Swap_ExactInput(t *testing.T) {
	testCases := []struct {
		swap sdk.Int
		fee  sdk.Dec
	}{
		{i(0), d("0.003")},
		{i(-1), d("0.003")},
		{i(1), d("1")},
		{i(1), d("-0.003")},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("swap=%s fee=%s", tc.swap, tc.fee), func(t *testing.T) {
			assert.Panics(t, func() {
				pool, err := types.NewBasePool(i(1e6), i(1e6))
				require.NoError(t, err)

				pool.SwapExactAForB(tc.swap, tc.fee)
			}, "SwapExactAForB did not panic")

			assert.Panics(t, func() {
				pool, err := types.NewBasePool(i(1e6), i(1e6))
				require.NoError(t, err)

				pool.SwapExactBForA(tc.swap, tc.fee)
			}, "SwapExactBForA did not panic")
		})
	}
}

func TestBasePool_Panics_Swap_ExactOutput(t *testing.T) {
	testCases := []struct {
		swap sdk.Int
		fee  sdk.Dec
	}{
		{i(0), d("0.003")},
		{i(-1), d("0.003")},
		{i(1), d("1")},
		{i(1), d("-0.003")},
		{i(1000000), d("0.003")},
		{i(1000001), d("0.003")},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("swap=%s fee=%s", tc.swap, tc.fee), func(t *testing.T) {
			assert.Panics(t, func() {
				pool, err := types.NewBasePool(i(1e6), i(1e6))
				require.NoError(t, err)

				pool.SwapAForExactB(tc.swap, tc.fee)
			}, "SwapAForExactB did not panic")

			assert.Panics(t, func() {
				pool, err := types.NewBasePool(i(1e6), i(1e6))
				require.NoError(t, err)

				pool.SwapBForExactA(tc.swap, tc.fee)
			}, "SwapBForExactA did not panic")
		})
	}
}
