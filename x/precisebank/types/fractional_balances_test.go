package types_test

import (
	"math/big"
	"math/rand"
	"strings"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/precisebank/types"
	"github.com/stretchr/testify/require"
)

func TestFractionalBalances_Validate(t *testing.T) {
	app.SetSDKConfig()

	tests := []struct {
		name    string
		fbs     types.FractionalBalances
		wantErr string
	}{
		{
			"valid - empty",
			types.FractionalBalances{},
			"",
		},
		{
			"valid - nil",
			nil,
			"",
		},
		{
			"valid - multiple balances",
			types.FractionalBalances{
				types.NewFractionalBalance(sdk.AccAddress{1}.String(), sdkmath.NewInt(100)),
				types.NewFractionalBalance(sdk.AccAddress{2}.String(), sdkmath.NewInt(100)),
				types.NewFractionalBalance(sdk.AccAddress{3}.String(), sdkmath.NewInt(100)),
			},
			"",
		},
		{
			"invalid - single invalid balance",
			types.FractionalBalances{
				types.NewFractionalBalance(sdk.AccAddress{1}.String(), sdkmath.NewInt(100)),
				types.NewFractionalBalance(sdk.AccAddress{2}.String(), sdkmath.NewInt(-1)),
				types.NewFractionalBalance(sdk.AccAddress{3}.String(), sdkmath.NewInt(100)),
			},
			"invalid fractional balance for kava1qg7c45n6: non-positive amount -1",
		},
		{
			"invalid - duplicate address",
			types.FractionalBalances{
				types.NewFractionalBalance(sdk.AccAddress{1}.String(), sdkmath.NewInt(100)),
				types.NewFractionalBalance(sdk.AccAddress{1}.String(), sdkmath.NewInt(100)),
			},
			"duplicate address kava1qy0xn7za",
		},
		{
			"invalid - duplicate address upper/lower case",
			types.FractionalBalances{
				types.NewFractionalBalance(
					strings.ToLower(sdk.AccAddress{1}.String()),
					sdkmath.NewInt(100),
				),
				types.NewFractionalBalance(
					strings.ToUpper(sdk.AccAddress{1}.String()),
					sdkmath.NewInt(100),
				),
			},
			"duplicate address kava1qy0xn7za",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fbs.Validate()
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			require.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestFractionalBalances_SumAmount(t *testing.T) {
	generateRandomFractionalBalances := func(n int) (types.FractionalBalances, sdkmath.Int) {
		balances := make(types.FractionalBalances, n)
		sum := sdkmath.ZeroInt()

		for i := 0; i < n; i++ {
			addr := sdk.AccAddress{byte(i)}.String()
			amount := sdkmath.NewInt(rand.Int63())
			balances[i] = types.NewFractionalBalance(addr, amount)

			sum = sum.Add(amount)
		}

		return balances, sum
	}

	multiBalances, sum := generateRandomFractionalBalances(10)

	tests := []struct {
		name     string
		balances types.FractionalBalances
		wantSum  sdkmath.Int
	}{
		{
			"empty",
			types.FractionalBalances{},
			sdkmath.ZeroInt(),
		},
		{
			"single",
			types.FractionalBalances{
				types.NewFractionalBalance(sdk.AccAddress{1}.String(), sdkmath.NewInt(100)),
			},
			sdkmath.NewInt(100),
		},
		{
			"multiple",
			multiBalances,
			sum,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sum := tt.balances.SumAmount()
			require.Equal(t, tt.wantSum, sum)
		})
	}
}

func TestFractionalBalances_SumAmount_Overflow(t *testing.T) {
	// 2^256 - 1
	maxInt := new(big.Int).Sub(
		new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil),
		big.NewInt(1),
	)

	fbs := types.FractionalBalances{
		types.NewFractionalBalance(sdk.AccAddress{1}.String(), sdkmath.NewInt(100)),
		// This is NOT valid, but just to test overflows will panic
		types.NewFractionalBalance(
			sdk.AccAddress{2}.String(),
			sdkmath.NewIntFromBigInt(maxInt),
		),
	}

	require.PanicsWithError(t, sdkmath.ErrIntOverflow.Error(), func() {
		_ = fbs.SumAmount()
	})
}
