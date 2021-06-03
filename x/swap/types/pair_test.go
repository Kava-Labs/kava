package types_test

import (
	"strings"
	"testing"

	types "github.com/kava-labs/kava/x/swap/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPair_Validation(t *testing.T) {
	testCases := []struct {
		name        string
		pair        types.Pair
		expectedErr string
	}{
		{
			name:        "blank token a",
			pair:        types.NewPair("", "ukava"),
			expectedErr: "invalid denom: ",
		},
		{
			name:        "blank token b",
			pair:        types.NewPair("ukava", ""),
			expectedErr: "invalid denom: ",
		},
		{
			name:        "invalid token a",
			pair:        types.NewPair("1ukava", "ukava"),
			expectedErr: "invalid denom: 1ukava",
		},
		{
			name:        "invalid token b",
			pair:        types.NewPair("ukava", "1ukava"),
			expectedErr: "invalid denom: 1ukava",
		},
		{
			name:        "no uppercase letters token a",
			pair:        types.NewPair("uKava", "ukava"),
			expectedErr: "invalid denom: uKava",
		},
		{
			name:        "no uppercase letters token b",
			pair:        types.NewPair("ukava", "UKAVA"),
			expectedErr: "invalid denom: UKAVA",
		},
		{
			name:        "matching tokens",
			pair:        types.NewPair("ukava", "ukava"),
			expectedErr: "pair cannot have two tokens of the same type, received 'ukava' and 'ukava'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.pair.Validate()
			assert.EqualError(t, err, tc.expectedErr)
		})
	}
}

// ensure no regression in case insentive token matching if
// sdk.ValidateDenom ever allows upper case letters
func TestPair_TokenMatch(t *testing.T) {
	pair := types.NewPair("UKAVA", "ukava")
	err := pair.Validate()
	assert.Error(t, err)

	pair = types.NewPair("hard", "haRd")
	err = pair.Validate()
	assert.Error(t, err)

	pair = types.NewPair("Usdx", "uSdX")
	err = pair.Validate()
	assert.Error(t, err)
}

func TestPair_String(t *testing.T) {
	pair := types.NewPair("ukava", "hard")
	require.NoError(t, pair.Validate())

	output := `Pair:
  Name: hard/ukava
	Token A: ukava
	Token B: hard
`
	assert.Equal(t, output, pair.String())
}

func TestPair_Name(t *testing.T) {
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
			tokens: "ukava hard",
			name:   "hard/ukava",
		},
		{
			tokens: "hard bnb",
			name:   "bnb/hard",
		},
		{
			tokens: "xrpb bnb",
			name:   "bnb/xrpb",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.tokens, func(t *testing.T) {
			tokens := strings.Split(tc.tokens, " ")
			require.Equal(t, 2, len(tokens))

			pair := types.NewPair(tokens[0], tokens[1])
			require.NoError(t, pair.Validate())

			pairReverse := types.NewPair(tokens[1], tokens[0])
			require.NoError(t, pairReverse.Validate())

			assert.Equal(t, tc.name, pair.Name())
			assert.Equal(t, tc.name, pairReverse.Name())
		})
	}
}

func TestPairs_Validate(t *testing.T) {
	testCases := []struct {
		name        string
		pairs       types.Pairs
		expectedErr string
	}{
		{
			name: "invalid pair",
			pairs: types.NewPairs(
				types.NewPair("ukava", "hard"),
				types.NewPair("HARD", "UKAVA"),
			),
			expectedErr: "invalid denom: HARD",
		},
		{
			name: "duplicate pair",
			pairs: types.NewPairs(
				types.NewPair("ukava", "hard"),
				types.NewPair("hard", "ukava"),
			),
			expectedErr: "duplicate pair: hard/ukava",
		},
		{
			name: "duplicate pairs",
			pairs: types.NewPairs(
				types.NewPair("hard", "ukava"),
				types.NewPair("usdx", "bnb"),
				types.NewPair("btcb", "xrpb"),
				types.NewPair("bnb", "usdx"),
			),
			expectedErr: "duplicate pair: bnb/usdx",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.pairs.Validate()
			assert.EqualError(t, err, tc.expectedErr)
		})
	}
}
