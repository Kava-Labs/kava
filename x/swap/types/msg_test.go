package types_test

import (
	"testing"
	"time"

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
	signData := `{"type":"swap/MsgDeposit","value":{"deadline":"1623606299","depositor":"kava1gepm4nwzz40gtpur93alv9f9wm5ht4l0hzzw9d","slippage":"0.010000000000000000","token_a":{"amount":"1000000","denom":"ukava"},"token_b":{"amount":"5000000","denom":"usdx"}}}`
	signBytes := []byte(signData)

	addr, err := sdk.AccAddressFromBech32("kava1gepm4nwzz40gtpur93alv9f9wm5ht4l0hzzw9d")
	require.NoError(t, err)

	msg := types.NewMsgDeposit(addr, sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.MustNewDecFromStr("0.01"), 1623606299)
	assert.Equal(t, []sdk.AccAddress{addr}, msg.GetSigners())
	assert.Equal(t, signBytes, msg.GetSignBytes())
}

func TestMsgDeposit_Validation(t *testing.T) {
	validMsg := types.NewMsgDeposit(
		sdk.AccAddress("test1"),
		sdk.NewCoin("ukava", sdk.NewInt(1e6)),
		sdk.NewCoin("usdx", sdk.NewInt(5e6)),
		sdk.MustNewDecFromStr("0.01"),
		1623606299,
	)
	require.NoError(t, validMsg.ValidateBasic())

	testCases := []struct {
		name        string
		depositor   sdk.AccAddress
		tokenA      sdk.Coin
		tokenB      sdk.Coin
		slippage    sdk.Dec
		deadline    int64
		expectedErr string
	}{
		{
			name:        "empty address",
			depositor:   sdk.AccAddress(""),
			tokenA:      validMsg.TokenA,
			tokenB:      validMsg.TokenB,
			slippage:    validMsg.Slippage,
			deadline:    validMsg.Deadline,
			expectedErr: "invalid address: depositor address cannot be empty",
		},
		{
			name:        "negative token a",
			depositor:   validMsg.Depositor,
			tokenA:      sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(-1)},
			tokenB:      validMsg.TokenB,
			slippage:    validMsg.Slippage,
			deadline:    validMsg.Deadline,
			expectedErr: "invalid coins: token a deposit amount -1ukava",
		},
		{
			name:        "zero token a",
			depositor:   validMsg.Depositor,
			tokenA:      sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(0)},
			tokenB:      validMsg.TokenB,
			slippage:    validMsg.Slippage,
			deadline:    validMsg.Deadline,
			expectedErr: "invalid coins: token a deposit amount 0ukava",
		},
		{
			name:        "invalid denom token a",
			depositor:   validMsg.Depositor,
			tokenA:      sdk.Coin{Denom: "UKAVA", Amount: sdk.NewInt(1e6)},
			tokenB:      validMsg.TokenB,
			slippage:    validMsg.Slippage,
			deadline:    validMsg.Deadline,
			expectedErr: "invalid coins: token a deposit amount 1000000UKAVA",
		},
		{
			name:        "negative token b",
			depositor:   validMsg.Depositor,
			tokenA:      validMsg.TokenA,
			tokenB:      sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(-1)},
			slippage:    validMsg.Slippage,
			deadline:    validMsg.Deadline,
			expectedErr: "invalid coins: token b deposit amount -1ukava",
		},
		{
			name:        "zero token b",
			depositor:   validMsg.Depositor,
			tokenA:      validMsg.TokenA,
			tokenB:      sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(0)},
			slippage:    validMsg.Slippage,
			deadline:    validMsg.Deadline,
			expectedErr: "invalid coins: token b deposit amount 0ukava",
		},
		{
			name:        "invalid denom token b",
			depositor:   validMsg.Depositor,
			tokenA:      validMsg.TokenA,
			tokenB:      sdk.Coin{Denom: "UKAVA", Amount: sdk.NewInt(1e6)},
			slippage:    validMsg.Slippage,
			deadline:    validMsg.Deadline,
			expectedErr: "invalid coins: token b deposit amount 1000000UKAVA",
		},
		{
			name:        "denoms can not be the same",
			depositor:   validMsg.Depositor,
			tokenA:      sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(1e6)},
			tokenB:      sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(1e6)},
			slippage:    validMsg.Slippage,
			deadline:    validMsg.Deadline,
			expectedErr: "invalid coins: denominations can not be equal",
		},
		{
			name:        "zero deadline",
			depositor:   validMsg.Depositor,
			tokenA:      validMsg.TokenA,
			tokenB:      validMsg.TokenB,
			slippage:    validMsg.Slippage,
			deadline:    0,
			expectedErr: "invalid deadline: deadline 0",
		},
		{
			name:        "negative deadline",
			depositor:   validMsg.Depositor,
			tokenA:      validMsg.TokenA,
			tokenB:      validMsg.TokenB,
			slippage:    validMsg.Slippage,
			deadline:    -1,
			expectedErr: "invalid deadline: deadline -1",
		},
		{
			name:        "negative slippage",
			depositor:   validMsg.Depositor,
			tokenA:      validMsg.TokenA,
			tokenB:      validMsg.TokenB,
			slippage:    sdk.MustNewDecFromStr("-0.01"),
			deadline:    validMsg.Deadline,
			expectedErr: "invalid slippage: slippage can not be negative",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgDeposit(tc.depositor, tc.tokenA, tc.tokenB, tc.slippage, tc.deadline)
			err := msg.ValidateBasic()
			assert.EqualError(t, err, tc.expectedErr)
		})
	}
}

func TestMsgDeposit_Deadline(t *testing.T) {
	blockTime := time.Now()

	testCases := []struct {
		name       string
		deadline   int64
		isExceeded bool
	}{
		{
			name:       "deadline in future",
			deadline:   blockTime.Add(1 * time.Second).Unix(),
			isExceeded: false,
		},
		{
			name:       "deadline in past",
			deadline:   blockTime.Add(-1 * time.Second).Unix(),
			isExceeded: true,
		},
		{
			name:       "deadline is equal",
			deadline:   blockTime.Unix(),
			isExceeded: true,
		},
	}

	for _, tc := range testCases {
		msg := types.NewMsgDeposit(
			sdk.AccAddress("test1"),
			sdk.NewCoin("ukava", sdk.NewInt(1e6)),
			sdk.NewCoin("usdx", sdk.NewInt(5e6)),
			sdk.MustNewDecFromStr("0.01"),
			tc.deadline,
		)
		require.NoError(t, msg.ValidateBasic())
		assert.Equal(t, tc.isExceeded, msg.DeadlineExceeded(blockTime))
		assert.Equal(t, time.Unix(tc.deadline, 0), msg.GetDeadline())
	}
}

func TestMsgWithdraw_Attributes(t *testing.T) {
	msg := types.MsgWithdraw{}
	assert.Equal(t, "swap", msg.Route())
	assert.Equal(t, "swap_withdraw", msg.Type())
}

func TestMsgWithdraw_Signing(t *testing.T) {
	signData := `{"type":"swap/MsgWithdraw","value":{"deadline":"1623606299","expected_coin_a":{"amount":"1000000","denom":"ukava"},"expected_coin_b":{"amount":"2000000","denom":"usdx"},"from":"kava1gepm4nwzz40gtpur93alv9f9wm5ht4l0hzzw9d","shares":"1500000","slippage":"0.050000000000000000"}}`
	signBytes := []byte(signData)

	addr, err := sdk.AccAddressFromBech32("kava1gepm4nwzz40gtpur93alv9f9wm5ht4l0hzzw9d")
	require.NoError(t, err)

	msg := types.NewMsgWithdraw(
		addr,
		sdk.NewInt(1500000),
		sdk.MustNewDecFromStr("0.05"),
		sdk.NewCoin("ukava", sdk.NewInt(1000000)),
		sdk.NewCoin("usdx", sdk.NewInt(2000000)),
		1623606299,
	)
	assert.Equal(t, []sdk.AccAddress{addr}, msg.GetSigners())
	assert.Equal(t, signBytes, msg.GetSignBytes())
}

func TestMsgWithdraw_Validation(t *testing.T) {
	validMsg := types.NewMsgWithdraw(
		sdk.AccAddress("test1"),
		sdk.NewInt(1500000),
		sdk.MustNewDecFromStr("0.05"),
		sdk.NewCoin("ukava", sdk.NewInt(1000000)),
		sdk.NewCoin("usdx", sdk.NewInt(2000000)),
		1623606299,
	)
	require.NoError(t, validMsg.ValidateBasic())

	testCases := []struct {
		name          string
		from          sdk.AccAddress
		shares        sdk.Int
		slippage      sdk.Dec
		expectedCoinA sdk.Coin
		expectedCoinB sdk.Coin
		deadline      int64
		expectedErr   string
	}{
		{
			name:          "empty address",
			from:          sdk.AccAddress(""),
			shares:        validMsg.Shares,
			slippage:      validMsg.Slippage,
			expectedCoinA: validMsg.ExpectedCoinA,
			expectedCoinB: validMsg.ExpectedCoinB,
			deadline:      validMsg.Deadline,
			expectedErr:   "invalid address: from address cannot be empty",
		},
		{
			name:          "0 shares",
			from:          validMsg.From,
			shares:        sdk.ZeroInt(),
			slippage:      validMsg.Slippage,
			expectedCoinA: validMsg.ExpectedCoinA,
			expectedCoinB: validMsg.ExpectedCoinB,
			deadline:      validMsg.Deadline,
			expectedErr:   "invalid shares: 0",
		},
		{
			name:          "negative shares",
			from:          validMsg.From,
			shares:        sdk.ZeroInt().Sub(sdk.OneInt()),
			slippage:      validMsg.Slippage,
			expectedCoinA: validMsg.ExpectedCoinA,
			expectedCoinB: validMsg.ExpectedCoinB,
			deadline:      validMsg.Deadline,
			expectedErr:   "invalid shares: -1",
		},
		{
			name:          "negative slippage",
			from:          validMsg.From,
			shares:        validMsg.Shares,
			slippage:      sdk.MustNewDecFromStr("-0.05"),
			expectedCoinA: validMsg.ExpectedCoinA,
			expectedCoinB: validMsg.ExpectedCoinB,
			deadline:      validMsg.Deadline,
			expectedErr:   "invalid slippage: -0.050000000000000000",
		},
		{
			name:          "slippage too large",
			from:          validMsg.From,
			shares:        validMsg.Shares,
			slippage:      sdk.MustNewDecFromStr("1.1"),
			expectedCoinA: validMsg.ExpectedCoinA,
			expectedCoinB: validMsg.ExpectedCoinB,
			deadline:      validMsg.Deadline,
			expectedErr:   "invalid slippage: 1.100000000000000000",
		},
		{
			name:          "negative deadline",
			from:          validMsg.From,
			shares:        validMsg.Shares,
			slippage:      validMsg.Slippage,
			expectedCoinA: validMsg.ExpectedCoinA,
			expectedCoinB: validMsg.ExpectedCoinB,
			deadline:      -1,
			expectedErr:   "invalid deadline: deadline -1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgWithdraw(tc.from, tc.shares, tc.slippage, tc.expectedCoinA, tc.expectedCoinB, tc.deadline)
			err := msg.ValidateBasic()
			assert.EqualError(t, err, tc.expectedErr)
		})
	}
}

func TestMsgWithdraw_Deadline(t *testing.T) {
	blockTime := time.Now()

	testCases := []struct {
		name       string
		deadline   int64
		isExceeded bool
	}{
		{
			name:       "deadline in future",
			deadline:   blockTime.Add(1 * time.Second).Unix(),
			isExceeded: false,
		},
		{
			name:       "deadline in past",
			deadline:   blockTime.Add(-1 * time.Second).Unix(),
			isExceeded: true,
		},
		{
			name:       "deadline is equal",
			deadline:   blockTime.Unix(),
			isExceeded: true,
		},
	}

	for _, tc := range testCases {
		msg := types.NewMsgWithdraw(
			sdk.AccAddress("test1"),
			sdk.NewInt(1500000),
			sdk.MustNewDecFromStr("0.05"),
			sdk.NewCoin("ukava", sdk.NewInt(1000000)),
			sdk.NewCoin("usdx", sdk.NewInt(2000000)),
			tc.deadline,
		)
		require.NoError(t, msg.ValidateBasic())
		assert.Equal(t, tc.isExceeded, msg.DeadlineExceeded(blockTime))
		assert.Equal(t, time.Unix(tc.deadline, 0), msg.GetDeadline())
	}
}
