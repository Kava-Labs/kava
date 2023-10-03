package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kava-labs/kava/x/community/types"
)

func TestDefaultGenesisState(t *testing.T) {
	defaultGen := types.DefaultGenesisState()

	require.NoError(t, defaultGen.Validate())
	require.Equal(t, types.DefaultParams(), defaultGen.Params)
	require.Equal(t, types.DefaultStakingRewardsState(), defaultGen.StakingRewardsState)
}

func TestGenesisState_ValidateParams(t *testing.T) {
	for _, tc := range paramTestCases {
		t.Run(tc.name, func(t *testing.T) {
			genState := types.DefaultGenesisState()
			genState.Params = tc.params

			err := genState.Validate()

			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr)
			}
		})
	}
}

func TestGenesisState_ValidateStakingRewardsState(t *testing.T) {
	for _, tc := range stakingRewardsStateTestCases {
		t.Run(tc.name, func(t *testing.T) {
			genState := types.DefaultGenesisState()
			genState.StakingRewardsState = tc.stakingRewardsState

			err := genState.Validate()

			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr)
			}
		})
	}
}
