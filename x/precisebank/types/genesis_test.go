package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/precisebank/types"
	"github.com/stretchr/testify/require"
)

func TestGenesisStateValidate_Basic(t *testing.T) {
	testCases := []struct {
		name         string
		genesisState *types.GenesisState
		wantErr      string
	}{
		{
			"valid - default genesisState",
			types.DefaultGenesisState(),
			"",
		},
		{
			"valid - empty balances, zero remainder",
			&types.GenesisState{
				Remainder: sdkmath.ZeroInt(),
			},
			"",
		},
		{
			"valid - nil balances",
			types.NewGenesisState(nil, sdkmath.ZeroInt()),
			"",
		},
		{
			"valid - max remainder amount",
			types.NewGenesisState(
				types.FractionalBalances{
					types.NewFractionalBalance(sdk.AccAddress{1}.String(), sdkmath.NewInt(1)),
				},
				types.MaxFractionalAmount(),
			),
			"",
		},
		{
			"invalid - empty genesisState (nil remainder)",
			&types.GenesisState{},
			"nil remainder amount",
		},
		{
			"valid - balances add up",
			types.NewGenesisState(
				types.FractionalBalances{
					types.NewFractionalBalance(sdk.AccAddress{1}.String(), sdkmath.NewInt(1)),
					types.NewFractionalBalance(sdk.AccAddress{1}.String(), sdkmath.NewInt(1)),
				},
				sdkmath.ZeroInt(),
			),
			"invalid balances: duplicate address cosmos1qyfkm2y3",
		},
		{
			"invalid - calls (single) FractionalBalance.Validate()",
			types.NewGenesisState(
				types.FractionalBalances{
					types.NewFractionalBalance(sdk.AccAddress{1}.String(), sdkmath.NewInt(1)),
					types.NewFractionalBalance(sdk.AccAddress{2}.String(), sdkmath.NewInt(-1)),
				},
				sdkmath.ZeroInt(),
			),
			"invalid balances: invalid fractional balance for cosmos1qgcgaq4k: non-positive amount -1",
		},
		{
			"invalid - calls (slice) FractionalBalances.Validate()",
			types.NewGenesisState(
				types.FractionalBalances{
					types.NewFractionalBalance(sdk.AccAddress{1}.String(), sdkmath.NewInt(1)),
					types.NewFractionalBalance(sdk.AccAddress{1}.String(), sdkmath.NewInt(1)),
				},
				sdkmath.ZeroInt(),
			),
			"invalid balances: duplicate address cosmos1qyfkm2y3",
		},
		{
			"invalid - negative remainder",
			types.NewGenesisState(
				types.FractionalBalances{
					types.NewFractionalBalance(sdk.AccAddress{1}.String(), sdkmath.NewInt(1)),
					types.NewFractionalBalance(sdk.AccAddress{2}.String(), sdkmath.NewInt(1)),
				},
				sdkmath.NewInt(-1),
			),
			"negative remainder amount -1",
		},
		{
			"invalid - too large remainder",
			types.NewGenesisState(
				types.FractionalBalances{
					types.NewFractionalBalance(sdk.AccAddress{1}.String(), sdkmath.NewInt(1)),
					types.NewFractionalBalance(sdk.AccAddress{2}.String(), sdkmath.NewInt(1)),
				},
				types.MaxFractionalAmount().AddRaw(1),
			),
			"remainder 1000000000000 exceeds max of 999999999999",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			err := tc.genesisState.Validate()

			if tc.wantErr == "" {
				require.NoError(tt, err)
			} else {
				require.Error(tt, err)
				require.EqualError(tt, err, tc.wantErr)
			}
		})
	}
}

func TestGenesisStateValidate_Total(t *testing.T) {
	testCases := []struct {
		name           string
		genesisStateFn func() *types.GenesisState
		wantErr        string
	}{
		{
			"valid - empty balances, zero remainder",
			func() *types.GenesisState {
				return types.NewGenesisState(nil, sdkmath.ZeroInt())
			},
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			err := tc.genesisStateFn().Validate()

			if tc.wantErr == "" {
				require.NoError(tt, err)
			} else {
				require.Error(tt, err)
				require.EqualError(tt, err, tc.wantErr)
			}
		})
	}
}
