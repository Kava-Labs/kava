package types_test

import (
	"encoding/json"
	"testing"

	types "github.com/kava-labs/kava/x/swap/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
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
	totalShares := sdk.NewInt(30e6)

	poolRecord := types.NewPoolRecord(reserves, totalShares)

	assert.Equal(t, reserves[0], poolRecord.ReservesA)
	assert.Equal(t, reserves[1], poolRecord.ReservesB)
	assert.Equal(t, reserves, poolRecord.Reserves())
	assert.Equal(t, totalShares, poolRecord.TotalShares)

	assert.PanicsWithValue(t, "reserves must have two denominations", func() {
		reserves := sdk.NewCoins(ukava(10e6))
		_ = types.NewPoolRecord(reserves, totalShares)
	}, "expected panic with 1 coin in reserves")

	assert.PanicsWithValue(t, "reserves must have two denominations", func() {
		reserves := sdk.NewCoins(ukava(10e6), hard(1e6), usdx(20e6))
		_ = types.NewPoolRecord(reserves, totalShares)
	}, "expected panic with 3 coins in reserves")
}

func TestState_NewPoolRecordFromPool(t *testing.T) {
	reserves := sdk.NewCoins(usdx(50e6), ukava(10e6))

	pool, err := types.NewDenominatedPool(reserves)
	require.NoError(t, err)

	record := types.NewPoolRecordFromPool(pool)

	assert.Equal(t, types.PoolID("ukava", "usdx"), record.PoolID)
	assert.Equal(t, ukava(10e6), record.ReservesA)
	assert.Equal(t, record.ReservesB, usdx(50e6))
	assert.Equal(t, pool.TotalShares(), record.TotalShares)
	assert.Equal(t, sdk.NewCoins(ukava(10e6), usdx(50e6)), record.Reserves())
	assert.Nil(t, record.Validate())
}

func TestState_PoolRecord_JSONEncoding(t *testing.T) {
	raw := `{
		"pool_id": "ukava/usdx",
		"reserves_a": { "denom": "ukava", "amount": "1000000" },
		"reserves_b": { "denom": "usdx", "amount": "5000000" },
		"total_shares": "3000000"
	}`

	var record types.PoolRecord
	err := json.Unmarshal([]byte(raw), &record)
	require.NoError(t, err)

	assert.Equal(t, "ukava/usdx", record.PoolID)
	assert.Equal(t, ukava(1e6), record.ReservesA)
	assert.Equal(t, usdx(5e6), record.ReservesB)
	assert.Equal(t, i(3e6), record.TotalShares)
}

func TestState_PoolRecord_YamlEncoding(t *testing.T) {
	expected := `pool_id: ukava/usdx
reserves_a:
  denom: ukava
  amount: "1000000"
reserves_b:
  denom: usdx
  amount: "5000000"
total_shares: "3000000"
`
	record := types.NewPoolRecord(sdk.NewCoins(ukava(1e6), usdx(5e6)), i(3e6))
	data, err := yaml.Marshal(record)
	require.NoError(t, err)

	assert.Equal(t, expected, string(data))
}

func TestState_InvalidPoolRecordNegativeShares(t *testing.T) {
	record := types.PoolRecord{
		PoolID:      types.PoolID("ukava", "usdx"),
		ReservesA:   usdx(50e6),
		ReservesB:   ukava(10e6),
		TotalShares: i(-100),
	}
	require.Error(t, record.Validate())
}

func TestState_InvalidPoolRecordNegativeReserves(t *testing.T) {
	recordA := types.PoolRecord{
		PoolID:      types.PoolID("ukava", "usdx"),
		ReservesA:   sdk.Coin{Denom: "usdx", Amount: i(-50e6)},
		ReservesB:   ukava(10e6),
		TotalShares: i(100),
	}
	require.Error(t, recordA.Validate())

	recordB := types.PoolRecord{
		PoolID:      types.PoolID("ukava", "usdx"),
		ReservesA:   usdx(50e6),
		ReservesB:   sdk.Coin{Denom: "ukava", Amount: i(-10e6)},
		TotalShares: i(100),
	}
	require.Error(t, recordB.Validate())
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

func TestState_ShareRecord_JSONEncoding(t *testing.T) {
	raw := `{
		"depositor": "kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w",
		"pool_id": "ukava/usdx",
		"shares_owned": "3000000"
	}`

	var record types.ShareRecord
	err := json.Unmarshal([]byte(raw), &record)
	require.NoError(t, err)

	assert.Equal(t, "kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w", record.Depositor.String())
	assert.Equal(t, "ukava/usdx", record.PoolID)
	assert.Equal(t, i(3e6), record.SharesOwned)
}

func TestState_ShareRecord_YamlEncoding(t *testing.T) {
	expected := `depositor: kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w
pool_id: ukava/usdx
shares_owned: "3000000"
`
	depositor, err := sdk.AccAddressFromBech32("kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w")
	require.NoError(t, err)

	record := types.NewShareRecord(depositor, "ukava/usdx", i(3e6))
	data, err := yaml.Marshal(record)
	require.NoError(t, err)

	assert.Equal(t, expected, string(data))
}

func TestState_InvalidShareRecordEmptyDepositor(t *testing.T) {
	record := types.ShareRecord{
		Depositor:   sdk.AccAddress{},
		PoolID:      types.PoolID("ukava", "usdx"),
		SharesOwned: sdk.NewInt(1e6),
	}
	require.Error(t, record.Validate())
}

func TestState_InvalidShareRecordNegativeShares(t *testing.T) {
	record := types.ShareRecord{
		Depositor:   sdk.AccAddress("some user"),
		PoolID:      types.PoolID("ukava", "usdx"),
		SharesOwned: sdk.NewInt(-1e6),
	}
	require.Error(t, record.Validate())
}
