package app_test

import (
	"strconv"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	evmutiltypes "github.com/kava-labs/kava/x/evmutil/types"
	precisebankkeeper "github.com/kava-labs/kava/x/precisebank/keeper"
	precisebanktypes "github.com/kava-labs/kava/x/precisebank/types"
	"github.com/stretchr/testify/require"
)

func TestMigrateEvmutilToPrecisebank(t *testing.T) {
	// Full test case with all components together
	tests := []struct {
		name               string
		initialReserve     sdkmath.Int
		fractionalBalances []sdkmath.Int
	}{
		{
			"no fractional balances",
			sdkmath.NewInt(0),
			[]sdkmath.Int{},
		},
		{
			"sufficient reserve, 0 remainder",
			// Accounts adding up to 2 int units, same as reserve
			sdkmath.NewInt(2),
			[]sdkmath.Int{
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
			},
		},
		{
			"insufficient reserve, 0 remainder",
			// Accounts adding up to 2 int units, but only 1 int unit in reserve
			sdkmath.NewInt(1),
			[]sdkmath.Int{
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
			},
		},
		{
			"excess reserve, 0 remainder",
			// Accounts adding up to 2 int units, but 3 int unit in reserve
			sdkmath.NewInt(3),
			[]sdkmath.Int{
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
			},
		},
		{
			"sufficient reserve, non-zero remainder",
			// Accounts adding up to 1.5 int units, same as reserve
			sdkmath.NewInt(2),
			[]sdkmath.Int{
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
			},
		},
		{
			"insufficient reserve, non-zero remainder",
			// Accounts adding up to 1.5 int units, less than reserve,
			// Reserve should be 2 and remainder 0.5
			sdkmath.NewInt(1),
			[]sdkmath.Int{
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
			},
		},
		{
			"excess reserve, non-zero remainder",
			// Accounts adding up to 1.5 int units, 3 int units in reserve
			sdkmath.NewInt(3),
			[]sdkmath.Int{
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tApp := app.NewTestApp()
			tApp.InitializeFromGenesisStates()
			ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: time.Now()})

			ak := tApp.GetAccountKeeper()
			bk := tApp.GetBankKeeper()
			evmuk := tApp.GetEvmutilKeeper()
			pbk := tApp.GetPrecisebankKeeper()

			reserveCoin := sdk.NewCoin(precisebanktypes.IntegerCoinDenom, tt.initialReserve)
			err := bk.MintCoins(ctx, evmutiltypes.ModuleName, sdk.NewCoins(reserveCoin))
			require.NoError(t, err)

			oldReserveAddr := tApp.GetAccountKeeper().GetModuleAddress(evmutiltypes.ModuleName)
			newReserveAddr := tApp.GetAccountKeeper().GetModuleAddress(precisebanktypes.ModuleName)

			// Double check balances
			oldReserveBalance := bk.GetBalance(ctx, oldReserveAddr, precisebanktypes.IntegerCoinDenom)
			newReserveBalance := bk.GetBalance(ctx, newReserveAddr, precisebanktypes.IntegerCoinDenom)

			require.Equal(t, tt.initialReserve, oldReserveBalance.Amount, "initial x/evmutil reserve balance")
			require.True(t, newReserveBalance.IsZero(), "empty initial new reserve")

			// Set accounts
			for i, balance := range tt.fractionalBalances {
				addr := sdk.AccAddress([]byte(strconv.Itoa(i)))

				err := evmuk.SetBalance(ctx, addr, balance)
				require.NoError(t, err)
			}

			// Run full x/evmutil -> x/precisebank migration
			err = app.MigrateEvmutilToPrecisebank(
				ctx,
				ak,
				bk,
				evmuk,
				pbk,
			)
			require.NoError(t, err)

			// Check old reserve is empty
			oldReserveBalanceAfter := bk.GetBalance(ctx, oldReserveAddr, precisebanktypes.IntegerCoinDenom)
			require.True(t, oldReserveBalanceAfter.IsZero(), "old reserve should be empty")

			// Check new reserve fully backs fractional balances
			newReserveBalanceAfter := bk.GetBalance(ctx, newReserveAddr, precisebanktypes.IntegerCoinDenom)
			fractionalBalanceTotal := pbk.GetTotalSumFractionalBalances(ctx)
			remainder := pbk.GetRemainderAmount(ctx)

			expectedReserveBal := fractionalBalanceTotal.Add(remainder)
			require.Equal(
				t,
				expectedReserveBal,
				newReserveBalanceAfter.Amount.Mul(precisebanktypes.ConversionFactor()),
				"new reserve should equal total fractional balances",
			)

			// Check balances are deleted in evmutil and migrated to precisebank
			for i := range tt.fractionalBalances {
				addr := sdk.AccAddress([]byte(strconv.Itoa(i)))
				acc := evmuk.GetAccount(ctx, addr)
				require.Nil(t, acc, "account should be deleted")

				balance := pbk.GetFractionalBalance(ctx, addr)
				require.Equal(t, tt.fractionalBalances[i], balance, "balance should be migrated")
			}

			// Checks balances valid and remainder
			res, stop := precisebankkeeper.AllInvariants(pbk)(ctx)
			require.Falsef(t, stop, "invariants should pass: %s", res)
		})
	}
}

func TestTransferFractionalBalances(t *testing.T) {
	tests := []struct {
		name               string
		fractionalBalances []sdkmath.Int
	}{
		{
			"no fractional balances",
			[]sdkmath.Int{},
		},
		{
			"balanced fractional balances",
			[]sdkmath.Int{
				// 4 accounts
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
			},
		},
		{
			"unbalanced balances",
			[]sdkmath.Int{
				// 3 accounts
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tApp := app.NewTestApp()
			tApp.InitializeFromGenesisStates()
			ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: time.Now()})

			evmutilk := tApp.GetEvmutilKeeper()
			pbk := tApp.GetPrecisebankKeeper()

			for i, balance := range tt.fractionalBalances {
				addr := sdk.AccAddress([]byte(strconv.Itoa(i)))

				err := evmutilk.SetBalance(ctx, addr, balance)
				require.NoError(t, err)
			}

			// Run balance transfer
			aggregateSum, err := app.TransferFractionalBalances(
				ctx,
				evmutilk,
				pbk,
			)
			require.NoError(t, err)

			// Check balances are deleted in evmutil and migrated to precisebank
			sum := sdkmath.ZeroInt()
			for i := range tt.fractionalBalances {
				sum = sum.Add(tt.fractionalBalances[i])

				addr := sdk.AccAddress([]byte(strconv.Itoa(i)))
				acc := evmutilk.GetAccount(ctx, addr)
				require.Nil(t, acc, "account should be deleted")

				balance := pbk.GetFractionalBalance(ctx, addr)
				require.Equal(t, tt.fractionalBalances[i], balance, "balance should be migrated")
			}

			require.Equal(t, sum, aggregateSum, "aggregate sum should be correct")
		})
	}
}

func TestInitializeRemainder(t *testing.T) {
	tests := []struct {
		name             string
		giveAggregateSum sdkmath.Int
		wantRemainder    sdkmath.Int
	}{
		{
			"0 remainder, 1ukava",
			precisebanktypes.ConversionFactor(),
			sdkmath.NewInt(0),
		},
		{
			"0 remainder, multiple ukava",
			precisebanktypes.ConversionFactor().MulRaw(5),
			sdkmath.NewInt(0),
		},
		{
			"non-zero remainder, min",
			precisebanktypes.ConversionFactor().SubRaw(1),
			sdkmath.NewInt(1),
		},
		{
			"non-zero remainder, max",
			sdkmath.NewInt(1),
			precisebanktypes.ConversionFactor().SubRaw(1),
		},
		{
			"non-zero remainder, half",
			precisebanktypes.ConversionFactor().QuoRaw(2),
			precisebanktypes.ConversionFactor().QuoRaw(2),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tApp := app.NewTestApp()
			tApp.InitializeFromGenesisStates()

			pbk := tApp.GetPrecisebankKeeper()

			ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: time.Now()})

			remainder := app.InitializeRemainder(
				ctx,
				tApp.GetPrecisebankKeeper(),
				tt.giveAggregateSum,
			)
			require.Equal(t, tt.wantRemainder, remainder)

			// Check actual state
			remainderAfter := pbk.GetRemainderAmount(ctx)
			require.Equal(t, tt.wantRemainder, remainderAfter)

			// Not checking invariants here since it requires actual balance state
			aggregateSumWithRemainder := tt.giveAggregateSum.Add(remainder)
			require.True(
				t,
				aggregateSumWithRemainder.
					Mod(precisebanktypes.ConversionFactor()).
					IsZero(),
				"remainder + aggregate sum should be a multiple of the conversion factor",
			)
		})
	}
}

func TestTransferFractionalBalanceReserve(t *testing.T) {
	tests := []struct {
		name               string
		initialReserve     sdk.Coin
		fractionalBalances []sdkmath.Int
	}{
		{
			"balanced reserve, no remainder",
			sdk.NewCoin(precisebanktypes.IntegerCoinDenom, sdk.NewInt(1)),
			[]sdkmath.Int{
				// 2 accounts
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
			},
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
		},
		{
			"extra reserve funds",
			sdk.NewCoin(precisebanktypes.IntegerCoinDenom, sdk.NewInt(2)),
			[]sdkmath.Int{
				// 2 accounts, total 1 int units
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
			},
		},
		{
			"insufficient reserve, with remainder",
			sdk.NewCoin(precisebanktypes.IntegerCoinDenom, sdk.NewInt(1)),
			[]sdkmath.Int{
				// 5 accounts, total 2.5 int units
				// Expected 3 int units in reserve, 0.5 remainder
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
			},
		},
		{
			"extra reserve funds, with remainder",
			sdk.NewCoin(precisebanktypes.IntegerCoinDenom, sdk.NewInt(3)),
			[]sdkmath.Int{
				// 3 accounts, total 1.5 int units.
				// Expected 2 int units in reserve, 0.5 remainder
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
				precisebanktypes.ConversionFactor().QuoRaw(2),
			},
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
			err = app.TransferFractionalBalanceReserve(
				ctx,
				tApp.GetAccountKeeper(),
				bk,
				tApp.GetPrecisebankKeeper(),
			)
			require.NoError(t, err)

			// Check old reserve is empty
			oldReserveBalanceAfter := bk.GetBalance(ctx, oldReserveAddr, precisebanktypes.IntegerCoinDenom)
			require.True(t, oldReserveBalanceAfter.IsZero(), "old reserve should be empty")

			// Check new reserve fully backs fractional balances
			newReserveBalanceAfter := bk.GetBalance(ctx, newReserveAddr, precisebanktypes.IntegerCoinDenom)
			fractionalBalanceTotal := pbk.GetTotalSumFractionalBalances(ctx)

			expectedReserveBal := fractionalBalanceTotal.
				Quo(precisebanktypes.ConversionFactor())

			// Check if theres a remainder
			if fractionalBalanceTotal.Mod(precisebanktypes.ConversionFactor()).IsPositive() {
				expectedReserveBal = expectedReserveBal.Add(sdkmath.OneInt())
			}

			require.Equal(
				t,
				expectedReserveBal,
				newReserveBalanceAfter.Amount,
				"new reserve should equal total fractional balances + remainder",
			)
		})
	}
}
