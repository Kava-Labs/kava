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

	msg := types.NewMsgDeposit(addr.String(), sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.MustNewDecFromStr("0.01"), 1623606299)
	assert.Equal(t, []sdk.AccAddress{addr}, msg.GetSigners())
	assert.Equal(t, signBytes, msg.GetSignBytes())
}

func TestMsgDeposit_Validation(t *testing.T) {
	addr, err := sdk.AccAddressFromBech32("kava1gepm4nwzz40gtpur93alv9f9wm5ht4l0hzzw9d")
	require.NoError(t, err)

	validMsg := types.NewMsgDeposit(
		addr.String(),
		sdk.NewCoin("ukava", sdk.NewInt(1e6)),
		sdk.NewCoin("usdx", sdk.NewInt(5e6)),
		sdk.MustNewDecFromStr("0.01"),
		1623606299,
	)
	require.NoError(t, validMsg.ValidateBasic())

	testCases := []struct {
		name        string
		depositor   string
		tokenA      sdk.Coin
		tokenB      sdk.Coin
		slippage    sdk.Dec
		deadline    int64
		expectedErr string
	}{
		{
			name:        "empty address",
			depositor:   "",
			tokenA:      validMsg.TokenA,
			tokenB:      validMsg.TokenB,
			slippage:    validMsg.Slippage,
			deadline:    validMsg.Deadline,
			expectedErr: "depositor address cannot be empty: invalid address",
		},
		{
			name:        "invalid address",
			depositor:   "kava1abcde",
			tokenA:      validMsg.TokenA,
			tokenB:      validMsg.TokenB,
			slippage:    validMsg.Slippage,
			deadline:    validMsg.Deadline,
			expectedErr: "invalid depositor address: decoding bech32 failed: invalid index of 1: invalid address",
		},
		{
			name:        "negative token a",
			depositor:   validMsg.Depositor,
			tokenA:      sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(-1)},
			tokenB:      validMsg.TokenB,
			slippage:    validMsg.Slippage,
			deadline:    validMsg.Deadline,
			expectedErr: "token a deposit amount -1ukava: invalid coins",
		},
		{
			name:        "zero token a",
			depositor:   validMsg.Depositor,
			tokenA:      sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(0)},
			tokenB:      validMsg.TokenB,
			slippage:    validMsg.Slippage,
			deadline:    validMsg.Deadline,
			expectedErr: "token a deposit amount 0ukava: invalid coins",
		},
		// TODO: denom can now be uppercase
		// {
		// 	name:        "invalid denom token a",
		// 	depositor:   validMsg.Depositor,
		// 	tokenA:      sdk.Coin{Denom: "UKAVA", Amount: sdk.NewInt(1e6)},
		// 	tokenB:      validMsg.TokenB,
		// 	slippage:    validMsg.Slippage,
		// 	deadline:    validMsg.Deadline,
		// 	expectedErr: "token a deposit amount 1000000UKAVA: invalid coins",
		// },
		{
			name:        "negative token b",
			depositor:   validMsg.Depositor,
			tokenA:      validMsg.TokenA,
			tokenB:      sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(-1)},
			slippage:    validMsg.Slippage,
			deadline:    validMsg.Deadline,
			expectedErr: "token b deposit amount -1ukava: invalid coins",
		},
		{
			name:        "zero token b",
			depositor:   validMsg.Depositor,
			tokenA:      validMsg.TokenA,
			tokenB:      sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(0)},
			slippage:    validMsg.Slippage,
			deadline:    validMsg.Deadline,
			expectedErr: "token b deposit amount 0ukava: invalid coins",
		},
		// TODO: denom can now be uppercase
		// {
		// 	name:        "invalid denom token b",
		// 	depositor:   validMsg.Depositor,
		// 	tokenA:      validMsg.TokenA,
		// 	tokenB:      sdk.Coin{Denom: "UKAVA", Amount: sdk.NewInt(1e6)},
		// 	slippage:    validMsg.Slippage,
		// 	deadline:    validMsg.Deadline,
		// 	expectedErr: "token b deposit amount 1000000UKAVA: invalid coins",
		// },
		{
			name:        "denoms can not be the same",
			depositor:   validMsg.Depositor,
			tokenA:      sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(1e6)},
			tokenB:      sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(1e6)},
			slippage:    validMsg.Slippage,
			deadline:    validMsg.Deadline,
			expectedErr: "denominations can not be equal: invalid coins",
		},
		{
			name:        "zero deadline",
			depositor:   validMsg.Depositor,
			tokenA:      validMsg.TokenA,
			tokenB:      validMsg.TokenB,
			slippage:    validMsg.Slippage,
			deadline:    0,
			expectedErr: "deadline 0: invalid deadline",
		},
		{
			name:        "negative deadline",
			depositor:   validMsg.Depositor,
			tokenA:      validMsg.TokenA,
			tokenB:      validMsg.TokenB,
			slippage:    validMsg.Slippage,
			deadline:    -1,
			expectedErr: "deadline -1: invalid deadline",
		},
		{
			name:        "negative slippage",
			depositor:   validMsg.Depositor,
			tokenA:      validMsg.TokenA,
			tokenB:      validMsg.TokenB,
			slippage:    sdk.MustNewDecFromStr("-0.01"),
			deadline:    validMsg.Deadline,
			expectedErr: "slippage can not be negative: invalid slippage",
		},
		{
			name:        "nil slippage",
			depositor:   validMsg.Depositor,
			tokenA:      validMsg.TokenA,
			tokenB:      validMsg.TokenB,
			slippage:    sdk.Dec{},
			deadline:    validMsg.Deadline,
			expectedErr: "slippage must be set: invalid slippage",
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
			sdk.AccAddress("test1").String(),
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
	signData := `{"type":"swap/MsgWithdraw","value":{"deadline":"1623606299","from":"kava1gepm4nwzz40gtpur93alv9f9wm5ht4l0hzzw9d","min_token_a":{"amount":"1000000","denom":"ukava"},"min_token_b":{"amount":"2000000","denom":"usdx"},"shares":"1500000"}}`
	signBytes := []byte(signData)

	addr, err := sdk.AccAddressFromBech32("kava1gepm4nwzz40gtpur93alv9f9wm5ht4l0hzzw9d")
	require.NoError(t, err)

	msg := types.NewMsgWithdraw(
		addr.String(),
		sdk.NewInt(1500000),
		sdk.NewCoin("ukava", sdk.NewInt(1000000)),
		sdk.NewCoin("usdx", sdk.NewInt(2000000)),
		1623606299,
	)
	assert.Equal(t, []sdk.AccAddress{addr}, msg.GetSigners())
	assert.Equal(t, signBytes, msg.GetSignBytes())
}

func TestMsgWithdraw_Validation(t *testing.T) {
	validMsg := types.NewMsgWithdraw(
		sdk.AccAddress("test1").String(),
		sdk.NewInt(1500000),
		sdk.NewCoin("ukava", sdk.NewInt(1000000)),
		sdk.NewCoin("usdx", sdk.NewInt(2000000)),
		1623606299,
	)
	require.NoError(t, validMsg.ValidateBasic())

	testCases := []struct {
		name        string
		from        string
		shares      sdk.Int
		minTokenA   sdk.Coin
		minTokenB   sdk.Coin
		deadline    int64
		expectedErr string
	}{
		{
			name:        "empty address",
			from:        "",
			shares:      validMsg.Shares,
			minTokenA:   validMsg.MinTokenA,
			minTokenB:   validMsg.MinTokenB,
			deadline:    validMsg.Deadline,
			expectedErr: "from address cannot be empty: invalid address",
		},
		{
			name:        "invalid address",
			from:        "kava1abcde",
			shares:      validMsg.Shares,
			minTokenA:   validMsg.MinTokenA,
			minTokenB:   validMsg.MinTokenB,
			deadline:    validMsg.Deadline,
			expectedErr: "invalid from address: decoding bech32 failed: invalid index of 1: invalid address",
		},
		{
			name:        "zero token a",
			from:        validMsg.From,
			shares:      validMsg.Shares,
			minTokenA:   sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(0)},
			minTokenB:   validMsg.MinTokenB,
			deadline:    validMsg.Deadline,
			expectedErr: "min token a amount 0ukava: invalid coins",
		},
		// TODO: Uppercase denoms allowed
		// {
		// 	name:        "invalid denom token a",
		// 	from:        validMsg.From,
		// 	shares:      validMsg.Shares,
		// 	minTokenA:   sdk.Coin{Denom: "UKAVA", Amount: sdk.NewInt(1e6)},
		// 	minTokenB:   validMsg.MinTokenB,
		// 	deadline:    validMsg.Deadline,
		// 	expectedErr: "min token a amount 1000000UKAVA: invalid coins",
		// },
		{
			name:        "negative token b",
			from:        validMsg.From,
			shares:      validMsg.Shares,
			minTokenA:   validMsg.MinTokenA,
			minTokenB:   sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(-1)},
			deadline:    validMsg.Deadline,
			expectedErr: "min token b amount -1ukava: invalid coins",
		},
		{
			name:        "zero token b",
			from:        validMsg.From,
			shares:      validMsg.Shares,
			minTokenA:   validMsg.MinTokenA,
			minTokenB:   sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(0)},
			deadline:    validMsg.Deadline,
			expectedErr: "min token b amount 0ukava: invalid coins",
		},
		// TODO: Uppercase denoms allowed
		// {
		// 	name:        "invalid denom token b",
		// 	from:        validMsg.From,
		// 	shares:      validMsg.Shares,
		// 	minTokenA:   validMsg.MinTokenA,
		// 	minTokenB:   sdk.Coin{Denom: "UKAVA", Amount: sdk.NewInt(1e6)},
		// 	deadline:    validMsg.Deadline,
		// 	expectedErr: "min token b amount 1000000UKAVA: invalid coins",
		// },
		{
			name:        "denoms can not be the same",
			from:        validMsg.From,
			shares:      validMsg.Shares,
			minTokenA:   sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(1e6)},
			minTokenB:   sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(1e6)},
			deadline:    validMsg.Deadline,
			expectedErr: "denominations can not be equal: invalid coins",
		},
		{
			name:        "zero shares",
			from:        validMsg.From,
			shares:      sdk.ZeroInt(),
			minTokenA:   validMsg.MinTokenA,
			minTokenB:   validMsg.MinTokenB,
			deadline:    validMsg.Deadline,
			expectedErr: "0: invalid shares",
		},
		{
			name:        "negative shares",
			from:        validMsg.From,
			shares:      sdk.NewInt(-1),
			minTokenA:   validMsg.MinTokenA,
			minTokenB:   validMsg.MinTokenB,
			deadline:    validMsg.Deadline,
			expectedErr: "-1: invalid shares",
		},
		{
			name:        "nil shares",
			from:        validMsg.From,
			shares:      sdk.Int{},
			minTokenA:   validMsg.MinTokenA,
			minTokenB:   validMsg.MinTokenB,
			deadline:    validMsg.Deadline,
			expectedErr: "shares must be set: invalid shares",
		},
		{
			name:        "zero deadline",
			from:        validMsg.From,
			shares:      validMsg.Shares,
			minTokenA:   validMsg.MinTokenA,
			minTokenB:   validMsg.MinTokenB,
			deadline:    0,
			expectedErr: "deadline 0: invalid deadline",
		},
		{
			name:        "negative deadline",
			from:        validMsg.From,
			shares:      validMsg.Shares,
			minTokenA:   validMsg.MinTokenA,
			minTokenB:   validMsg.MinTokenB,
			deadline:    -1,
			expectedErr: "deadline -1: invalid deadline",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgWithdraw(tc.from, tc.shares, tc.minTokenA, tc.minTokenB, tc.deadline)
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
			sdk.AccAddress("test1").String(),
			sdk.NewInt(1500000),
			sdk.NewCoin("ukava", sdk.NewInt(1000000)),
			sdk.NewCoin("usdx", sdk.NewInt(2000000)),
			tc.deadline,
		)
		require.NoError(t, msg.ValidateBasic())
		assert.Equal(t, tc.isExceeded, msg.DeadlineExceeded(blockTime))
		assert.Equal(t, time.Unix(tc.deadline, 0), msg.GetDeadline())
	}
}

func TestMsgSwapExactForTokens_Attributes(t *testing.T) {
	msg := types.MsgSwapExactForTokens{}
	assert.Equal(t, "swap", msg.Route())
	assert.Equal(t, "swap_exact_for_tokens", msg.Type())
}

func TestMsgSwapExactForTokens_Signing(t *testing.T) {
	signData := `{"type":"swap/MsgSwapExactForTokens","value":{"deadline":"1623606299","exact_token_a":{"amount":"1000000","denom":"ukava"},"requester":"kava1gepm4nwzz40gtpur93alv9f9wm5ht4l0hzzw9d","slippage":"0.010000000000000000","token_b":{"amount":"5000000","denom":"usdx"}}}`
	signBytes := []byte(signData)

	addr, err := sdk.AccAddressFromBech32("kava1gepm4nwzz40gtpur93alv9f9wm5ht4l0hzzw9d")
	require.NoError(t, err)

	msg := types.NewMsgSwapExactForTokens(addr.String(), sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.MustNewDecFromStr("0.01"), 1623606299)
	assert.Equal(t, []sdk.AccAddress{addr}, msg.GetSigners())
	assert.Equal(t, signBytes, msg.GetSignBytes())
}

func TestMsgSwapExactForTokens_Validation(t *testing.T) {
	validMsg := types.NewMsgSwapExactForTokens(
		sdk.AccAddress("test1").String(),
		sdk.NewCoin("ukava", sdk.NewInt(1e6)),
		sdk.NewCoin("usdx", sdk.NewInt(5e6)),
		sdk.MustNewDecFromStr("0.01"),
		1623606299,
	)
	require.NoError(t, validMsg.ValidateBasic())

	testCases := []struct {
		name        string
		requester   string
		exactTokenA sdk.Coin
		tokenB      sdk.Coin
		slippage    sdk.Dec
		deadline    int64
		expectedErr string
	}{
		{
			name:        "empty address",
			requester:   sdk.AccAddress("").String(),
			exactTokenA: validMsg.ExactTokenA,
			tokenB:      validMsg.TokenB,
			slippage:    validMsg.Slippage,
			deadline:    validMsg.Deadline,
			expectedErr: "requester address cannot be empty: invalid address",
		},
		{
			name:        "invalid address",
			requester:   "kava1abcde",
			exactTokenA: validMsg.ExactTokenA,
			tokenB:      validMsg.TokenB,
			slippage:    validMsg.Slippage,
			deadline:    validMsg.Deadline,
			expectedErr: "invalid requester address: decoding bech32 failed: invalid index of 1: invalid address",
		},
		{
			name:        "negative token a",
			requester:   validMsg.Requester,
			exactTokenA: sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(-1)},
			tokenB:      validMsg.TokenB,
			slippage:    validMsg.Slippage,
			deadline:    validMsg.Deadline,
			expectedErr: "exact token a deposit amount -1ukava: invalid coins",
		},
		{
			name:        "zero token a",
			requester:   validMsg.Requester,
			exactTokenA: sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(0)},
			tokenB:      validMsg.TokenB,
			slippage:    validMsg.Slippage,
			deadline:    validMsg.Deadline,
			expectedErr: "exact token a deposit amount 0ukava: invalid coins",
		},
		// TODO:
		// {
		// 	name:        "invalid denom token a",
		// 	requester:   validMsg.Requester,
		// 	exactTokenA: sdk.Coin{Denom: "UKAVA", Amount: sdk.NewInt(1e6)},
		// 	tokenB:      validMsg.TokenB,
		// 	slippage:    validMsg.Slippage,
		// 	deadline:    validMsg.Deadline,
		// 	expectedErr: "exact token a deposit amount 1000000UKAVA: invalid coins",
		// },
		{
			name:        "negative token b",
			requester:   validMsg.Requester,
			exactTokenA: validMsg.ExactTokenA,
			tokenB:      sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(-1)},
			slippage:    validMsg.Slippage,
			deadline:    validMsg.Deadline,
			expectedErr: "token b deposit amount -1ukava: invalid coins",
		},
		{
			name:        "zero token b",
			requester:   validMsg.Requester,
			exactTokenA: validMsg.ExactTokenA,
			tokenB:      sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(0)},
			slippage:    validMsg.Slippage,
			deadline:    validMsg.Deadline,
			expectedErr: "token b deposit amount 0ukava: invalid coins",
		},
		// TODO:
		// {
		// 	name:        "invalid denom token b",
		// 	requester:   validMsg.Requester,
		// 	exactTokenA: validMsg.ExactTokenA,
		// 	tokenB:      sdk.Coin{Denom: "UKAVA", Amount: sdk.NewInt(1e6)},
		// 	slippage:    validMsg.Slippage,
		// 	deadline:    validMsg.Deadline,
		// 	expectedErr: "token b deposit amount 1000000UKAVA: invalid coins",
		// },
		{
			name:        "denoms can not be the same",
			requester:   validMsg.Requester,
			exactTokenA: sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(1e6)},
			tokenB:      sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(1e6)},
			slippage:    validMsg.Slippage,
			deadline:    validMsg.Deadline,
			expectedErr: "denominations can not be equal: invalid coins",
		},
		{
			name:        "zero deadline",
			requester:   validMsg.Requester,
			exactTokenA: validMsg.ExactTokenA,
			tokenB:      validMsg.TokenB,
			slippage:    validMsg.Slippage,
			deadline:    0,
			expectedErr: "deadline 0: invalid deadline",
		},
		{
			name:        "negative deadline",
			requester:   validMsg.Requester,
			exactTokenA: validMsg.ExactTokenA,
			tokenB:      validMsg.TokenB,
			slippage:    validMsg.Slippage,
			deadline:    -1,
			expectedErr: "deadline -1: invalid deadline",
		},
		{
			name:        "negative slippage",
			requester:   validMsg.Requester,
			exactTokenA: validMsg.ExactTokenA,
			tokenB:      validMsg.TokenB,
			slippage:    sdk.MustNewDecFromStr("-0.01"),
			deadline:    validMsg.Deadline,
			expectedErr: "slippage can not be negative: invalid slippage",
		},
		{
			name:        "nil slippage",
			requester:   validMsg.Requester,
			exactTokenA: validMsg.ExactTokenA,
			tokenB:      validMsg.TokenB,
			slippage:    sdk.Dec{},
			deadline:    validMsg.Deadline,
			expectedErr: "slippage must be set: invalid slippage",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgSwapExactForTokens(tc.requester, tc.exactTokenA, tc.tokenB, tc.slippage, tc.deadline)
			err := msg.ValidateBasic()
			assert.EqualError(t, err, tc.expectedErr)
		})
	}
}

func TestMsgSwapExactForTokens_Deadline(t *testing.T) {
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
		msg := types.NewMsgSwapExactForTokens(
			sdk.AccAddress("test1").String(),
			sdk.NewCoin("ukava", sdk.NewInt(1000000)),
			sdk.NewCoin("usdx", sdk.NewInt(2000000)),
			sdk.MustNewDecFromStr("0.01"),
			tc.deadline,
		)
		require.NoError(t, msg.ValidateBasic())
		assert.Equal(t, tc.isExceeded, msg.DeadlineExceeded(blockTime))
		assert.Equal(t, time.Unix(tc.deadline, 0), msg.GetDeadline())
	}
}

func TestMsgSwapForExactTokens_Attributes(t *testing.T) {
	msg := types.MsgSwapForExactTokens{}
	assert.Equal(t, "swap", msg.Route())
	assert.Equal(t, "swap_for_exact_tokens", msg.Type())
}

func TestMsgSwapForExactTokens_Signing(t *testing.T) {
	signData := `{"type":"swap/MsgSwapForExactTokens","value":{"deadline":"1623606299","exact_token_b":{"amount":"5000000","denom":"usdx"},"requester":"kava1gepm4nwzz40gtpur93alv9f9wm5ht4l0hzzw9d","slippage":"0.010000000000000000","token_a":{"amount":"1000000","denom":"ukava"}}}`
	signBytes := []byte(signData)

	addr, err := sdk.AccAddressFromBech32("kava1gepm4nwzz40gtpur93alv9f9wm5ht4l0hzzw9d")
	require.NoError(t, err)

	msg := types.NewMsgSwapForExactTokens(addr.String(), sdk.NewCoin("ukava", sdk.NewInt(1e6)), sdk.NewCoin("usdx", sdk.NewInt(5e6)), sdk.MustNewDecFromStr("0.01"), 1623606299)
	assert.Equal(t, []sdk.AccAddress{addr}, msg.GetSigners())
	assert.Equal(t, signBytes, msg.GetSignBytes())
}

func TestMsgSwapForExactTokens_Validation(t *testing.T) {
	validMsg := types.NewMsgSwapForExactTokens(
		sdk.AccAddress("test1").String(),
		sdk.NewCoin("ukava", sdk.NewInt(1e6)),
		sdk.NewCoin("usdx", sdk.NewInt(5e6)),
		sdk.MustNewDecFromStr("0.01"),
		1623606299,
	)
	require.NoError(t, validMsg.ValidateBasic())

	testCases := []struct {
		name        string
		requester   string
		tokenA      sdk.Coin
		exactTokenB sdk.Coin
		slippage    sdk.Dec
		deadline    int64
		expectedErr string
	}{
		{
			name:        "empty address",
			requester:   sdk.AccAddress("").String(),
			tokenA:      validMsg.TokenA,
			exactTokenB: validMsg.ExactTokenB,
			slippage:    validMsg.Slippage,
			deadline:    validMsg.Deadline,
			expectedErr: "requester address cannot be empty: invalid address",
		},
		{
			name:        "invalid address",
			requester:   "kava1abcde",
			tokenA:      validMsg.TokenA,
			exactTokenB: validMsg.ExactTokenB,
			slippage:    validMsg.Slippage,
			deadline:    validMsg.Deadline,
			expectedErr: "invalid requester address: decoding bech32 failed: invalid index of 1: invalid address",
		},
		{
			name:        "negative token a",
			requester:   validMsg.Requester,
			tokenA:      sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(-1)},
			exactTokenB: validMsg.ExactTokenB,
			slippage:    validMsg.Slippage,
			deadline:    validMsg.Deadline,
			expectedErr: "token a deposit amount -1ukava: invalid coins",
		},
		{
			name:        "zero token a",
			requester:   validMsg.Requester,
			tokenA:      sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(0)},
			exactTokenB: validMsg.ExactTokenB,
			slippage:    validMsg.Slippage,
			deadline:    validMsg.Deadline,
			expectedErr: "token a deposit amount 0ukava: invalid coins",
		},
		// TODO:
		// {
		// 	name:        "invalid denom token a",
		// 	requester:   validMsg.Requester,
		// 	tokenA:      sdk.Coin{Denom: "UKAVA", Amount: sdk.NewInt(1e6)},
		// 	exactTokenB: validMsg.ExactTokenB,
		// 	slippage:    validMsg.Slippage,
		// 	deadline:    validMsg.Deadline,
		// 	expectedErr: "token a deposit amount 1000000UKAVA: invalid coins",
		// },
		{
			name:        "negative token b",
			requester:   validMsg.Requester,
			tokenA:      validMsg.TokenA,
			exactTokenB: sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(-1)},
			slippage:    validMsg.Slippage,
			deadline:    validMsg.Deadline,
			expectedErr: "exact token b deposit amount -1ukava: invalid coins",
		},
		{
			name:        "zero token b",
			requester:   validMsg.Requester,
			tokenA:      validMsg.TokenA,
			exactTokenB: sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(0)},
			slippage:    validMsg.Slippage,
			deadline:    validMsg.Deadline,
			expectedErr: "exact token b deposit amount 0ukava: invalid coins",
		},
		// TODO:
		// {
		// 	name:        "invalid denom token b",
		// 	requester:   validMsg.Requester,
		// 	tokenA:      validMsg.TokenA,
		// 	exactTokenB: sdk.Coin{Denom: "UKAVA", Amount: sdk.NewInt(1e6)},
		// 	slippage:    validMsg.Slippage,
		// 	deadline:    validMsg.Deadline,
		// 	expectedErr: "exact token b deposit amount 1000000UKAVA: invalid coins",
		// },
		{
			name:        "denoms can not be the same",
			requester:   validMsg.Requester,
			tokenA:      sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(1e6)},
			exactTokenB: sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(1e6)},
			slippage:    validMsg.Slippage,
			deadline:    validMsg.Deadline,
			expectedErr: "denominations can not be equal: invalid coins",
		},
		{
			name:        "zero deadline",
			requester:   validMsg.Requester,
			tokenA:      validMsg.TokenA,
			exactTokenB: validMsg.ExactTokenB,
			slippage:    validMsg.Slippage,
			deadline:    0,
			expectedErr: "deadline 0: invalid deadline",
		},
		{
			name:        "negative deadline",
			requester:   validMsg.Requester,
			tokenA:      validMsg.TokenA,
			exactTokenB: validMsg.ExactTokenB,
			slippage:    validMsg.Slippage,
			deadline:    -1,
			expectedErr: "deadline -1: invalid deadline",
		},
		{
			name:        "negative slippage",
			requester:   validMsg.Requester,
			tokenA:      validMsg.TokenA,
			exactTokenB: validMsg.ExactTokenB,
			slippage:    sdk.MustNewDecFromStr("-0.01"),
			deadline:    validMsg.Deadline,
			expectedErr: "slippage can not be negative: invalid slippage",
		},
		{
			name:        "nil slippage",
			requester:   validMsg.Requester,
			tokenA:      validMsg.TokenA,
			exactTokenB: validMsg.ExactTokenB,
			slippage:    sdk.Dec{},
			deadline:    validMsg.Deadline,
			expectedErr: "slippage must be set: invalid slippage",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgSwapForExactTokens(tc.requester, tc.tokenA, tc.exactTokenB, tc.slippage, tc.deadline)
			err := msg.ValidateBasic()
			assert.EqualError(t, err, tc.expectedErr)
		})
	}
}

func TestMsgSwapForExactTokens_Deadline(t *testing.T) {
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
		msg := types.NewMsgSwapForExactTokens(
			sdk.AccAddress("test1").String(),
			sdk.NewCoin("ukava", sdk.NewInt(1000000)),
			sdk.NewCoin("usdx", sdk.NewInt(2000000)),
			sdk.MustNewDecFromStr("0.01"),
			tc.deadline,
		)
		require.NoError(t, msg.ValidateBasic())
		assert.Equal(t, tc.isExceeded, msg.DeadlineExceeded(blockTime))
		assert.Equal(t, time.Unix(tc.deadline, 0), msg.GetDeadline())
	}
}
