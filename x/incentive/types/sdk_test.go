package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"

	"github.com/kava-labs/kava/x/incentive/types"
)

func TestGetTotalVestingPeriodLength(t *testing.T) {
	testCases := []struct {
		name        string
		periods     vesting.Periods
		expectedVal int64
	}{
		{
			name: "two period lengths are added together",
			periods: vesting.Periods{
				{
					Length: 100,
				},
				{
					Length: 200,
				},
			},
			expectedVal: 300,
		},
		{
			name:        "no periods returns zero",
			periods:     nil,
			expectedVal: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			length := types.GetTotalVestingPeriodLength(tc.periods)
			require.Equal(t, tc.expectedVal, length)
		})
	}
}

func TestMultiplyCoins(t *testing.T) {
	testCases := []struct {
		name     string
		coins    sdk.Coins
		multiple sdk.Dec
		expected sdk.Coins
	}{
		{
			name: "decimals are rounded to nearest even, up",
			coins: sdk.NewCoins(
				sdk.NewInt64Coin("hard", 1e3),
			),
			multiple: sdk.MustNewDecFromStr("3.1415"),
			expected: sdk.NewCoins(
				sdk.NewInt64Coin("hard", 3142),
			),
		},
		{
			name: "decimals are rounded to nearest even, down",
			coins: sdk.NewCoins(
				sdk.NewInt64Coin("hard", 1e7),
			),
			multiple: sdk.MustNewDecFromStr("3.14159265"),
			expected: sdk.NewCoins(
				sdk.NewInt64Coin("hard", 31415926),
			),
		},
		{
			name: "multiple coin amounts are multiplied",
			coins: sdk.NewCoins(
				sdk.NewInt64Coin("hard", 1),
				sdk.NewInt64Coin("ukava", 1e18),
			),
			multiple: sdk.MustNewDecFromStr("2.000000000000000002"),
			expected: sdk.NewCoins(
				sdk.NewInt64Coin("hard", 2),
				sdk.NewInt64Coin("ukava", 2_000_000_000_000_000_002),
			),
		},
		{
			name:     "empty coins return nil",
			coins:    sdk.Coins{},
			multiple: sdk.MustNewDecFromStr("2.5"),
			expected: nil,
		},
		{
			name:     "nil coins return nil",
			coins:    nil,
			multiple: sdk.MustNewDecFromStr("2.5"),
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t,
				tc.expected,
				types.MultiplyCoins(tc.coins, tc.multiple),
			)
		})
	}
}

func TestFilterCoins(t *testing.T) {
	testCases := []struct {
		name     string
		coins    sdk.Coins
		denoms   []string
		expected sdk.Coins
	}{
		{
			name: "non-empty filter selects subset of coins",
			coins: sdk.NewCoins(
				sdk.NewInt64Coin("hard", 1e3),
				sdk.NewInt64Coin("ukava", 2e3),
				sdk.NewInt64Coin("btc", 3e3),
			),
			denoms: []string{"hard", "btc"},
			expected: sdk.NewCoins(
				sdk.NewInt64Coin("hard", 1e3),
				sdk.NewInt64Coin("btc", 3e3),
			),
		},
		{
			name:     "when coins are nil a non-empty filter returns nil coins",
			coins:    nil,
			denoms:   []string{"hard", "btc"},
			expected: nil,
		},
		{
			name: "nil filter returns original coins",
			coins: sdk.NewCoins(
				sdk.NewInt64Coin("hard", 1e3),
				sdk.NewInt64Coin("ukava", 2e3),
			),
			denoms: nil,
			expected: sdk.NewCoins(
				sdk.NewInt64Coin("hard", 1e3),
				sdk.NewInt64Coin("ukava", 2e3),
			),
		},
		{
			name: "empty filter returns original coins",
			coins: sdk.NewCoins(
				sdk.NewInt64Coin("hard", 1e3),
				sdk.NewInt64Coin("ukava", 2e3),
			),
			denoms: []string{},
			expected: sdk.NewCoins(
				sdk.NewInt64Coin("hard", 1e3),
				sdk.NewInt64Coin("ukava", 2e3),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t,
				tc.expected,
				types.FilterCoins(tc.coins, tc.denoms),
			)
		})
	}
}
