package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
)

func TestSplitIntIntoWeightedBuckets(t *testing.T) {
	testCases := []struct {
		name        string
		amount      sdkmath.Int
		buckets     []sdkmath.Int
		want        []sdkmath.Int
		expectPanic bool
	}{
		{
			name:        "0split0",
			amount:      i(0),
			buckets:     is(0),
			expectPanic: true,
		},
		{
			name:        "5splitnil",
			amount:      i(5),
			buckets:     is(),
			expectPanic: true,
		},
		{
			name:        "-2split1,1",
			amount:      i(-2),
			buckets:     is(1, 1),
			expectPanic: true,
		},
		{
			name:        "2split1,-1",
			amount:      i(2),
			buckets:     is(1, -1),
			expectPanic: true,
		},
		{
			name:    "0split0,0,0,1",
			amount:  i(0),
			buckets: is(0, 0, 0, 1),
			want:    is(0, 0, 0, 0),
		},
		{
			name:    "2split1,1",
			amount:  i(2),
			buckets: is(1, 1),
			want:    is(1, 1),
		},
		{
			name:    "100split1,9",
			amount:  i(100),
			buckets: is(1, 9),
			want:    is(10, 90),
		},
		{
			name:    "100split9,1",
			amount:  i(100),
			buckets: is(9, 1),
			want:    is(90, 10),
		},
		{
			name:    "7split1,2",
			amount:  i(7),
			buckets: is(1, 2),
			want:    is(2, 5),
		},
		{
			name:    "17split1,1,1",
			amount:  i(17),
			buckets: is(1, 1, 1),
			want:    is(6, 6, 5),
		},
		{
			name:    "10split1000000,1",
			amount:  i(10),
			buckets: is(1000000, 1),
			want:    is(10, 0),
		},
		{
			name:    "334733353split730777,31547",
			amount:  i(334733353),
			buckets: is(730777, 31547),
			want:    is(320881194, 13852159),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var got []sdkmath.Int
			run := func() {
				got = splitIntIntoWeightedBuckets(tc.amount, tc.buckets)
			}
			if tc.expectPanic {
				require.Panics(t, run)
			} else {
				require.NotPanics(t, run)
			}

			require.Equal(t, tc.want, got)
		})
	}
}

func i(n int64) sdkmath.Int { return sdkmath.NewInt(n) }
func is(ns ...int64) (is []sdkmath.Int) {
	for _, n := range ns {
		is = append(is, sdkmath.NewInt(n))
	}
	return
}
