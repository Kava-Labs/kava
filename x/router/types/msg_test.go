package types_test

import (
	fmt "fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/kava-labs/kava/x/router/types"
)

func TestMsg_Validate(t *testing.T) {
	validAddress := sdk.AccAddress("test address--------").String()
	validValidatorAddress := sdk.ValAddress("test address--------").String()
	validCoin := sdk.NewInt64Coin("ukava", 1e9)

	type msgArgs struct {
		depositor string
		validator string
		amount    sdk.Coin
	}
	tests := []struct {
		name        string
		msgArgs     msgArgs
		expectedErr error
	}{
		{
			name: "normal multiplier is valid",
			msgArgs: msgArgs{
				depositor: validAddress,
				validator: validValidatorAddress,
				amount:    validCoin,
			},
		},
		{
			name: "invalid depositor",
			msgArgs: msgArgs{
				depositor: "invalid",
				validator: validValidatorAddress,
				amount:    validCoin,
			},
			expectedErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "empty depositor",
			msgArgs: msgArgs{
				depositor: "",
				validator: validValidatorAddress,
				amount:    validCoin,
			},
			expectedErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid validator",
			msgArgs: msgArgs{
				depositor: validAddress,
				validator: "invalid",
				amount:    validCoin,
			},
			expectedErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "nil coin",
			msgArgs: msgArgs{
				depositor: validAddress,
				validator: validValidatorAddress,
				amount:    sdk.Coin{},
			},
			expectedErr: sdkerrors.ErrInvalidCoins,
		},
		{
			name: "zero coin",
			msgArgs: msgArgs{
				depositor: validAddress,
				validator: validValidatorAddress,
				amount:    sdk.NewCoin("ukava", sdk.ZeroInt()),
			},
			expectedErr: sdkerrors.ErrInvalidCoins,
		},
	}

	for _, tc := range tests {
		msgMintDeposit := types.MsgMintDeposit{tc.msgArgs.depositor, tc.msgArgs.validator, tc.msgArgs.amount}
		msgDelegateMintDeposit := types.MsgDelegateMintDeposit{tc.msgArgs.depositor, tc.msgArgs.validator, tc.msgArgs.amount}

		msgWithdrawBurn := types.MsgWithdrawBurn{tc.msgArgs.depositor, tc.msgArgs.validator, tc.msgArgs.amount}
		msgWithdrawBurnUndelegate := types.MsgWithdrawBurnUndelegate{tc.msgArgs.depositor, tc.msgArgs.validator, tc.msgArgs.amount}

		msgs := []sdk.Msg{&msgMintDeposit, &msgDelegateMintDeposit, &msgWithdrawBurn, &msgWithdrawBurnUndelegate}
		for _, msg := range msgs {
			t.Run(fmt.Sprintf("%s%T", tc.name, msg), func(t *testing.T) {
				err := msg.ValidateBasic()
				if tc.expectedErr == nil {
					require.NoError(t, err)
				} else {
					require.ErrorIs(t, err, tc.expectedErr, "expected error '%s' not found in actual '%s'", tc.expectedErr, err)
				}
			})
		}
	}
}
