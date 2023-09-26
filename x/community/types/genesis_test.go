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
}

func TestGenesisState_ValidateParams(t *testing.T) {
	for _, tc := range paramTestCases {
		t.Run(tc.name, func(t *testing.T) {
			genState := types.NewGenesisState(tc.params)

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
