package types_test

import (
	"strings"
	"testing"

	types "github.com/kava-labs/kava/x/swap/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllowedPool_Validation(t *testing.T) {
	testCases := []struct {
		name        string
		allowedPool types.AllowedPool
		expectedErr string
	}{
		{
			name:        "blank token a",
			allowedPool: types.NewAllowedPool("", "ukava"),
			expectedErr: "invalid denom: ",
		},
		{
			name:        "blank token b",
			allowedPool: types.NewAllowedPool("ukava", ""),
			expectedErr: "invalid denom: ",
		},
		{
			name:        "invalid token a",
			allowedPool: types.NewAllowedPool("1ukava", "ukava"),
			expectedErr: "invalid denom: 1ukava",
		},
		{
			name:        "invalid token b",
			allowedPool: types.NewAllowedPool("ukava", "1ukava"),
			expectedErr: "invalid denom: 1ukava",
		},
		{
			name:        "no uppercase letters token a",
			allowedPool: types.NewAllowedPool("uKava", "ukava"),
			expectedErr: "invalid denom: uKava",
		},
		{
			name:        "no uppercase letters token b",
			allowedPool: types.NewAllowedPool("ukava", "UKAVA"),
			expectedErr: "invalid denom: UKAVA",
		},
		{
			name:        "matching tokens",
			allowedPool: types.NewAllowedPool("ukava", "ukava"),
			expectedErr: "pool cannot have two tokens of the same type, received 'ukava' and 'ukava'",
		},
		{
			name:        "invalid token order",
			allowedPool: types.NewAllowedPool("usdx", "ukava"),
			expectedErr: "invalid token order: 'ukava' must come before 'usdx'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.allowedPool.Validate()
			assert.EqualError(t, err, tc.expectedErr)
		})
	}
}

// ensure no regression in case insentive token matching if
// sdk.ValidateDenom ever allows upper case letters
func TestAllowedPool_TokenMatch(t *testing.T) {
	allowedPool := types.NewAllowedPool("UKAVA", "ukava")
	err := allowedPool.Validate()
	assert.Error(t, err)

	allowedPool = types.NewAllowedPool("hard", "haRd")
	err = allowedPool.Validate()
	assert.Error(t, err)

	allowedPool = types.NewAllowedPool("Usdx", "uSdX")
	err = allowedPool.Validate()
	assert.Error(t, err)
}

func TestAllowedPool_String(t *testing.T) {
	allowedPool := types.NewAllowedPool("hard", "ukava")
	require.NoError(t, allowedPool.Validate())

	output := `AllowedPool:
  Name: hard/ukava
	Token A: hard
	Token B: ukava
`
	assert.Equal(t, output, allowedPool.String())
}

func TestAllowedPool_Name(t *testing.T) {
	testCases := []struct {
		tokens string
		name   string
	}{
		{
			tokens: "atoken btoken",
			name:   "atoken/btoken",
		},
		{
			tokens: "aaa aaaa",
			name:   "aaa/aaaa",
		},
		{
			tokens: "aaaa aaab",
			name:   "aaaa/aaab",
		},
		{
			tokens: "a001 a002",
			name:   "a001/a002",
		},
		{
			tokens: "hard ukava",
			name:   "hard/ukava",
		},
		{
			tokens: "bnb hard",
			name:   "bnb/hard",
		},
		{
			tokens: "bnb xrpb",
			name:   "bnb/xrpb",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.tokens, func(t *testing.T) {
			tokens := strings.Split(tc.tokens, " ")
			require.Equal(t, 2, len(tokens))

			allowedPool := types.NewAllowedPool(tokens[0], tokens[1])
			require.NoError(t, allowedPool.Validate())

			assert.Equal(t, tc.name, allowedPool.Name())
		})
	}
}

func TestAllowedPools_Validate(t *testing.T) {
	testCases := []struct {
		name         string
		allowedPools types.AllowedPools
		expectedErr  string
	}{
		{
			name: "invalid pool",
			allowedPools: types.NewAllowedPools(
				types.NewAllowedPool("hard", "ukava"),
				types.NewAllowedPool("HARD", "UKAVA"),
			),
			expectedErr: "invalid denom: HARD",
		},
		{
			name: "duplicate pool",
			allowedPools: types.NewAllowedPools(
				types.NewAllowedPool("hard", "ukava"),
				types.NewAllowedPool("hard", "ukava"),
			),
			expectedErr: "duplicate pool: hard/ukava",
		},
		{
			name: "duplicate pools",
			allowedPools: types.NewAllowedPools(
				types.NewAllowedPool("hard", "ukava"),
				types.NewAllowedPool("bnb", "usdx"),
				types.NewAllowedPool("btcb", "xrpb"),
				types.NewAllowedPool("bnb", "usdx"),
			),
			expectedErr: "duplicate pool: bnb/usdx",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.allowedPools.Validate()
			assert.EqualError(t, err, tc.expectedErr)
		})
	}
}

// TODO: example tests for init share math, cover edge cases
func TestPool_InitShares(t *testing.T) {
	a := sdk.NewCoin("ukava", sdk.NewInt(1e6))
	b := sdk.NewCoin("usdx", sdk.NewInt(5e6))
	pool := types.NewPool(a, b)

	assert.Equal(t, a, pool.ReservesA)
	assert.Equal(t, b, pool.ReservesB)

	assert.Equal(t, sdk.NewInt(2236067), pool.TotalShares)
}

func TestPool_Name(t *testing.T) {
	a := sdk.NewCoin("ukava", sdk.NewInt(1e6))
	b := sdk.NewCoin("usdx", sdk.NewInt(5e6))
	pool := types.NewPool(a, b)

	assert.Equal(t, "ukava/usdx", pool.Name())
}

func TestPool_ShareValue(t *testing.T) {
	a := sdk.NewCoin("ukava", sdk.NewInt(1e6))
	b := sdk.NewCoin("usdx", sdk.NewInt(5e6))
	pool := types.NewPool(a, b)

	assert.Equal(t, sdk.NewCoins(a, b), pool.ShareValue(pool.TotalShares))
}
