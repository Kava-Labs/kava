package types_test

import (
	fmt "fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kava-labs/kava/x/router/types"
)

func TestMsgMintDeposit_Signing(t *testing.T) {
	address := mustAccAddressFromBech32("kava1gepm4nwzz40gtpur93alv9f9wm5ht4l0hzzw9d")
	validatorAddress := mustValAddressFromBech32("kavavaloper1ypjp0m04pyp73hwgtc0dgkx0e9rrydeckewa42")

	msg := types.NewMsgMintDeposit(
		address,
		validatorAddress,
		sdk.NewCoin("ukava", sdk.NewInt(1e9)),
	)

	// checking for the "type" field ensures the msg is registered on the amino codec
	signBytes := []byte(
		`{"type":"router/MsgMintDeposit","value":{"amount":{"amount":"1000000000","denom":"ukava"},"depositor":"kava1gepm4nwzz40gtpur93alv9f9wm5ht4l0hzzw9d","validator":"kavavaloper1ypjp0m04pyp73hwgtc0dgkx0e9rrydeckewa42"}}`,
	)

	assert.Equal(t, []sdk.AccAddress{address}, msg.GetSigners())
	assert.Equal(t, signBytes, msg.GetSignBytes())
}

func TestMsgDelegateMintDeposit_Signing(t *testing.T) {
	address := mustAccAddressFromBech32("kava1gepm4nwzz40gtpur93alv9f9wm5ht4l0hzzw9d")
	validatorAddress := mustValAddressFromBech32("kavavaloper1ypjp0m04pyp73hwgtc0dgkx0e9rrydeckewa42")

	msg := types.NewMsgDelegateMintDeposit(
		address,
		validatorAddress,
		sdk.NewCoin("ukava", sdk.NewInt(1e9)),
	)

	// checking for the "type" field ensures the msg is registered on the amino codec
	signBytes := []byte(
		`{"type":"router/MsgDelegateMintDeposit","value":{"amount":{"amount":"1000000000","denom":"ukava"},"depositor":"kava1gepm4nwzz40gtpur93alv9f9wm5ht4l0hzzw9d","validator":"kavavaloper1ypjp0m04pyp73hwgtc0dgkx0e9rrydeckewa42"}}`,
	)

	assert.Equal(t, []sdk.AccAddress{address}, msg.GetSigners())
	assert.Equal(t, signBytes, msg.GetSignBytes())
}

func TestMsgWithdrawBurn_Signing(t *testing.T) {
	address := mustAccAddressFromBech32("kava1gepm4nwzz40gtpur93alv9f9wm5ht4l0hzzw9d")
	validatorAddress := mustValAddressFromBech32("kavavaloper1ypjp0m04pyp73hwgtc0dgkx0e9rrydeckewa42")

	msg := types.NewMsgWithdrawBurn(
		address,
		validatorAddress,
		sdk.NewCoin("ukava", sdk.NewInt(1e9)),
	)

	// checking for the "type" field ensures the msg is registered on the amino codec
	signBytes := []byte(
		`{"type":"router/MsgWithdrawBurn","value":{"amount":{"amount":"1000000000","denom":"ukava"},"from":"kava1gepm4nwzz40gtpur93alv9f9wm5ht4l0hzzw9d","validator":"kavavaloper1ypjp0m04pyp73hwgtc0dgkx0e9rrydeckewa42"}}`,
	)

	assert.Equal(t, []sdk.AccAddress{address}, msg.GetSigners())
	assert.Equal(t, signBytes, msg.GetSignBytes())
}

func TestMsgWithdrawBurnUndelegate_Signing(t *testing.T) {
	address := mustAccAddressFromBech32("kava1gepm4nwzz40gtpur93alv9f9wm5ht4l0hzzw9d")
	validatorAddress := mustValAddressFromBech32("kavavaloper1ypjp0m04pyp73hwgtc0dgkx0e9rrydeckewa42")

	msg := types.NewMsgWithdrawBurnUndelegate(
		address,
		validatorAddress,
		sdk.NewCoin("ukava", sdk.NewInt(1e9)),
	)

	// checking for the "type" field ensures the msg is registered on the amino codec
	signBytes := []byte(
		`{"type":"router/MsgWithdrawBurnUndelegate","value":{"amount":{"amount":"1000000000","denom":"ukava"},"from":"kava1gepm4nwzz40gtpur93alv9f9wm5ht4l0hzzw9d","validator":"kavavaloper1ypjp0m04pyp73hwgtc0dgkx0e9rrydeckewa42"}}`,
	)

	assert.Equal(t, []sdk.AccAddress{address}, msg.GetSigners())
	assert.Equal(t, signBytes, msg.GetSignBytes())
}

func TestMsg_Validate(t *testing.T) {
	validAddress := "kava1gepm4nwzz40gtpur93alv9f9wm5ht4l0hzzw9d"
	validValidatorAddress := "kavavaloper1ypjp0m04pyp73hwgtc0dgkx0e9rrydeckewa42"
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
		{
			name: "negative coin",
			msgArgs: msgArgs{
				depositor: validAddress,
				validator: validValidatorAddress,
				amount:    sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(-1)},
			},
			expectedErr: sdkerrors.ErrInvalidCoins,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msgMintDeposit := types.MsgMintDeposit{tc.msgArgs.depositor, tc.msgArgs.validator, tc.msgArgs.amount}
			msgDelegateMintDeposit := types.MsgDelegateMintDeposit{tc.msgArgs.depositor, tc.msgArgs.validator, tc.msgArgs.amount}

			msgWithdrawBurn := types.MsgWithdrawBurn{tc.msgArgs.depositor, tc.msgArgs.validator, tc.msgArgs.amount}
			msgWithdrawBurnUndelegate := types.MsgWithdrawBurnUndelegate{tc.msgArgs.depositor, tc.msgArgs.validator, tc.msgArgs.amount}

			msgs := []sdk.Msg{&msgMintDeposit, &msgDelegateMintDeposit, &msgWithdrawBurn, &msgWithdrawBurnUndelegate}
			for _, msg := range msgs {
				t.Run(fmt.Sprintf("%T", msg), func(t *testing.T) {
					err := msg.ValidateBasic()
					if tc.expectedErr == nil {
						require.NoError(t, err)
					} else {
						require.ErrorIs(t, err, tc.expectedErr, "expected error '%s' not found in actual '%s'", tc.expectedErr, err)
					}
				})
			}
		})
	}
}

func mustAccAddressFromBech32(address string) sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		panic(err)
	}
	return addr
}

func mustValAddressFromBech32(address string) sdk.ValAddress {
	addr, err := sdk.ValAddressFromBech32(address)
	if err != nil {
		panic(err)
	}
	return addr
}
