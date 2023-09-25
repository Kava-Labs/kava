package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	"github.com/kava-labs/kava/x/community/types"
)

type paramTestCase struct {
	name        string
	params      types.Params
	expectedErr string
}

var paramTestCases = []paramTestCase{
	{
		name:        "default params are valid",
		params:      types.DefaultParams(),
		expectedErr: "",
	},
	{
		name: "valid params",
		params: types.Params{
			UpgradeTimeDisableInflation: time.Time{},
			StakingRewardsPerSecond:     sdkmath.LegacyNewDec(1000),
		},
		expectedErr: "",
	},
	{
		name: "rewards per second are allowed to be zero",
		params: types.Params{
			UpgradeTimeDisableInflation: time.Time{},
			StakingRewardsPerSecond:     sdkmath.LegacyNewDec(0),
		},
		expectedErr: "",
	},
	{
		name: "nil rewards per second",
		params: types.Params{
			UpgradeTimeDisableInflation: time.Time{},
			StakingRewardsPerSecond:     sdkmath.LegacyDec{},
		},
		expectedErr: "StakingRewardsPerSecond should not be nil",
	},
	{
		name: "negative rewards per second",
		params: types.Params{
			UpgradeTimeDisableInflation: time.Time{},
			StakingRewardsPerSecond:     sdkmath.LegacyNewDec(-5),
		},
		expectedErr: "StakingRewardsPerSecond should not be negative",
	},
}

func TestParamsValidate(t *testing.T) {
	for _, tc := range paramTestCases {
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
