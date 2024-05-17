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
		wantExtBalSender      sdkmath.Int
		wantExtBalRecipient   sdkmath.Int
		wantErr               string
	}{
		{
			"passthrough - ukava",
			cs(c(types.IntegerCoinDenom, 1000)),
			cs(c(types.IntegerCoinDenom, 1000)),
			cs(c(types.IntegerCoinDenom, 1000)),
			sdkmath.ZeroInt(),
			sdkmath.ZeroInt(),
			"",
		},
		{
			"passthrough - busd",
			cs(c("busd", 1000)),
			cs(),
			cs(c("busd", 1000)),
			sdkmath.ZeroInt(),
			sdkmath.ZeroInt(),
			"",
		},
		{
			"akava send - 1akava to 0 balance",
			// Starting balances
			cs(c(types.IntegerCoinDenom, 1000)),
			cs(),
			// Send amount
			cs(c(types.ExtendedCoinDenom, 1)), // akava
			types.ConversionFactor().MulRaw(100).SubRaw(1),
			sdkmath.NewInt(1),
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			td := NewMockedTestData(t)

			td.bk.EXPECT().
				SendCoins(td.ctx, sender, recipient, tt.giveAmt).
				Return(nil)

			err := td.keeper.SendCoins(td.ctx, sender, recipient, tt.giveAmt)

			if tt.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)

			extendedBalSender := td.keeper.GetBalance(td.ctx, sender, types.ExtendedCoinDenom).Amount
			require.Equal(
				t,
				tt.wantExtBalSender,
				extendedBalSender,
				"extended balance should be updated for sender",
			)

			extendedBalRecipient := td.keeper.GetBalance(td.ctx, recipient, types.ExtendedCoinDenom).Amount
			require.Equal(
				t,
				tt.wantExtBalRecipient,
				extendedBalRecipient,
				"extended balance should be updated for receiver",
			)

			// Double check fractional balances in store
			fractionalBal := td.keeper.GetFractionalBalance(td.ctx, sender)
			require.Equal(
				t,
				tt.wantExtBalSender.Mod(types.ConversionFactor()),
				fractionalBal,
			)

			fractionalBal = td.keeper.GetFractionalBalance(td.ctx, recipient)
			require.Equal(
				t,
				tt.wantExtBalRecipient.Mod(types.ConversionFactor()),
				fractionalBal,
			)
		})
	}
}
