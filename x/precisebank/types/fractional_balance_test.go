package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/precisebank/types"
	"github.com/stretchr/testify/require"
)

func TestNewFractionalBalance(t *testing.T) {
	tests := []struct {
		name        string
		giveAddress string
		giveAmount  sdkmath.Int
	}{
		{
			"correctly sets fields",
			"cosmos1qperwt9wrnkg5k9e5gzfgjppzpqur82k6c5a0n",
			sdkmath.NewInt(100),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fb := types.NewFractionalBalance(tt.giveAddress, tt.giveAmount)

			require.Equal(t, tt.giveAddress, fb.Address)
			require.Equal(t, tt.giveAmount, fb.Amount)
		})
	}
}

func TestFractionalBalance_Validate(t *testing.T) {
	app.SetSDKConfig()

	tests := []struct {
		name        string
		giveAddress string
		giveAmount  sdkmath.Int
		wantErr     string
	}{
		{
			"valid",
			"kava1gpxd677pp8zr97xvy3pmgk70a9vcpagsakv0tx",
			sdkmath.NewInt(100),
			"",
		},
		{
			"valid - min balance",
			"kava1gpxd677pp8zr97xvy3pmgk70a9vcpagsakv0tx",
			sdkmath.NewInt(1),
			"",
		},
		{
			"valid - max balance",
			"kava1gpxd677pp8zr97xvy3pmgk70a9vcpagsakv0tx",
			types.MAX_FRACTIONAL_AMOUNT,
			"",
		},
		{
			"invalid - uppercase letter",
			"kava1gpxd677pP8zr97xvy3pmgk70a9vcpagsakv0tx",
			sdkmath.NewInt(100),
			"decoding bech32 failed: string not all lowercase or all uppercase",
		},
		{
			"invalid - non-bech32 address",
			"invalid",
			sdkmath.NewInt(100),
			"decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			"invalid - wrong bech32 prefix",
			"cosmos1qperwt9wrnkg5k9e5gzfgjppzpqur82k7gqd8n",
			sdkmath.NewInt(100),
			"invalid Bech32 prefix; expected kava, got cosmos",
		},
		{
			"invalid - negative amount",
			"kava1gpxd677pp8zr97xvy3pmgk70a9vcpagsakv0tx",
			sdkmath.NewInt(-100),
			"non-positive amount: -100",
		},
		{
			"invalid - max amount + 1",
			"kava1gpxd677pp8zr97xvy3pmgk70a9vcpagsakv0tx",
			types.MAX_FRACTIONAL_AMOUNT.AddRaw(1),
			"amount exceeds max of 999999999999: 1000000000000",
		},
		{
			"invalid - much more than max amount",
			"kava1gpxd677pp8zr97xvy3pmgk70a9vcpagsakv0tx",
			sdkmath.NewInt(100000000000_000),
			"amount exceeds max of 999999999999: 100000000000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fb := types.NewFractionalBalance(tt.giveAddress, tt.giveAmount)
			err := fb.Validate()

			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			require.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestFractionalBalance_GetAddress(t *testing.T) {
	tests := []struct {
		name        string
		giveAddress string
		giveAmount  sdkmath.Int
		wantAddr    sdk.AccAddress
	}{
		{
			"correctly returns address",
			sdk.AccAddress{1}.String(),
			sdkmath.NewInt(100),
			sdk.AccAddress{1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fb := types.NewFractionalBalance(tt.giveAddress, tt.giveAmount)

			addr := fb.GetAddress()
			require.Equal(t, tt.wantAddr, addr)
		})
	}
}
