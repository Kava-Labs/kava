package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/precisebank/types"
	"github.com/stretchr/testify/require"
)

func TestKeeper_GetBalance(t *testing.T) {
	tests := []struct {
		name      string
		giveDenom string // queried denom for balance

		giveBankBal       sdk.Coins   // mocked bank balance for giveAddr
		giveFractionalBal sdkmath.Int // stored fractional balance for giveAddr

		wantBal sdk.Coin
	}{
		{
			"extended denom - no fractional balance",
			types.ExtendedCoinDenom,
			// queried bank balance in ukava when querying for akava
			sdk.NewCoins(sdk.NewCoin(types.IntegerCoinDenom, sdk.NewInt(1000))),
			sdkmath.ZeroInt(),
			// integer + fractional
			sdk.NewCoin(types.ExtendedCoinDenom, sdk.NewInt(1000_000_000_000_000)),
		},
		{
			"extended denom - with fractional balance",
			types.ExtendedCoinDenom,
			sdk.NewCoins(sdk.NewCoin(types.IntegerCoinDenom, sdk.NewInt(1000))),
			sdkmath.NewInt(100),
			// integer + fractional
			sdk.NewCoin(types.ExtendedCoinDenom, sdk.NewInt(1000_000_000_000_100)),
		},
		{
			"extended denom - only fractional balance",
			types.ExtendedCoinDenom,
			// no coins in bank, only fractional balance
			sdk.NewCoins(),
			sdkmath.NewInt(100),
			sdk.NewCoin(types.ExtendedCoinDenom, sdk.NewInt(100)),
		},
		{
			"extended denom - max fractional balance",
			types.ExtendedCoinDenom,
			sdk.NewCoins(sdk.NewCoin(types.IntegerCoinDenom, sdk.NewInt(1000))),
			types.ConversionFactor().SubRaw(1),
			// integer + fractional
			sdk.NewCoin(types.ExtendedCoinDenom, sdk.NewInt(1000_999_999_999_999)),
		},
		{
			"non-extended denom - ukava returns ukava",
			types.IntegerCoinDenom,
			sdk.NewCoins(sdk.NewCoin(types.IntegerCoinDenom, sdk.NewInt(1000))),
			sdk.ZeroInt(),
			sdk.NewCoin("ukava", sdk.NewInt(1000)),
		},
		{
			"non-extended denom - unaffected by fractional balance",
			"ukava",
			sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000))),
			sdkmath.NewInt(100),
			sdk.NewCoin("ukava", sdk.NewInt(1000)),
		},
		{
			"unrelated denom - no fractional",
			"busd",
			sdk.NewCoins(sdk.NewCoin("busd", sdk.NewInt(1000))),
			sdk.ZeroInt(),
			sdk.NewCoin("busd", sdk.NewInt(1000)),
		},
		{
			"unrelated denom - unaffected by fractional balance",
			"busd",
			sdk.NewCoins(sdk.NewCoin("busd", sdk.NewInt(1000))),
			sdkmath.NewInt(100),
			sdk.NewCoin("busd", sdk.NewInt(1000)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tk := NewMockedTestData(t)
			addr := sdk.AccAddress([]byte("test-address"))

			// Set fractional balance in store before query
			tk.keeper.SetFractionalBalance(tk.ctx, addr, tt.giveFractionalBal)

			if tt.giveDenom == types.ExtendedCoinDenom {
				// No balance pass through
				tk.bk.EXPECT().
					SpendableCoins(tk.ctx, addr).
					RunAndReturn(func(_ sdk.Context, _ sdk.AccAddress) sdk.Coins {
						return tt.giveBankBal
					}).
					Once()
			} else {
				// Pass through to x/bank for denoms except ExtendedCoinDenom
				tk.bk.EXPECT().
					GetBalance(tk.ctx, addr, tt.giveDenom).
					RunAndReturn(func(ctx sdk.Context, aa sdk.AccAddress, s string) sdk.Coin {
						require.Equal(t, s, tt.giveDenom, "unexpected denom passed to x/bank.GetBalance")

						return sdk.NewCoin(tt.giveDenom, tt.giveBankBal.AmountOf(s))
					}).
					Once()
			}

			bal := tk.keeper.GetBalance(tk.ctx, addr, tt.giveDenom)
			require.Equal(t, tt.wantBal, bal)
		})
	}
}
