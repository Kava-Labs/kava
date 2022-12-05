package types_test

import (
	"testing"

	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/kavamint/types"
	"github.com/stretchr/testify/require"
)

func Test_ValidateGenesisAndParams(t *testing.T) {
	testCases := []struct {
		name       string
		gs         *types.GenesisState
		shouldPass bool
	}{
		{
			"valid - default genesis is valid",
			types.DefaultGenesisState(),
			true,
		},
		{
			"valid - valid genesis",
			types.NewGenesisState(
				types.NewParams(sdk.MustNewDecFromStr("0.1"), sdk.MustNewDecFromStr("0.2")),
				time.Now(),
			),
			true,
		},
		{
			"valid - no inflation",
			types.NewGenesisState(
				types.NewParams(sdk.ZeroDec(), sdk.ZeroDec()),
				time.Now(),
			),
			true,
		},
		{
			"invalid - no time set",
			types.NewGenesisState(
				types.NewParams(sdk.MustNewDecFromStr("0.1"), sdk.MustNewDecFromStr("0.2")),
				time.Time{},
			),
			false,
		},
		{
			"invalid - community inflation param too big",
			types.NewGenesisState(
				// inflation is larger than is allowed!
				types.NewParams(types.MaxMintingRate.Add(sdk.OneDec()), sdk.ZeroDec()),
				time.Now(),
			),
			false,
		},
		{
			"invalid - staking reward inflation param too big",
			types.NewGenesisState(
				// inflation is larger than is allowed!
				types.NewParams(sdk.ZeroDec(), types.MaxMintingRate.Add(sdk.OneDec())),
				time.Now(),
			),
			false,
		},
		{
			"invalid - negative community inflation param",
			types.NewGenesisState(
				types.NewParams(sdk.OneDec().MulInt64(-1), sdk.OneDec()),
				time.Now(),
			),
			false,
		},
		{
			"invalid - negative staking inflation param",
			types.NewGenesisState(
				types.NewParams(sdk.OneDec(), sdk.OneDec().MulInt64(-1)),
				time.Now(),
			),
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.gs.Validate()
			if tc.shouldPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
