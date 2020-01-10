package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestSplitIntIntoWeightedBuckets(t *testing.T) {
	testCases := []struct {
		name    string
		amount  sdk.Int
		buckets []sdk.Int
		want    []sdk.Int
	}{
		{"2split1,1", i(2), is(1, 1), is(1, 1)},
		{"100split1,9", i(100), is(1, 9), is(10, 90)},
		{"7split1,2", i(7), is(1, 2), is(2, 5)},
		{"17split1,1,1", i(17), is(1, 1, 1), is(6, 6, 5)},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := splitIntIntoWeightedBuckets(tc.amount, tc.buckets)
			require.Equal(t, tc.want, got)
		})
	}
}

func i(n int64) sdk.Int { return sdk.NewInt(n) }
func is(ns ...int64) (is []sdk.Int) {
	for _, n := range ns {
		is = append(is, sdk.NewInt(n))
	}
	return
}
