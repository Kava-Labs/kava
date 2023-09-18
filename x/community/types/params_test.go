package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	"github.com/kava-labs/kava/x/community/types"
)

func TestParamsValidate(t *testing.T) {
	testCases := []struct {
		name        string
		params      types.Params
		expectedErr string
	}{
		{
			name: "valid parms",
			params: types.Params{
				UpgradeTimeDisableInflation: time.Time{},
				StakingRewardsPerSecond:     sdkmath.NewInt(1000),
			},
			expectedErr: "",
		},
		{
			name: "nil rewards per second",
			params: types.Params{
				UpgradeTimeDisableInflation: time.Time{},
				StakingRewardsPerSecond:     sdkmath.Int{},
			},
			expectedErr: "StakingRewardsPerSecond should not be nil",
		},
		{
			name: "negative rewards per second",
			params: types.Params{
				UpgradeTimeDisableInflation: time.Time{},
				StakingRewardsPerSecond:     sdkmath.NewInt(-5),
			},
			expectedErr: "StakingRewardsPerSecond should not be negative",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.params.Validate()

			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr)
			}
		})
	}
}
