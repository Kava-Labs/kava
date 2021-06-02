package types

import (
	"fmt"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPair_Validation(t *testing.T) {
	testCases := []struct {
		name        string
		pair        Pair
		expectedErr string
	}{
		{
			name:        "blank token a",
			pair:        NewPair("", "ukava", sdk.ZeroDec()),
			expectedErr: "invalid denom: ",
		},
		{
			name:        "blank token b",
			pair:        NewPair("ukava", "", sdk.ZeroDec()),
			expectedErr: "invalid denom: ",
		},
		{
			name:        "invalid token a",
			pair:        NewPair("1ukava", "ukava", sdk.ZeroDec()),
			expectedErr: "invalid denom: 1ukava",
		},
		{
			name:        "invalid token b",
			pair:        NewPair("ukava", "1ukava", sdk.ZeroDec()),
			expectedErr: "invalid denom: 1ukava",
		},
		{
			name:        "no uppercase letters token a",
			pair:        NewPair("uKava", "ukava", sdk.ZeroDec()),
			expectedErr: "invalid denom: uKava",
		},
		{
			name:        "no uppercase letters token b",
			pair:        NewPair("ukava", "UKAVA", sdk.ZeroDec()),
			expectedErr: "invalid denom: UKAVA",
		},
		{
			name:        "matching tokens",
			pair:        NewPair("ukava", "ukava", sdk.ZeroDec()),
			expectedErr: "pair cannot have two tokens of the same type, received 'ukava' and 'ukava'",
		},
		{
			name:        "nil reward apy",
			pair:        NewPair("ukava", "hard", sdk.Dec{}),
			expectedErr: fmt.Sprintf("invalid reward apy: %s", sdk.Dec{}),
		},
		{
			name:        "nil reward apy",
			pair:        NewPair("ukava", "hard", sdk.NewDec(-1)),
			expectedErr: fmt.Sprintf("invalid reward apy: %s", sdk.NewDec(-1)),
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
	pair := NewPair("UKAVA", "UKAVA", sdk.ZeroDec())
	err := pair.Validate()
	assert.Error(t, err)

	pair = NewPair("haRd", "haRd", sdk.ZeroDec())
	err = pair.Validate()
	assert.Error(t, err)

	pair = NewPair("Usdx", "Usdx", sdk.ZeroDec())
	err = pair.Validate()
	assert.Error(t, err)
}

func TestPair_String(t *testing.T) {
	apy, err := sdk.NewDecFromStr("0.5")
	require.NoError(t, err)
	pair := NewPair("ukava", "hard", apy)
	require.NoError(t, pair.Validate())

	output := `Pair:
  Name: hard/ukava
	Token A: ukava
	Token B: hard
	Reward APY: 0.500000000000000000
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

			pair := NewPair(tokens[0], tokens[1], sdk.ZeroDec())
			require.NoError(t, pair.Validate())

			pairReverse := NewPair(tokens[1], tokens[0], sdk.ZeroDec())
			require.NoError(t, pairReverse.Validate())

			assert.Equal(t, tc.name, pair.Name())
			assert.Equal(t, tc.name, pairReverse.Name())
		})
	}
}

func TestPairs_Validate(t *testing.T) {
	testCases := []struct {
		name        string
		pairs       Pairs
		expectedErr string
	}{
		{
			name: "invalid pair",
			pairs: NewPairs(
				NewPair("ukava", "hard", sdk.ZeroDec()),
				NewPair("HARD", "UKAVA", sdk.ZeroDec()),
			),
			expectedErr: "invalid denom: HARD",
		},
		{
			name: "duplicate pair",
			pairs: NewPairs(
				NewPair("ukava", "hard", sdk.ZeroDec()),
				NewPair("hard", "ukava", sdk.ZeroDec()),
			),
			expectedErr: "duplicate pair: hard/ukava",
		},
		{
			name: "duplicate pairs",
			pairs: NewPairs(
				NewPair("hard", "ukava", sdk.ZeroDec()),
				NewPair("usdx", "bnb", sdk.ZeroDec()),
				NewPair("btcb", "xrpb", sdk.ZeroDec()),
				NewPair("bnb", "usdx", sdk.ZeroDec()),
			),
			expectedErr: "duplicate pair: bnb/usdx",
		},
		{
			name: "invalid apy",
			pairs: NewPairs(
				NewPair("hard", "ukava", sdk.ZeroDec()),
				NewPair("usdx", "bnb", sdk.NewDec(-1)),
				NewPair("bnb", "usdx", sdk.ZeroDec()),
			),
			expectedErr: fmt.Sprintf("invalid reward apy: %s", sdk.NewDec(-1)),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.pairs.Validate()
			assert.EqualError(t, err, tc.expectedErr)
		})
	}
}
