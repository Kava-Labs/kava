package types_test

import (
	"testing"

	types "github.com/kava-labs/kava/x/swap/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestState_PoolID(t *testing.T) {
	testCases := []struct {
		reserveA   string
		reserveB   string
		expectedID string
	}{
		{"atoken", "btoken", "atoken/btoken"},
		{"btoken", "atoken", "atoken/btoken"},
		{"aaa", "aaaa", "aaa/aaaa"},
		{"aaaa", "aaa", "aaa/aaaa"},
		{"aaaa", "aaab", "aaaa/aaab"},
		{"aaab", "aaaa", "aaaa/aaab"},
		{"a001", "a002", "a001/a002"},
		{"a002", "a001", "a001/a002"},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedID, types.PoolID(tc.reserveA, tc.reserveB))
		assert.Equal(t, tc.expectedID, types.PoolID(tc.reserveB, tc.reserveA))

		assert.Equal(t, tc.expectedID, types.PoolIDFromCoins(sdk.NewCoins(sdk.NewCoin(tc.reserveA, i(1)), sdk.NewCoin(tc.reserveB, i(1)))))
		assert.Equal(t, tc.expectedID, types.PoolIDFromCoins(sdk.NewCoins(sdk.NewCoin(tc.reserveB, i(1)), sdk.NewCoin(tc.reserveA, i(1)))))
	}
}

func TestState_NewPoolRecord(t *testing.T) {
	reserves := sdk.NewCoins(usdx(50e6), ukava(10e6))

	pool, err := types.NewDenominatedPool(reserves)
	require.NoError(t, err)

	record := types.NewPoolRecord(pool)

	assert.Equal(t, types.PoolID("ukava", "usdx"), record.PoolID)
	assert.Equal(t, ukava(10e6), record.ReservesA)
	assert.Equal(t, record.ReservesB, usdx(50e6))
	assert.Equal(t, pool.TotalShares(), record.TotalShares)
	assert.Equal(t, sdk.NewCoins(ukava(10e6), usdx(50e6)), record.Reserves())
}

func TestState_NewShareRecord(t *testing.T) {
	depositor := sdk.AccAddress("some user")
	poolID := types.PoolID("ukava", "usdx")
	shares := sdk.NewInt(1e6)

	record := types.NewShareRecord(depositor, poolID, shares)

	assert.Equal(t, depositor, record.Depositor)
	assert.Equal(t, poolID, record.PoolID)
	assert.Equal(t, shares, record.SharesOwned)
}
