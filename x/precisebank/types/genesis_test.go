package types_test

import (
	"math/rand"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/precisebank/types"
	"github.com/stretchr/testify/require"
)

func generateEqualFractionalBalances(
	t *testing.T,
	count int,
) types.FractionalBalances {
	t.Helper()

	fbs := make(types.FractionalBalances, count)
	sum := sdkmath.ZeroInt()

	// Random amounts for count - 1 FractionalBalances
	for i := 0; i < count-1; i++ {
		addr := sdk.AccAddress{byte(i)}.String()

		// Random 0 <= amt < CONVERSION_FACTOR
		amt := rand.Int63n(types.ConversionFactor().Int64())
		amtInt := sdkmath.NewInt(amt)

		fb := types.NewFractionalBalance(addr, amtInt)
		require.NoError(t, fb.Validate())

		fbs[i] = fb

		sum = sum.Add(amtInt)
	}

	// Last FractionalBalance must make sum of all balances equal to have 0
	// fractional remainder. Effectively the amount needed to round up to the
	// nearest integer amount to make this true.
	// (sum + lastAmt) % CONVERSION_FACTOR = 0
	// aka
	// CONVERSION_FACTOR - (sum % CONVERSION_FACTOR) = lastAmt
	addr := sdk.AccAddress{byte(count - 1)}.String()
	amt := types.ConversionFactor().
		Sub(sum.Mod(types.ConversionFactor()))

	fb := types.NewFractionalBalance(addr, amt)
	require.NoError(t, fb.Validate())

	fbs[count-1] = fb

	// Lets double check this before returning
	verificationSum := sdkmath.ZeroInt()
	for _, fb := range fbs {
		verificationSum = verificationSum.Add(fb.Amount)
	}
	require.True(t, verificationSum.Mod(types.ConversionFactor()).IsZero())

	// Also make sure no duplicate addresses
	require.NoError(t, fbs.Validate())

	return fbs
}

func TestGenesisStateValidate(t *testing.T) {
	testCases := []struct {
		name         string
		genesisState *types.GenesisState
		expErr       bool
	}{
		{
			"default genesisState",
			types.DefaultGenesisState(),
			false,
		},
		{
			"valid - empty balances, zero remainder",
			&types.GenesisState{
				Remainder: sdkmath.ZeroInt(),
			},
			false,
		},
		{
			"valid - nil balances",
			types.NewGenesisState(nil, sdkmath.ZeroInt()),
			false,
		},
		{
			"invalid - empty genesisState (nil remainder)",
			&types.GenesisState{},
			true,
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
			true,
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
			true,
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
			true,
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
			true,
		},
		{
			"valid - max remainder amount",
			types.NewGenesisState(
				types.FractionalBalances{
					types.NewFractionalBalance(sdk.AccAddress{1}.String(), sdkmath.NewInt(1)),
				},
				types.MaxFractionalAmount(),
			),
			false,
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
			true,
		},
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
