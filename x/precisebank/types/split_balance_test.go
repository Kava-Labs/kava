package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/kava-labs/kava/x/precisebank/types"
	"github.com/stretchr/testify/require"
)

func TestNewSplitBalance(t *testing.T) {
	tests := []struct {
		name                 string
		giveIntegerAmount    sdkmath.Int
		giveFractionalAmount sdkmath.Int
	}{
		{
			"0 amount",
			sdkmath.ZeroInt(),
			sdkmath.ZeroInt(),
		},
		{
			"1 integer amount",
			sdkmath.OneInt(),
			sdkmath.ZeroInt(),
		},
		{
			"1 fractional amount",
			sdkmath.ZeroInt(),
			sdkmath.OneInt(),
		},
		{
			"1 integer and 1 fractional amount",
			sdkmath.OneInt(),
			sdkmath.OneInt(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sbal := types.NewSplitBalance(tt.giveIntegerAmount, tt.giveFractionalAmount)

			require.Equal(t, tt.giveIntegerAmount, sbal.IntegerAmount)
			require.Equal(t, tt.giveFractionalAmount, sbal.FractionalAmount)
		})
	}
}

func TestNewSplitBalanceFromFullAmount(t *testing.T) {
	tests := []struct {
		name           string
		giveFullAmount sdkmath.Int
		wantBalances   types.SplitBalance
	}{
		{
			"0 amount",
			sdkmath.ZeroInt(),
			types.SplitBalance{
				IntegerAmount:    sdkmath.ZeroInt(),
				FractionalAmount: sdkmath.ZeroInt(),
			},
		},
		{
			"conversionFactor - 1 amount",
			types.ConversionFactor().SubRaw(1),
			types.SplitBalance{
				IntegerAmount:    sdkmath.ZeroInt(),
				FractionalAmount: types.ConversionFactor().SubRaw(1),
			},
		},
		{
			"conversionFactor amount",
			types.ConversionFactor(),
			types.SplitBalance{
				IntegerAmount:    sdkmath.OneInt(),
				FractionalAmount: sdkmath.ZeroInt(),
			},
		},
		{
			"conversionFactor + 1 amount",
			types.ConversionFactor().AddRaw(1),
			types.SplitBalance{
				IntegerAmount:    sdkmath.OneInt(),
				FractionalAmount: sdkmath.OneInt(),
			},
		},
		{
			"yuuge amount, no fractional",
			types.ConversionFactor().MulRaw(100000),
			types.SplitBalance{
				IntegerAmount:    sdkmath.NewInt(100000),
				FractionalAmount: sdkmath.OneInt(),
			},
		},
		{
			"yuuge amount, with fractional",
			types.ConversionFactor().MulRaw(100000).AddRaw(1234),
			types.SplitBalance{
				IntegerAmount:    sdkmath.NewInt(100000),
				FractionalAmount: sdkmath.NewInt(1234),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sbal := types.NewSplitBalanceFromFullAmount(tt.giveFullAmount)

			require.True(
				t,
				sbal.FractionalAmount.LT(types.ConversionFactor()),
				"fractional amount should always be less than conversion factor",
			)
			require.Equal(t, tt.wantBalances.IntegerAmount, sbal.IntegerAmount)
		})
	}
}
