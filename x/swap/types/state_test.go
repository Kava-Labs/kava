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
		{"atoken", "btoken", "atoken:btoken"},
		{"btoken", "atoken", "atoken:btoken"},
		{"aaa", "aaaa", "aaa:aaaa"},
		{"aaaa", "aaa", "aaa:aaaa"},
		{"aaaa", "aaab", "aaaa:aaab"},
		{"aaab", "aaaa", "aaaa:aaab"},
		{"a001", "a002", "a001:a002"},
		{"a002", "a001", "a001:a002"},
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
		"pool_id": "ukava:usdx",
		"reserves_a": { "denom": "ukava", "amount": "1000000" },
		"reserves_b": { "denom": "usdx", "amount": "5000000" },
		"total_shares": "3000000"
	}`

	var record types.PoolRecord
	err := json.Unmarshal([]byte(raw), &record)
	require.NoError(t, err)

	assert.Equal(t, types.PoolID("ukava", "usdx"), record.PoolID)
	assert.Equal(t, ukava(1e6), record.ReservesA)
	assert.Equal(t, usdx(5e6), record.ReservesB)
	assert.Equal(t, i(3e6), record.TotalShares)
}

func TestState_PoolRecord_YamlEncoding(t *testing.T) {
	expected := `pool_id: ukava:usdx
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

func TestState_PoolRecord_Validations(t *testing.T) {
	validRecord := types.NewPoolRecord(
		sdk.NewCoins(usdx(500e6), ukava(100e6)),
		i(300e6),
	)
	testCases := []struct {
		name        string
		poolID      string
		reservesA   sdk.Coin
		reservesB   sdk.Coin
		totalShares sdk.Int
		expectedErr string
	}{
		{
			name:        "empty pool id",
			poolID:      "",
			reservesA:   validRecord.ReservesA,
			reservesB:   validRecord.ReservesB,
			totalShares: validRecord.TotalShares,
			expectedErr: "poolID must be set",
		},
		{
			name:        "no poolID tokens",
			poolID:      "ukavausdx",
			reservesA:   validRecord.ReservesA,
			reservesB:   validRecord.ReservesB,
			totalShares: validRecord.TotalShares,
			expectedErr: "poolID 'ukavausdx' is invalid",
		},
		{
			name:        "poolID empty tokens",
			poolID:      ":",
			reservesA:   validRecord.ReservesA,
			reservesB:   validRecord.ReservesB,
			totalShares: validRecord.TotalShares,
			expectedErr: "poolID ':' is invalid",
		},
		{
			name:        "poolID empty token a",
			poolID:      ":usdx",
			reservesA:   validRecord.ReservesA,
			reservesB:   validRecord.ReservesB,
			totalShares: validRecord.TotalShares,
			expectedErr: "poolID ':usdx' is invalid",
		},
		{
			name:        "poolID empty token b",
			poolID:      "ukava:",
			reservesA:   validRecord.ReservesA,
			reservesB:   validRecord.ReservesB,
			totalShares: validRecord.TotalShares,
			expectedErr: "poolID 'ukava:' is invalid",
		},
		{
			name:        "poolID is not sorted",
			poolID:      "usdx:ukava",
			reservesA:   validRecord.ReservesA,
			reservesB:   validRecord.ReservesB,
			totalShares: validRecord.TotalShares,
			expectedErr: "poolID 'usdx:ukava' is invalid",
		},
		{
			name:        "poolID has invalid denom a",
			poolID:      "UKAVA:usdx",
			reservesA:   validRecord.ReservesA,
			reservesB:   validRecord.ReservesB,
			totalShares: validRecord.TotalShares,
			expectedErr: "poolID 'UKAVA:usdx' is invalid",
		},
		{
			name:        "poolID has invalid denom b",
			poolID:      "ukava:USDX",
			reservesA:   validRecord.ReservesA,
			reservesB:   validRecord.ReservesB,
			totalShares: validRecord.TotalShares,
			expectedErr: "poolID 'ukava:USDX' is invalid",
		},
		{
			name:        "poolID has duplicate denoms",
			poolID:      "ukava:ukava",
			reservesA:   validRecord.ReservesA,
			reservesB:   validRecord.ReservesB,
			totalShares: validRecord.TotalShares,
			expectedErr: "poolID 'ukava:ukava' is invalid",
		},
		{
			name:        "poolID does not match reserve A",
			poolID:      "ukava:usdx",
			reservesA:   hard(5e6),
			reservesB:   validRecord.ReservesB,
			totalShares: validRecord.TotalShares,
			expectedErr: "poolID 'ukava:usdx' does not match reserves",
		},
		{
			name:        "poolID does not match reserve B",
			poolID:      "ukava:usdx",
			reservesA:   validRecord.ReservesA,
			reservesB:   hard(5e6),
			totalShares: validRecord.TotalShares,
			expectedErr: "poolID 'ukava:usdx' does not match reserves",
		},
		{
			name:        "negative reserve a",
			poolID:      "ukava:usdx",
			reservesA:   sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(-1)},
			reservesB:   validRecord.ReservesB,
			totalShares: validRecord.TotalShares,
			expectedErr: "pool 'ukava:usdx' has invalid reserves: -1ukava",
		},
		{
			name:        "zero reserve a",
			poolID:      "ukava:usdx",
			reservesA:   sdk.Coin{Denom: "ukava", Amount: sdk.ZeroInt()},
			reservesB:   validRecord.ReservesB,
			totalShares: validRecord.TotalShares,
			expectedErr: "pool 'ukava:usdx' has invalid reserves: 0ukava",
		},
		{
			name:        "negative reserve b",
			poolID:      "ukava:usdx",
			reservesA:   validRecord.ReservesA,
			reservesB:   sdk.Coin{Denom: "usdx", Amount: sdk.NewInt(-1)},
			totalShares: validRecord.TotalShares,
			expectedErr: "pool 'ukava:usdx' has invalid reserves: -1usdx",
		},
		{
			name:        "zero reserve b",
			poolID:      "ukava:usdx",
			reservesA:   validRecord.ReservesA,
			reservesB:   sdk.Coin{Denom: "usdx", Amount: sdk.ZeroInt()},
			totalShares: validRecord.TotalShares,
			expectedErr: "pool 'ukava:usdx' has invalid reserves: 0usdx",
		},
		{
			name:        "negative total shares",
			poolID:      validRecord.PoolID,
			reservesA:   validRecord.ReservesA,
			reservesB:   validRecord.ReservesB,
			totalShares: sdk.NewInt(-1),
			expectedErr: "pool 'ukava:usdx' has invalid total shares: -1",
		},
		{
			name:        "zero total shares",
			poolID:      validRecord.PoolID,
			reservesA:   validRecord.ReservesA,
			reservesB:   validRecord.ReservesB,
			totalShares: sdk.ZeroInt(),
			expectedErr: "pool 'ukava:usdx' has invalid total shares: 0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			record := types.PoolRecord{
				PoolID:      tc.poolID,
				ReservesA:   tc.reservesA,
				ReservesB:   tc.reservesB,
				TotalShares: tc.totalShares,
			}
			err := record.Validate()
			assert.EqualError(t, err, tc.expectedErr)
		})
	}
}

func TestState_PoolRecord_OrderedReserves(t *testing.T) {
	invalidOrder := types.NewPoolRecord(
		// force order to not be sorted
		sdk.Coins{usdx(500e6), ukava(100e6)},
		i(300e6),
	)
	assert.Error(t, invalidOrder.Validate())

	validOrder := types.NewPoolRecord(
		// force order to not be sorted
		sdk.Coins{ukava(500e6), usdx(100e6)},
		i(300e6),
	)
	assert.NoError(t, validOrder.Validate())

	record_1 := types.NewPoolRecord(sdk.NewCoins(usdx(500e6), ukava(100e6)), i(300e6))
	record_2 := types.NewPoolRecord(sdk.NewCoins(ukava(100e6), usdx(500e6)), i(300e6))
	// ensure no regresssions in NewCoins ordering
	assert.Equal(t, record_1, record_2)
	assert.Equal(t, types.PoolID("ukava", "usdx"), record_1.PoolID)
	assert.Equal(t, types.PoolID("ukava", "usdx"), record_2.PoolID)
}

func TestState_PoolRecords_Validation(t *testing.T) {
	validRecord := types.NewPoolRecord(
		sdk.NewCoins(usdx(500e6), ukava(100e6)),
		i(300e6),
	)

	invalidRecord := types.NewPoolRecord(
		sdk.NewCoins(usdx(500e6), ukava(100e6)),
		i(-1),
	)

	records := types.PoolRecords{
		validRecord,
	}
	assert.NoError(t, records.Validate())

	records = append(records, invalidRecord)
	err := records.Validate()
	assert.Error(t, err)
	assert.EqualError(t, err, "pool 'ukava:usdx' has invalid total shares: -1")
}

func TestState_PoolRecords_ValidateUniquePools(t *testing.T) {
	record_1 := types.NewPoolRecord(
		sdk.NewCoins(usdx(500e6), ukava(100e6)),
		i(300e6),
	)

	record_2 := types.NewPoolRecord(
		sdk.NewCoins(usdx(5000e6), ukava(1000e6)),
		i(3000e6),
	)

	record_3 := types.NewPoolRecord(
		sdk.NewCoins(usdx(5000e6), hard(1000e6)),
		i(3000e6),
	)

	validRecords := types.PoolRecords{record_1, record_3}
	assert.NoError(t, validRecords.Validate())

	invalidRecords := types.PoolRecords{record_1, record_2}
	assert.EqualError(t, invalidRecords.Validate(), "duplicate poolID 'ukava:usdx'")
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
		"pool_id": "ukava:usdx",
		"shares_owned": "3000000"
	}`

	var record types.ShareRecord
	err := json.Unmarshal([]byte(raw), &record)
	require.NoError(t, err)

	assert.Equal(t, "kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w", record.Depositor.String())
	assert.Equal(t, types.PoolID("ukava", "usdx"), record.PoolID)
	assert.Equal(t, i(3e6), record.SharesOwned)
}

func TestState_ShareRecord_YamlEncoding(t *testing.T) {
	expected := `depositor: kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w
pool_id: ukava:usdx
shares_owned: "3000000"
`
	depositor, err := sdk.AccAddressFromBech32("kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w")
	require.NoError(t, err)

	record := types.NewShareRecord(depositor, "ukava:usdx", i(3e6))
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

func TestState_ShareRecord_Validations(t *testing.T) {
	depositor, err := sdk.AccAddressFromBech32("kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w")
	require.NoError(t, err)
	validRecord := types.NewShareRecord(
		depositor,
		types.PoolID("ukava", "usdx"),
		i(30e6),
	)
	testCases := []struct {
		name        string
		depositor   sdk.AccAddress
		poolID      string
		sharesOwned sdk.Int
		expectedErr string
	}{
		{
			name:        "empty pool id",
			depositor:   validRecord.Depositor,
			poolID:      "",
			sharesOwned: validRecord.SharesOwned,
			expectedErr: "poolID must be set",
		},
		{
			name:        "no poolID tokens",
			depositor:   validRecord.Depositor,
			poolID:      "ukavausdx",
			sharesOwned: validRecord.SharesOwned,
			expectedErr: "poolID 'ukavausdx' is invalid",
		},
		{
			name:        "poolID empty tokens",
			depositor:   validRecord.Depositor,
			poolID:      ":",
			sharesOwned: validRecord.SharesOwned,
			expectedErr: "poolID ':' is invalid",
		},
		{
			name:        "poolID empty token a",
			depositor:   validRecord.Depositor,
			poolID:      ":usdx",
			sharesOwned: validRecord.SharesOwned,
			expectedErr: "poolID ':usdx' is invalid",
		},
		{
			name:        "poolID empty token b",
			depositor:   validRecord.Depositor,
			poolID:      "ukava:",
			sharesOwned: validRecord.SharesOwned,
			expectedErr: "poolID 'ukava:' is invalid",
		},
		{
			name:        "poolID is not sorted",
			depositor:   validRecord.Depositor,
			poolID:      "usdx:ukava",
			sharesOwned: validRecord.SharesOwned,
			expectedErr: "poolID 'usdx:ukava' is invalid",
		},
		{
			name:        "poolID has invalid denom a",
			depositor:   validRecord.Depositor,
			poolID:      "UKAVA:usdx",
			sharesOwned: validRecord.SharesOwned,
			expectedErr: "poolID 'UKAVA:usdx' is invalid",
		},
		{
			name:        "poolID has invalid denom b",
			depositor:   validRecord.Depositor,
			poolID:      "ukava:USDX",
			sharesOwned: validRecord.SharesOwned,
			expectedErr: "poolID 'ukava:USDX' is invalid",
		},
		{
			name:        "poolID has duplicate denoms",
			depositor:   validRecord.Depositor,
			poolID:      "ukava:ukava",
			sharesOwned: validRecord.SharesOwned,
			expectedErr: "poolID 'ukava:ukava' is invalid",
		},
		{
			name:        "negative total shares",
			depositor:   validRecord.Depositor,
			poolID:      validRecord.PoolID,
			sharesOwned: sdk.NewInt(-1),
			expectedErr: "depositor 'kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w' and pool 'ukava:usdx' has invalid total shares: -1",
		},
		{
			name:        "zero total shares",
			depositor:   validRecord.Depositor,
			poolID:      validRecord.PoolID,
			sharesOwned: sdk.ZeroInt(),
			expectedErr: "depositor 'kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w' and pool 'ukava:usdx' has invalid total shares: 0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			record := types.ShareRecord{
				Depositor:   tc.depositor,
				PoolID:      tc.poolID,
				SharesOwned: tc.sharesOwned,
			}
			err := record.Validate()
			assert.EqualError(t, err, tc.expectedErr)
		})
	}
}

func TestState_ShareRecords_Validation(t *testing.T) {
	depositor, err := sdk.AccAddressFromBech32("kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w")
	require.NoError(t, err)

	validRecord := types.NewShareRecord(
		depositor,
		types.PoolID("hard", "usdx"),
		i(300e6),
	)

	invalidRecord := types.NewShareRecord(
		depositor,
		types.PoolID("hard", "usdx"),
		i(-1),
	)

	records := types.ShareRecords{
		validRecord,
	}
	assert.NoError(t, records.Validate())

	records = append(records, invalidRecord)
	err = records.Validate()
	assert.Error(t, err)
	assert.EqualError(t, err, "depositor 'kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w' and pool 'hard:usdx' has invalid total shares: -1")
}

func TestState_ShareRecords_ValidateUniqueShareRecords(t *testing.T) {
	depositor_1, err := sdk.AccAddressFromBech32("kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w")
	require.NoError(t, err)

	depositor_2, err := sdk.AccAddressFromBech32("kava1esagqd83rhqdtpy5sxhklaxgn58k2m3s3mnpea")
	require.NoError(t, err)

	record_1 := types.NewShareRecord(depositor_1, "ukava:usdx", i(20e6))
	record_2 := types.NewShareRecord(depositor_1, "ukava:usdx", i(10e6))
	record_3 := types.NewShareRecord(depositor_1, "hard:usdx", i(20e6))
	record_4 := types.NewShareRecord(depositor_2, "ukava:usdx", i(20e6))

	validRecords := types.ShareRecords{record_1, record_3, record_4}
	assert.NoError(t, validRecords.Validate())

	invalidRecords := types.ShareRecords{record_1, record_3, record_2, record_4}
	assert.EqualError(t, invalidRecords.Validate(), "duplicate depositor 'kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w' and poolID 'ukava:usdx'")
}
