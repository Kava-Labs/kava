package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestRawPriceKey_Iteration(t *testing.T) {
	// An iterator key should only match price keys with the same market
	iteratorKey := RawPriceIteratorKey("kava:usd")

	addr, err := sdk.AccAddressFromBech32("kava17e8afwcxk0k72hhr7xapugtf5xhxth0a6d5jwz")
	require.NoError(t, err)

	testCases := []struct {
		name      string
		priceKey  []byte
		expectErr bool
	}{
		{
			name:      "equal marketID is included in iteration",
			priceKey:  RawPriceKey("kava:usd", addr),
			expectErr: false,
		},
		{
			name:      "prefix overlapping marketID excluded from iteration",
			priceKey:  RawPriceKey("kava:usd:30", addr),
			expectErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			matchedSubKey := tc.priceKey[:len(iteratorKey)]
			if tc.expectErr {
				require.NotEqual(t, iteratorKey, matchedSubKey)
			} else {
				require.Equal(t, iteratorKey, matchedSubKey)
			}
		})
	}
}
