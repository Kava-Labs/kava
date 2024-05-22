package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/precisebank/types"
	"github.com/stretchr/testify/require"
)

func TestSendCoins(t *testing.T) {
	sender := sdk.AccAddress([]byte{1})
	recipient := sdk.AccAddress([]byte{2})

	tests := []struct {
		name string
		// Must not contain ExtendedCoinDenom - or only after Mint is implemented
		// to support fractional balances and that is used of native x/bank mint.
		giveStartBalSender    sdk.Coins
		giveStartBalRecipient sdk.Coins
		giveAmt               sdk.Coins
		wantErr               string
	}{
		{
			"passthrough - ukava",
			cs(c(types.IntegerCoinDenom, 1000)),
			cs(c(types.IntegerCoinDenom, 1000)),
			cs(c(types.IntegerCoinDenom, 1000)),
			"",
		},
		{
			"passthrough - busd",
			cs(c("busd", 1000)),
			cs(),
			cs(c("busd", 1000)),
			"",
		},
		{
			"akava send - 1akava to 0 balance",
			// Starting balances
			cs(c(types.IntegerCoinDenom, 1000)),
			cs(),
			// Send amount
			cs(c(types.ExtendedCoinDenom, 1)), // akava
			"",
		},
		{
			"sender borrow from integer",
			// 1ukava, 0 fractional
			cs(ci(types.IntegerCoinDenom, types.ConversionFactor())),
			cs(),
			// Send 1 with 0 fractional
			cs(c(types.ExtendedCoinDenom, 1)),
			"",
		},
		{
			"receiver carry",
			cs(c(types.IntegerCoinDenom, 1000)),
			// max fractional amount, carries over to integer
			cs(ci(types.IntegerCoinDenom, types.ConversionFactor().SubRaw(1))),
			cs(c(types.ExtendedCoinDenom, 1)),
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			td := NewMockedTestData(t)

			// -------------------------------
			// Passthrough SendCoins x/bank expectations

			passthroughCoins := tt.giveAmt
			extendedAmount := tt.giveAmt.AmountOf(types.ExtendedCoinDenom)
			if extendedAmount.IsPositive() {
				removeCoin := sdk.NewCoin(types.ExtendedCoinDenom, extendedAmount)
				passthroughCoins = tt.giveAmt.Sub(removeCoin)
			}

			// Direct x/bank send if there are some unmanaged coins, e.g. not akava
			if !passthroughCoins.IsZero() {
				td.bk.EXPECT().
					SendCoins(td.ctx, sender, recipient, passthroughCoins).
					Return(nil).
					Once()
			}

			// -------------------------------------------
			// x/precisebank reserve exchange expectations
			// If account has integer balance but insufficient fractional
			// balance an integer borrow will be made from the integer balance
			// by transferring 1 integer unit to x/precisebank module account
			// in exchange for equivalent fractional units.
			senderIntBal, senderFracBal := splitExtendedAmount(tt.giveStartBalSender.AmountOf(types.ExtendedCoinDenom))
			_, sendFracAmt := splitExtendedAmount(tt.giveAmt.AmountOf(types.ExtendedCoinDenom))

			if senderFracBal.LT(sendFracAmt) {
				// Insufficient fractional balance, borrow from integer balance.
				// But only if there is an integer balance to borrow from.
				if !senderIntBal.IsZero() {
					td.bk.EXPECT().
						SendCoins(td.ctx, sender, types.ModuleName, cs(c(types.IntegerCoinDenom, 1))).
						Return(nil).
						Once()
				}
			}

			// -------------------------------------------
			// Do the thing
			err := td.keeper.SendCoins(td.ctx, sender, recipient, tt.giveAmt)

			if tt.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)

			// -----------------------------------------------------------------
			// Assertions
			// We only care about extended balance here as x/bank is external.
			// bankkeeper.SendCoins() is already expected to be called with
			// the correct parameters above, and we don't need to verify that
			// x/bank state is correctly set.

			// sender = startBal - amt
			_, expectedFracBalSender := splitExtendedAmount(tt.giveStartBalSender.
				Sub(tt.giveAmt...).
				AmountOf(types.ExtendedCoinDenom))

			// recipient = startBal + amt
			_, expectedFracBalRecipient := splitExtendedAmount(tt.giveStartBalRecipient.
				Add(tt.giveAmt...).
				AmountOf(types.ExtendedCoinDenom))

			fBalSender := td.keeper.GetFractionalBalance(td.ctx, sender)
			require.Equal(
				t,
				expectedFracBalSender,
				fBalSender,
				"extended balance should be updated for sender",
			)

			fBalRecipient := td.keeper.GetFractionalBalance(td.ctx, sender)
			require.Equal(
				t,
				expectedFracBalRecipient,
				fBalRecipient,
				"extended balance should be updated for receiver",
			)
		})
	}
}

func splitExtendedAmount(extendedAmount sdkmath.Int) (integer, fractional sdkmath.Int) {
	return extendedAmount.Quo(types.ConversionFactor()),
		extendedAmount.Mod(types.ConversionFactor())
}
