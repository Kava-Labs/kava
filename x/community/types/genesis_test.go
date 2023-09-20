package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	"github.com/kava-labs/kava/x/community/types"
)

func TestDefaultGenesisState(t *testing.T) {
	defaultGen := types.DefaultGenesisState()

	require.NoError(t, defaultGen.Validate())
	require.Equal(t, types.DefaultParams(), defaultGen.Params)
}

func TestGenesisState_Validate(t *testing.T) {
	testCases := []struct {
		name        string
		genesis     types.GenesisState
		expectedErr string
	}{
		{
			name: "valid state",
			genesis: types.GenesisState{
				Params: types.Params{
					UpgradeTimeDisableInflation: time.Time{},
					StakingRewardsPerSecond:     sdkmath.NewInt(1000),
				},
			},
			expectedErr: "",
		},
		{
			name: "invalid params",
			genesis: types.GenesisState{
				Params: types.Params{
					UpgradeTimeDisableInflation: time.Time{},
					StakingRewardsPerSecond:     sdkmath.Int{},
				},
			},
			expectedErr: "StakingRewardsPerSecond should not be nil",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.genesis.Validate()

			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr)
			}
		})
	}
}
