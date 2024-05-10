package types_test

import (
	"testing"

	"github.com/kava-labs/kava/x/precisebank/types"
	"github.com/stretchr/testify/require"
)

func TestGenesisStateValidate(t *testing.T) {
	testCases := []struct {
		name         string
		genesisState types.GenesisState
		expErr       bool
	}{
		{
			"empty genesisState",
			types.GenesisState{},
			false,
		},
		{
			"valid genesisState",
			// TODO: Fill out fields
			types.GenesisState{},
			false,
		},
		// TODO: Sum of balances does not equal an integer amount
		// {
		// 	"invalid balances",
		// 	types.GenesisState{},
		// 	true,
		// },
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			err := tc.genesisState.Validate()

			if tc.expErr {
				require.Error(tt, err)
			} else {
				require.NoError(tt, err)
			}
		})
	}
}
