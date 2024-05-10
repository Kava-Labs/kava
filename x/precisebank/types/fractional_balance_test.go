package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
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
			"valid - uppercase address",
			"KAVA1GPXD677PP8ZR97XVY3PMGK70A9VCPAGSAKV0TX",
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
			types.GetMaxFractionalAmount(),
			"",
		},
		{
			"invalid - 0 balance",
			"kava1gpxd677pp8zr97xvy3pmgk70a9vcpagsakv0tx",
			sdkmath.NewInt(0),
			"non-positive amount 0",
		},
		{
			"invalid - empty",
			"kava1gpxd677pp8zr97xvy3pmgk70a9vcpagsakv0tx",
			sdkmath.Int{},
			"nil amount",
		},
		{
			"invalid - mixed case address",
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
			"non-positive amount -100",
		},
		{
			"invalid - max amount + 1",
			"kava1gpxd677pp8zr97xvy3pmgk70a9vcpagsakv0tx",
			types.GetMaxFractionalAmount().AddRaw(1),
			"amount 1000000000000 exceeds max of 999999999999",
		},
		{
			"invalid - much more than max amount",
			"kava1gpxd677pp8zr97xvy3pmgk70a9vcpagsakv0tx",
			sdkmath.NewInt(100000000000_000),
			"amount 100000000000000 exceeds max of 999999999999",
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
