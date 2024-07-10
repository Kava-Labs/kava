package app_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	evmutiltypes "github.com/kava-labs/kava/x/evmutil/types"
	precisebanktypes "github.com/kava-labs/kava/x/precisebank/types"
	"github.com/stretchr/testify/require"
)

func TestReserveMigration(t *testing.T) {
	tests := []struct {
		name               string
		initialReserve     sdk.Coin
		fractionalBalances []sdkmath.Int
		wantErr            string
	}{
		{
			"error - zero fractional balances",
			sdk.NewCoin(precisebanktypes.IntegerCoinDenom, sdk.NewInt(10)),
			[]sdkmath.Int{
				// No accounts
			},
			"invalid state, total fractional balances should not be zero",
		},
		{
			"error - unbalanced fractional balances",
			sdk.NewCoin(precisebanktypes.IntegerCoinDenom, sdk.NewInt(1)),
			[]sdkmath.Int{
				// 1 account, only 0.5 int units
				precisebanktypes.ConversionFactor().QuoRaw(2),
			},
			"invalid state, total fractional balances should be a multiple of the conversion factor but is 500000000000",
		},
		{
			"balanced reserve",
			sdk.NewCoin(precisebanktypes.IntegerCoinDenom, sdk.NewInt(1)),
			[]sdkmath.Int{
				// 2 accounts
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
			},
			"",
		},
		{
			"insufficient reserve",
			sdk.NewCoin(precisebanktypes.IntegerCoinDenom, sdk.NewInt(1)),
			[]sdkmath.Int{
				// 4 accounts, total 2 int units
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
			},
			"",
		},
		{
			"extra reserve funds",
			sdk.NewCoin(precisebanktypes.IntegerCoinDenom, sdk.NewInt(2)),
			[]sdkmath.Int{
				// 2 accounts, total 1 int units
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
			},
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tApp := app.NewTestApp()
			tApp.InitializeFromGenesisStates()
			ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: time.Now()})

			bk := tApp.GetBankKeeper()
			pbk := tApp.GetPrecisebankKeeper()
			err := bk.MintCoins(ctx, evmutiltypes.ModuleName, sdk.NewCoins(tt.initialReserve))
			require.NoError(t, err)

			oldReserveAddr := tApp.GetAccountKeeper().GetModuleAddress(evmutiltypes.ModuleName)
			newReserveAddr := tApp.GetAccountKeeper().GetModuleAddress(precisebanktypes.ModuleName)

			// Double check balances
			oldReserveBalance := bk.GetBalance(ctx, oldReserveAddr, precisebanktypes.IntegerCoinDenom)
			newReserveBalance := bk.GetBalance(ctx, newReserveAddr, precisebanktypes.IntegerCoinDenom)

			require.Equal(t, tt.initialReserve, oldReserveBalance)
			require.True(t, newReserveBalance.IsZero(), "empty initial new reserve")

			for i, balance := range tt.fractionalBalances {
				addr := sdk.AccAddress([]byte{byte(i)})

				require.NotPanics(t, func() {
					pbk.SetFractionalBalance(ctx, addr, balance)
				}, "given fractional balances should be valid")
			}

			// Run reserve migration
			err = app.MigrateFractionalBalanceReserve(
				ctx,
				tApp.Logger(),
				tApp.GetAccountKeeper(),
				bk,
				tApp.GetPrecisebankKeeper(),
			)

			// Expect error, no additional balance checks
			if tt.wantErr != "" {
				require.Error(t, err)
				require.Equal(t, tt.wantErr, err.Error())
				return
			}

			require.NoError(t, err)

			// Check old reserve is empty
			oldReserveBalanceAfter := bk.GetBalance(ctx, oldReserveAddr, precisebanktypes.IntegerCoinDenom)
			require.True(t, oldReserveBalanceAfter.IsZero(), "old reserve should be empty")

			// Check new reserve fully backs fractional balances
			newReserveBalanceAfter := bk.GetBalance(ctx, newReserveAddr, precisebanktypes.IntegerCoinDenom)
			fractionalBalanceTotal := pbk.GetTotalSumFractionalBalances(ctx)
			require.Equal(
				t,
				fractionalBalanceTotal,
				newReserveBalanceAfter.Amount.Mul(precisebanktypes.ConversionFactor()),
				"new reserve should equal total fractional balances",
			)
		})
	}
}
