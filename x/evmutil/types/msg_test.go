package types_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/evmutil/types"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func TestMsgConvertCoinToERC20(t *testing.T) {
	app.SetSDKConfig()

	type errArgs struct {
		expectPass bool
		contains   string
	}

	tests := []struct {
		name          string
		giveInitiator string
		giveReceiver  string
		giveAmount    sdk.Coin
		errArgs       errArgs
	}{
		{
			"valid",
			"kava123fxg0l602etulhhcdm0vt7l57qya5wjcrwhzz",
			"0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
			sdk.NewCoin("erc20/weth", sdk.NewInt(1234)),
			errArgs{
				expectPass: true,
			},
		},
		{
			"invalid - odd length hex address",
			"kava123fxg0l602etulhhcdm0vt7l57qya5wjcrwhzz",
			"0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc",
			sdk.NewCoin("erc20/weth", sdk.NewInt(1234)),
			errArgs{
				expectPass: false,
				contains:   "Receiver is not a valid hex address: invalid address",
			},
		},
		{
			"invalid - zero amount",
			"kava123fxg0l602etulhhcdm0vt7l57qya5wjcrwhzz",
			"0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
			sdk.NewCoin("erc20/weth", sdk.NewInt(0)),
			errArgs{
				expectPass: false,
				contains:   "amount cannot be zero",
			},
		},
		{
			"invalid - negative amount",
			"kava123fxg0l602etulhhcdm0vt7l57qya5wjcrwhzz",
			"0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
			// Create manually so there is no validation
			sdk.Coin{Denom: "erc20/weth", Amount: sdk.NewInt(-1234)},
			errArgs{
				expectPass: false,
				contains:   "negative coin amount",
			},
		},
		{
			"invalid - empty denom",
			"kava123fxg0l602etulhhcdm0vt7l57qya5wjcrwhzz",
			"0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
			sdk.Coin{Denom: "", Amount: sdk.NewInt(-1234)},
			errArgs{
				expectPass: false,
				contains:   "invalid denom",
			},
		},
		{
			"invalid - invalid denom",
			"kava123fxg0l602etulhhcdm0vt7l57qya5wjcrwhzz",
			"0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
			sdk.Coin{Denom: "h", Amount: sdk.NewInt(-1234)},
			errArgs{
				expectPass: false,
				contains:   "invalid denom",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.NewMsgConvertCoinToERC20(
				tc.giveInitiator,
				tc.giveReceiver,
				tc.giveAmount,
			)
			err := msg.ValidateBasic()

			if tc.errArgs.expectPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errArgs.contains)
			}
		})
	}
}

func TestMsgConvertERC20ToCoin(t *testing.T) {
	app.SetSDKConfig()

	type errArgs struct {
		expectPass bool
		contains   string
	}

	tests := []struct {
		name         string
		receiver     string
		initiator    string
		contractAddr string
		amount       sdk.Int
		errArgs      errArgs
	}{
		{
			"valid",
			"kava123fxg0l602etulhhcdm0vt7l57qya5wjcrwhzz",
			"0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
			"0x404F9466d758eA33eA84CeBE9E444b06533b369e",
			sdk.NewInt(1234),
			errArgs{
				expectPass: true,
			},
		},
		{
			"invalid - odd length hex address",
			"kava123fxg0l602etulhhcdm0vt7l57qya5wjcrwhzz",
			"0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc",
			"0x404F9466d758eA33eA84CeBE9E444b06533b369e",
			sdk.NewInt(1234),
			errArgs{
				expectPass: false,
				contains:   "initiator is not a valid hex address",
			},
		},
		{
			"invalid - zero amount",
			"kava123fxg0l602etulhhcdm0vt7l57qya5wjcrwhzz",
			"0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
			"0x404F9466d758eA33eA84CeBE9E444b06533b369e",
			sdk.NewInt(0),
			errArgs{
				expectPass: false,
				contains:   "amount cannot be zero",
			},
		},
		{
			"invalid - negative amount",
			"kava123fxg0l602etulhhcdm0vt7l57qya5wjcrwhzz",
			"0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
			"0x404F9466d758eA33eA84CeBE9E444b06533b369e",
			sdk.NewInt(-1234),
			errArgs{
				expectPass: false,
				contains:   "amount cannot be zero or less",
			},
		},
		{
			"invalid - invalid contract address",
			"kava123fxg0l602etulhhcdm0vt7l57qya5wjcrwhzz",
			"0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
			"0x404F9466d758eA33eA84CeBE9E444b06533b369",
			sdk.NewInt(1234),
			errArgs{
				expectPass: false,
				contains:   "erc20 contract address is not a valid hex address",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.MsgConvertERC20ToCoin{
				Initiator:        tc.initiator,
				Receiver:         tc.receiver,
				KavaERC20Address: tc.contractAddr,
				Amount:           tc.amount,
			}
			err := msg.ValidateBasic()

			if tc.errArgs.expectPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errArgs.contains)
			}
		})
	}
}

func TestMsgEVMCall_ValidateAndDecode(t *testing.T) {
	app.SetSDKConfig()

	type errArgs struct {
		expectPass bool
		contains   string
	}

	contractAddr := "0x15932E26f5BD4923d46a2b205191C4b5d5f43FE3"
	authority := "kava1cj7njkw2g9fqx4e768zc75dp9sks8u9znxrf0w"
	validParams := make([]interface{}, 2)
	validAddr := common.HexToAddress("0x71586E5B3468B5720BAa9162A02366Fae6933BfE")
	validAmt := int64(10)
	validParams[0] = validAddr
	validParams[1] = sdk.NewInt(validAmt).BigInt()
	validFnAbi := `{
		"inputs": [
			{ "type": "address", "name": "to" },
			{ "type": "uint256", "name": "amount" }
		],
		"name": "transfer",
		"type": "function"
	}`
	validData := encodeTransferFn(validAddr, validAmt)

	tests := []struct {
		name    string
		msg     types.MsgEVMCall
		params  []interface{}
		errArgs errArgs
	}{
		{
			name: "valid contract call",
			msg: types.MsgEVMCall{
				To:        contractAddr,
				FnAbi:     validFnAbi,
				Data:      validData,
				Amount:    sdk.ZeroInt(),
				Authority: authority,
			},
			params: validParams,
			errArgs: errArgs{
				expectPass: true,
			},
		},
		{
			name: "valid non-contract call",
			msg: types.MsgEVMCall{
				To:        contractAddr,
				Amount:    sdk.OneInt(),
				Authority: authority,
			},
			errArgs: errArgs{
				expectPass: true,
			},
		},
		{
			name: "invalid - data with no abi",
			msg: types.MsgEVMCall{
				To:        contractAddr,
				Data:      validData,
				Amount:    sdk.ZeroInt(),
				Authority: authority,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "fnAbi is not provided: this required when passing in data",
			},
		},
		{
			name: "invalid - 'to' not hex address",
			msg: types.MsgEVMCall{
				To:        authority,
				Amount:    sdk.ZeroInt(),
				Authority: authority,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "to 'kava1cj7njkw2g9fqx4e768zc75dp9sks8u9znxrf0w' is not hex address",
			},
		},
		{
			name: "invalid - invalid authority",
			msg: types.MsgEVMCall{
				To:        contractAddr,
				Amount:    sdk.ZeroInt(),
				Authority: "",
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "invalid authority address",
			},
		},
		{
			name: "invalid - negative amount",
			msg: types.MsgEVMCall{
				To:        contractAddr,
				Amount:    sdk.NewInt(-10),
				Authority: authority,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "amount cannot be negative: -10",
			},
		},
		{
			name: "invalid - cannot decode data with fnAbi",
			msg: types.MsgEVMCall{
				To:        contractAddr,
				FnAbi:     validFnAbi,
				Data:      "0xa9059cbb00",
				Amount:    sdk.ZeroInt(),
				Authority: authority,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "unable to decode method args: abi",
			},
		},
		{
			name: "invalid - fnAbi does not match params",
			msg: types.MsgEVMCall{
				To: contractAddr,
				// transfer params is reversed
				FnAbi: `{
					"inputs": [
						{ "type": "uint256", "name": "amount" },
						{ "type": "address", "name": "to" }
					],
					"name": "transfer",
					"type": "function"
				}`,
				Data:      validData,
				Amount:    sdk.ZeroInt(),
				Authority: authority,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "method not found in fnAbi: 0xa9059cbb",
			},
		},
		{
			name: "invalid - data fn signature mismatch",
			msg: types.MsgEVMCall{
				To:    contractAddr,
				FnAbi: validFnAbi,
				Data: fmt.Sprintf(
					"0x%s%s%s",
					"5d359fbd", // transfer(address,uint64)
					encodeAddress(validAddr),
					encodeInt(validAmt),
				),
				Amount:    sdk.ZeroInt(),
				Authority: authority,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "method not found in fnAbi: 0x5d359fbd",
			},
		},
		{
			name: "invalid - fnAbi invalid",
			msg: types.MsgEVMCall{
				To:        contractAddr,
				FnAbi:     "100",
				Data:      validData,
				Amount:    sdk.ZeroInt(),
				Authority: authority,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "unable to parse fn abi",
			},
		},
		{
			name: "invalid - data must be hex string",
			msg: types.MsgEVMCall{
				To:        contractAddr,
				FnAbi:     validFnAbi,
				Data:      "100",
				Amount:    sdk.ZeroInt(),
				Authority: authority,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "invalid data format",
			},
		},
		{
			name: "invalid - no amount",
			msg: types.MsgEVMCall{
				To:        contractAddr,
				Authority: authority,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "amount must not be nil",
			},
		},
		{
			name: "invalid - extra param data is passed",
			msg: types.MsgEVMCall{
				To:        contractAddr,
				FnAbi:     validFnAbi,
				Data:      fmt.Sprintf("%s%s", validData, encodeInt(100)),
				Amount:    sdk.ZeroInt(),
				Authority: authority,
			},
			params: validParams,
			errArgs: errArgs{
				expectPass: false,
				contains:   "invalid call data: call data does not match unpacked data",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()

			if tc.errArgs.expectPass {
				require.NoError(t, err)

				if len(tc.msg.Data) > 0 {
					decoded, err := tc.msg.Decode()
					require.NoError(t, err)
					require.Equal(t, tc.params, decoded)
				}
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errArgs.contains)
			}
		})
	}
}

func encodeTransferFn(addr common.Address, amt int64) string {
	return fmt.Sprintf(
		"0x%s%s%s",
		"a9059cbb", // transfer(address,uint256)
		encodeAddress(addr),
		encodeInt(amt),
	)
}

func encodeAddress(addr common.Address) string {
	return hexutil.Encode(common.LeftPadBytes(addr.Bytes(), 32))[2:]
}

func encodeInt(amt int64) string {
	return hexutil.Encode(common.LeftPadBytes(big.NewInt(amt).Bytes(), 32))[2:]
}
