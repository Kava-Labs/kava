package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/precisebank/types"
	"github.com/stretchr/testify/require"
)

func TestSumExtendedCoin(t *testing.T) {
	tests := []struct {
		name string
		amt  sdk.Coins
		want sdk.Coin
	}{
		{
			"empty",
			sdk.NewCoins(),
			sdk.NewCoin(types.ExtendedCoinDenom, sdk.ZeroInt()),
		},
		{
			"only integer",
			sdk.NewCoins(sdk.NewInt64Coin(types.IntegerCoinDenom, 100)),
			sdk.NewCoin(types.ExtendedCoinDenom, types.ConversionFactor().MulRaw(100)),
		},
		{
			"only extended",
			sdk.NewCoins(sdk.NewInt64Coin(types.ExtendedCoinDenom, 100)),
			sdk.NewCoin(types.ExtendedCoinDenom, sdk.NewInt(100)),
		},
		{
			"integer and extended",
			sdk.NewCoins(
				sdk.NewInt64Coin(types.IntegerCoinDenom, 100),
				sdk.NewInt64Coin(types.ExtendedCoinDenom, 100),
			),
			sdk.NewCoin(types.ExtendedCoinDenom, types.ConversionFactor().MulRaw(100).AddRaw(100)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extVal := types.SumExtendedCoin(tt.amt)
			require.Equal(t, tt.want, extVal)
		})
	}
}
