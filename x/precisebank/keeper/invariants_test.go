package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/precisebank/keeper"
	"github.com/kava-labs/kava/x/precisebank/types"
	"github.com/stretchr/testify/require"
)

func TestBalancedFractionalTotalInvariant(t *testing.T) {
	var ctx sdk.Context
	var k keeper.Keeper

	tests := []struct {
		name       string
		setupFn    func()
		wantBroken bool
		wantMsg    string
	}{
		{
			"valid - empty state",
			func() {},
			false,
			"",
		},
		{
			"valid - balances, 0 remainder",
			func() {
				k.SetFractionalBalance(ctx, sdk.AccAddress{1}, types.ConversionFactor().QuoRaw(2))
				k.SetFractionalBalance(ctx, sdk.AccAddress{2}, types.ConversionFactor().QuoRaw(2))
			},
			false,
			"",
		},
		{
			"valid - balances, non-zero remainder",
			func() {
				k.SetFractionalBalance(ctx, sdk.AccAddress{1}, types.ConversionFactor().QuoRaw(2))
				k.SetFractionalBalance(ctx, sdk.AccAddress{2}, types.ConversionFactor().QuoRaw(2).SubRaw(1))

				k.SetRemainderAmount(ctx, sdkmath.OneInt())
			},
			false,
			"",
		},
		{
			"invalid - balances, 0 remainder",
			func() {
				k.SetFractionalBalance(ctx, sdk.AccAddress{1}, types.ConversionFactor().QuoRaw(2))
				k.SetFractionalBalance(ctx, sdk.AccAddress{2}, types.ConversionFactor().QuoRaw(2).SubRaw(1))
			},
			true,
			"precisebank: invalid-fractional-total invariant\n(sum(FractionalBalances) + remainder) % conversionFactor should be 0 but got 999999999999\n",
		},
		{
			"invalid - invalid balances, non-zero (insufficient) remainder",
			func() {
				k.SetFractionalBalance(ctx, sdk.AccAddress{1}, types.ConversionFactor().QuoRaw(2))
				k.SetFractionalBalance(ctx, sdk.AccAddress{2}, types.ConversionFactor().QuoRaw(2).SubRaw(2))
				k.SetRemainderAmount(ctx, sdkmath.OneInt())
			},
			true,
			"precisebank: invalid-fractional-total invariant\n(sum(FractionalBalances) + remainder) % conversionFactor should be 0 but got 999999999999\n",
		},
		{
			"invalid - invalid balances, non-zero (excess) remainder",
			func() {
				k.SetFractionalBalance(ctx, sdk.AccAddress{1}, types.ConversionFactor().QuoRaw(2))
				k.SetFractionalBalance(ctx, sdk.AccAddress{2}, types.ConversionFactor().QuoRaw(2).SubRaw(2))
				k.SetRemainderAmount(ctx, sdkmath.NewInt(5))
			},
			true,
			"precisebank: invalid-fractional-total invariant\n(sum(FractionalBalances) + remainder) % conversionFactor should be 0 but got 3\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset each time
			tk := NewTestKeeper()
			ctx, k = tk.ctx, tk.keeper

			tt.setupFn()

			invariantFn := keeper.BalancedFractionalTotalInvariant(k)
			msg, broken := invariantFn(ctx)

			if tt.wantBroken {
				require.True(t, broken, "invariant should be broken but is not")
				require.Equal(t, tt.wantMsg, msg)
			} else {
				require.False(t, broken, "invariant should not be broken but is")
			}
		})
	}
}
