package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/kava-labs/kava/x/precisebank/types"
	"github.com/stretchr/testify/require"
)

func TestSendCoinsFromAccountToModule_BlockedReserve(t *testing.T) {
	// Other modules shouldn't be able to send x/precisebank coins as the module
	// account balance is for internal reserve use only.

	td := NewMockedTestData(t)
	td.ak.EXPECT().
		GetModuleAccount(td.ctx, types.ModuleName).
		Return(authtypes.NewModuleAccount(
			authtypes.NewBaseAccountWithAddress(sdk.AccAddress{100}),
			types.ModuleName,
		)).
		Once()

	fromAddr := sdk.AccAddress([]byte{1})
	err := td.keeper.SendCoinsFromAccountToModule(td.ctx, fromAddr, types.ModuleName, cs(c("busd", 1000)))

	require.Error(t, err)
	require.EqualError(t, err, "module account precisebank is not allowed to receive funds: unauthorized")
}

func TestSendCoinsFromModuleToAccount_BlockedReserve(t *testing.T) {
	// Other modules shouldn't be able to send x/precisebank module account
	// funds.

	td := NewMockedTestData(t)
	td.ak.EXPECT().
		GetModuleAddress(types.ModuleName).
		Return(sdk.AccAddress{100}).
		Once()

	toAddr := sdk.AccAddress([]byte{1})
	err := td.keeper.SendCoinsFromModuleToAccount(td.ctx, types.ModuleName, toAddr, cs(c("busd", 1000)))

	require.Error(t, err)
	require.EqualError(t, err, "module account precisebank is not allowed to send funds: unauthorized")
}

func TestSendCoins(t *testing.T) {
	t.Skip()

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
			"passthrough - busd",
			cs(c("busd", 1000)),
			cs(),
			cs(c("busd", 1000)),
			"",
		},
		{
			"akava send - 1akava to 0 balance",
			// Starting balances
			cs(ci(types.ExtendedCoinDenom, types.ConversionFactor().MulRaw(5))),
			cs(),
			// Send amount
			cs(c(types.ExtendedCoinDenom, 1)), // akava
			"",
		},
		{
			"sender borrow from integer",
			// 1ukava, 0 fractional
			cs(ci(types.ExtendedCoinDenom, types.ConversionFactor())),
			cs(),
			// Send 1 with 0 fractional
			cs(c(types.ExtendedCoinDenom, 1)),
			"",
		},
		{
			"receiver carry",
			cs(c(types.ExtendedCoinDenom, 1000)),
			// max fractional amount, carries over to integer
			cs(ci(types.ExtendedCoinDenom, types.ConversionFactor().SubRaw(1))),
			cs(c(types.ExtendedCoinDenom, 1)),
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			td := NewMockedTestData(t)

			require.Zero(
				t,
				tt.giveStartBalSender.AmountOf(types.IntegerCoinDenom).Int64(),
				"giveStartBalSender should use extended denom instead of integer",
			)

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
			_, senderFracBal := splitExtendedAmount(tt.giveStartBalSender.AmountOf(types.ExtendedCoinDenom))
			sendIntAmt, sendFracAmt := splitExtendedAmount(tt.giveAmt.AmountOf(types.ExtendedCoinDenom))

			_, receiverFracBal := splitExtendedAmount(tt.giveStartBalRecipient.AmountOf(types.ExtendedCoinDenom))

			// 4 cases:
			// 1. only sender borrow from integer (send 1ukava to reserve)
			// 2. only receiver carry (receive 1akava from reserve)
			// 3. both sender borrow and receiver carry
			// 4. neither sender borrow nor receiver carry

			senderBorrows := senderFracBal.Sub(sendFracAmt).IsNegative()
			receiverCarries := receiverFracBal.Add(sendFracAmt).GTE(types.ConversionFactor())

			if senderBorrows && !receiverCarries {
				// Sender borrows from integer balance.
				borrowCoin := c(types.IntegerCoinDenom, 1)
				td.bk.EXPECT().
					SendCoinsFromAccountToModule(td.ctx, sender, types.ModuleName, cs(borrowCoin)).
					Return(nil).
					Once()

				transferCoin := ci(types.IntegerCoinDenom, sendIntAmt)
				td.bk.EXPECT().
					SendCoins(td.ctx, sender, recipient, cs(transferCoin)).
					Return(nil).
					Once()
			} else if !senderBorrows && receiverCarries {
				// Receiver carries to integer balance.
				carryCoin := c(types.IntegerCoinDenom, 1)
				td.bk.EXPECT().
					SendCoinsFromModuleToAccount(td.ctx, types.ModuleName, recipient, cs(carryCoin)).
					Return(nil).
					Once()

				transferCoin := ci(types.IntegerCoinDenom, sendIntAmt)
				td.bk.EXPECT().
					SendCoins(td.ctx, sender, recipient, cs(transferCoin)).
					Return(nil).
					Once()
			} else if senderBorrows && receiverCarries {
				// Both sender borrows and receiver carries - direct transfer
				// between accounts.
				sendCoin := c(types.IntegerCoinDenom, 1)
				sendCoin = sendCoin.AddAmount(sendIntAmt)

				td.bk.EXPECT().
					SendCoins(td.ctx, sender, recipient, cs(sendCoin)).
					Return(nil).
					Once()
			}

			// -------------------------------------------
			// x/precisebank send extended coins expectations
			if extendedAmount.IsPositive() {
				td.bk.EXPECT().
					SpendableCoins(td.ctx, sender).
					Return(tt.giveStartBalSender).
					Once()

				// Set fractional balances
				td.keeper.SetFractionalBalance(
					td.ctx,
					sender,
					senderFracBal,
				)

				td.keeper.SetFractionalBalance(
					td.ctx,
					recipient,
					receiverFracBal,
				)
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
