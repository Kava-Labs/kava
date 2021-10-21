package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetDiff(t *testing.T) {
	tests := []struct {
		name     string
		setA     []string
		setB     []string
		expected []string
	}{
		{"empty", []string{}, []string{}, []string(nil)},
		{"diff equal sets", []string{"busd", "usdx"}, []string{"busd", "usdx"}, []string(nil)},
		{"diff set empty", []string{"bnb", "ukava", "usdx"}, []string{}, []string{"bnb", "ukava", "usdx"}},
		{"input set empty", []string{}, []string{"bnb", "ukava", "usdx"}, []string(nil)},
		{"diff set with common elements", []string{"bnb", "btcb", "usdx", "xrpb"}, []string{"bnb", "usdx"}, []string{"btcb", "xrpb"}},
		{"diff set with all common elements", []string{"bnb", "usdx"}, []string{"bnb", "btcb", "usdx", "xrpb"}, []string(nil)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, setDifference(tt.setA, tt.setB))
		})
	}
}
