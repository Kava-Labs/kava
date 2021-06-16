package types_test

import (
	"fmt"
	"testing"

	types "github.com/kava-labs/kava/x/swap/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// create a new ukava coin from int64
func ukava(amount int64) sdk.Coin {
	return sdk.NewCoin("ukava", sdk.NewInt(amount))
}

// create a new usdx coin from int64
func usdx(amount int64) sdk.Coin {
	return sdk.NewCoin("usdx", sdk.NewInt(amount))
}

func TestDenominatedPool_NewDenominatedPool_Validation(t *testing.T) {
	testCases := []struct {
		reservesA   sdk.Coin
		reservesB   sdk.Coin
		expectedErr string
	}{
		{ukava(0), usdx(1e6), "invalid pool: reserves must have two denominations"},
		{ukava(1e6), usdx(0), "invalid pool: reserves must have two denominations"},
		{usdx(0), ukava(1e6), "invalid pool: reserves must have two denominations"},
		{usdx(0), ukava(1e6), "invalid pool: reserves must have two denominations"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("reservesA=%s reservesB=%s", tc.reservesA, tc.reservesB), func(t *testing.T) {
			pool, err := types.NewDenominatedPool(sdk.NewCoins(tc.reservesA, tc.reservesB))
			require.EqualError(t, err, tc.expectedErr)
			assert.Nil(t, pool)
		})
	}
}

func TestDenominatedPool_NewDenominatedPoolWithExistingShares_Validation(t *testing.T) {
	testCases := []struct {
		reservesA   sdk.Coin
		reservesB   sdk.Coin
		totalShares sdk.Int
		expectedErr string
	}{
		{ukava(0), usdx(1e6), i(1), "invalid pool: reserves must have two denominations"},
		{usdx(0), ukava(1e6), i(1), "invalid pool: reserves must have two denominations"},
		{ukava(1e6), usdx(1e6), i(0), "invalid pool: total shares must be greater than zero"},
		{usdx(1e6), ukava(1e6), i(-1), "invalid pool: total shares must be greater than zero"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("reservesA=%s reservesB=%s", tc.reservesA, tc.reservesB), func(t *testing.T) {
			pool, err := types.NewDenominatedPoolWithExistingShares(sdk.NewCoins(tc.reservesA, tc.reservesB), tc.totalShares)
			require.EqualError(t, err, tc.expectedErr)
			assert.Nil(t, pool)
		})
	}
}

func TestDenominatedPool_InitialState(t *testing.T) {
	reserves := sdk.NewCoins(ukava(1e6), usdx(5e6))
	totalShares := i(2236067)

	pool, err := types.NewDenominatedPool(reserves)
	require.NoError(t, err)

	assert.Equal(t, pool.Reserves(), reserves)
	assert.Equal(t, pool.TotalShares(), totalShares)
}

func TestDenominatedPool_InitialState_ExistingShares(t *testing.T) {
	reserves := sdk.NewCoins(ukava(1e6), usdx(5e6))
	totalShares := i(2e6)

	pool, err := types.NewDenominatedPoolWithExistingShares(reserves, totalShares)
	require.NoError(t, err)

	assert.Equal(t, pool.Reserves(), reserves)
	assert.Equal(t, pool.TotalShares(), totalShares)
}

func TestDenominatedPool_ShareValue(t *testing.T) {
	reserves := sdk.NewCoins(ukava(10e6), usdx(50e6))

	pool, err := types.NewDenominatedPool(reserves)
	require.NoError(t, err)

	assert.Equal(t, reserves, pool.ShareValue(pool.TotalShares()))

	halfReserves := sdk.NewCoins(ukava(4999999), usdx(24999998))
	assert.Equal(t, halfReserves, pool.ShareValue(pool.TotalShares().Quo(i(2))))
}

func TestDenominatedPool_AddLiquidity(t *testing.T) {
	reserves := sdk.NewCoins(ukava(10e6), usdx(50e6))
	desired := sdk.NewCoins(ukava(1e6), usdx(1e6))

	pool, err := types.NewDenominatedPool(reserves)
	require.NoError(t, err)
	initialShares := pool.TotalShares()

	deposit, shares := pool.AddLiquidity(desired)
	require.True(t, shares.IsPositive())
	require.True(t, deposit.IsAllPositive())

	assert.Equal(t, reserves.Add(deposit...), pool.Reserves())
	assert.Equal(t, initialShares.Add(shares), pool.TotalShares())
}

func TestDenominatedPool_RemoveLiquidity(t *testing.T) {
	reserves := sdk.NewCoins(ukava(10e6), usdx(50e6))

	pool, err := types.NewDenominatedPool(reserves)
	require.NoError(t, err)

	withdraw := pool.RemoveLiquidity(pool.TotalShares())

	assert.True(t, pool.Reserves().IsZero())
	assert.True(t, pool.TotalShares().IsZero())
	assert.True(t, pool.IsEmpty())
	assert.Equal(t, reserves, withdraw)
}
