package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenesisState_Validate(t *testing.T) {
	testCases := []struct {
		name       string
		genState   GenesisState
		expectPass bool
	}{
		{
			name:       "normal",
			genState:   DefaultGenesisState(),
			expectPass: true,
		},
		// TODO test failure cases
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			err := tc.genState.Validate()

			if tc.expectPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}

}
