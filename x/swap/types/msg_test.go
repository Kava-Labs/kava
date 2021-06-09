package types_test

import (
	"testing"

	"github.com/kava-labs/kava/x/swap/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMsgDeposit_Attributes(t *testing.T) {
	msg := types.MsgDeposit{}
	assert.Equal(t, "swap", msg.Route())
	assert.Equal(t, "swap_deposit", msg.Type())
}

func TestMsgDeposit_Signing(t *testing.T) {
	signData := `{"depositor":"kava1gepm4nwzz40gtpur93alv9f9wm5ht4l0hzzw9d","token_a":{"amount":"1000000","denom":"ukava"},"token_b":{"amount":"5000000","denom":"usdx"}}`
	signBytes := []byte(signData)

	addr, err := sdk.AccAddressFromBech32("kava1gepm4nwzz40gtpur93alv9f9wm5ht4l0hzzw9d")
	require.NoError(t, err)

	msg := types.NewMsgDeposit(addr, sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.NewCoin("usdx", sdk.NewInt(5e6)))
	assert.Equal(t, []sdk.AccAddress{addr}, msg.GetSigners())
	assert.Equal(t, signBytes, msg.GetSignBytes())
}

func TestMsgDeposit_Validation(t *testing.T) {
	validMsg := types.NewMsgDeposit(
		sdk.AccAddress("test1"),
		sdk.NewCoin("ukava", sdk.NewInt(1e6)),
		sdk.NewCoin("usdx", sdk.NewInt(5e6)),
	)
	require.NoError(t, validMsg.ValidateBasic())

	testCases := []struct {
		name        string
		depositor   sdk.AccAddress
		tokenA      sdk.Coin
		tokenB      sdk.Coin
		expectedErr string
	}{
		{
			name:        "empty address",
			depositor:   sdk.AccAddress(""),
			tokenA:      validMsg.TokenA,
			tokenB:      validMsg.TokenB,
			expectedErr: "invalid address: depositor address cannot be empty",
		},
		{
			name:        "negative token a",
			depositor:   validMsg.Depositor,
			tokenA:      sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(-1)},
			tokenB:      validMsg.TokenB,
			expectedErr: "invalid coins: token a deposit amount -1ukava",
		},
		{
			name:        "zero token a",
			depositor:   validMsg.Depositor,
			tokenA:      sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(0)},
			tokenB:      validMsg.TokenB,
			expectedErr: "invalid coins: token a deposit amount 0ukava",
		},
		{
			name:        "invalid denom token a",
			depositor:   validMsg.Depositor,
			tokenA:      sdk.Coin{Denom: "UKAVA", Amount: sdk.NewInt(1e6)},
			tokenB:      validMsg.TokenB,
			expectedErr: "invalid coins: token a deposit amount 1000000UKAVA",
		},
		{
			name:        "negative token b",
			depositor:   validMsg.Depositor,
			tokenA:      validMsg.TokenA,
			tokenB:      sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(-1)},
			expectedErr: "invalid coins: token b deposit amount -1ukava",
		},
		{
			name:        "zero token b",
			depositor:   validMsg.Depositor,
			tokenA:      validMsg.TokenA,
			tokenB:      sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(0)},
			expectedErr: "invalid coins: token b deposit amount 0ukava",
		},
		{
			name:        "invalid denom token b",
			depositor:   validMsg.Depositor,
			tokenA:      validMsg.TokenA,
			tokenB:      sdk.Coin{Denom: "UKAVA", Amount: sdk.NewInt(1e6)},
			expectedErr: "invalid coins: token b deposit amount 1000000UKAVA",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgDeposit(tc.depositor, tc.tokenA, tc.tokenB)
			err := msg.ValidateBasic()
			assert.EqualError(t, err, tc.expectedErr)
		})
	}
}
