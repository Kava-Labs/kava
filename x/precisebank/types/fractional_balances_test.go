package types_test

import (
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
