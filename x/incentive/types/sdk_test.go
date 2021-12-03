package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/kava-labs/kava/x/incentive/types"
)

func TestGetTotalVestingPeriodLength(t *testing.T) {
	testCases := []struct {
		name        string
		periods     vestingtypes.Periods
		expectedVal int64
	}{
		{
			name: "two period lengths are added together",
			periods: vestingtypes.Periods{
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
