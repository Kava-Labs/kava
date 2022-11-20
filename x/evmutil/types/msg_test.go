package types_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/evmutil/types"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
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

	validFnAbi := `{
		"inputs": [
			{ "type": "address", "name": "to" },
			{ "type": "uint256", "name": "amount" }
		],
		"name": "transfer",
		"type": "function"
	}`
	validData := "0xa9059cbb00000000000000000000000071586e5b3468b5720baa9162a02366fae6933bfe000000000000000000000000000000000000000000000000000000000000000a"

	validParams := make([]interface{}, 2)
	validParams[0] = common.HexToAddress("0x71586E5B3468B5720BAa9162A02366Fae6933BfE")
	validParams[1] = sdk.NewInt(10).BigInt()

	tests := []struct {
		name    string
		msg     types.MsgEVMCall
		params  []interface{}
		errArgs errArgs
	}{
		{
			name: "valid contract call",
			msg: types.MsgEVMCall{
				To:        "0x15932E26f5BD4923d46a2b205191C4b5d5f43FE3",
				FnAbi:     validFnAbi,
				Data:      validData,
				Amount:    sdk.ZeroInt(),
				Authority: "kava1cj7njkw2g9fqx4e768zc75dp9sks8u9znxrf0w",
			},
			params: validParams,
			errArgs: errArgs{
				expectPass: true,
			},
		},
		{
			name: "valid non-contract call",
			msg: types.MsgEVMCall{
				To:        "0x15932E26f5BD4923d46a2b205191C4b5d5f43FE3",
				Amount:    sdk.OneInt(),
				Authority: "kava1cj7njkw2g9fqx4e768zc75dp9sks8u9znxrf0w",
			},
			errArgs: errArgs{
				expectPass: true,
			},
		},
		{
			name: "invalid - data with no abi",
			msg: types.MsgEVMCall{
				To:        "0x15932E26f5BD4923d46a2b205191C4b5d5f43FE3",
				Data:      validData,
				Amount:    sdk.ZeroInt(),
				Authority: "kava1cj7njkw2g9fqx4e768zc75dp9sks8u9znxrf0w",
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "fnAbi is not provided: this required when passing in data",
			},
		},
		{
			name: "invalid - 'to' not hex address",
			msg: types.MsgEVMCall{
				To:        "kava1cj7njkw2g9fqx4e768zc75dp9sks8u9znxrf0w",
				Amount:    sdk.ZeroInt(),
				Authority: "kava1cj7njkw2g9fqx4e768zc75dp9sks8u9znxrf0w",
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "to 'kava1cj7njkw2g9fqx4e768zc75dp9sks8u9znxrf0w' is not hex address",
			},
		},
		{
			name: "invalid - invalid authority",
			msg: types.MsgEVMCall{
				To:        "0x15932E26f5BD4923d46a2b205191C4b5d5f43FE3",
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
				To:        "0x15932E26f5BD4923d46a2b205191C4b5d5f43FE3",
				Amount:    sdk.NewInt(-10),
				Authority: "kava1cj7njkw2g9fqx4e768zc75dp9sks8u9znxrf0w",
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "amount cannot be negative: -10",
			},
		},
		{
			name: "invalid - cannot decode data with fnAbi",
			msg: types.MsgEVMCall{
				To:        "0x15932E26f5BD4923d46a2b205191C4b5d5f43FE3",
				FnAbi:     validFnAbi,
				Data:      "0xa9059cbb00",
				Amount:    sdk.ZeroInt(),
				Authority: "kava1cj7njkw2g9fqx4e768zc75dp9sks8u9znxrf0w",
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "unable to decode data",
			},
		},
		{
			name: "invalid - fnAbi does not match params",
			msg: types.MsgEVMCall{
				To: "0x15932E26f5BD4923d46a2b205191C4b5d5f43FE3",
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
				Authority: "kava1cj7njkw2g9fqx4e768zc75dp9sks8u9znxrf0w",
			},
			errArgs: errArgs{
				expectPass: false,
			},
		},
		{
			name: "invalid - data fn signature mismatch",
			msg: types.MsgEVMCall{
				To:        "0x15932E26f5BD4923d46a2b205191C4b5d5f43FE3",
				FnAbi:     validFnAbi,
				Data:      "0xa9059cab00000000000000000000000071586e5b3468b5720baa9162a02366fae6933bfe000000000000000000000000000000000000000000000000000000000000000a",
				Amount:    sdk.ZeroInt(),
				Authority: "kava1cj7njkw2g9fqx4e768zc75dp9sks8u9znxrf0w",
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "failed to validate fn signature `transfer(address,uint256)` with data `0xa9059cab`",
			},
		},
		{
			name: "invalid - fnAbi invalid",
			msg: types.MsgEVMCall{
				To:        "0x15932E26f5BD4923d46a2b205191C4b5d5f43FE3",
				FnAbi:     "100",
				Data:      validData,
				Amount:    sdk.ZeroInt(),
				Authority: "kava1cj7njkw2g9fqx4e768zc75dp9sks8u9znxrf0w",
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "unable to parse fn abi",
			},
		},
		{
			name: "invalid - data must be hex string",
			msg: types.MsgEVMCall{
				To:        "0x15932E26f5BD4923d46a2b205191C4b5d5f43FE3",
				FnAbi:     validFnAbi,
				Data:      "100",
				Amount:    sdk.ZeroInt(),
				Authority: "kava1cj7njkw2g9fqx4e768zc75dp9sks8u9znxrf0w",
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "invalid data format",
			},
		},
		{
			name: "invalid - no amount",
			msg: types.MsgEVMCall{
				To:        "0x15932E26f5BD4923d46a2b205191C4b5d5f43FE3",
				Authority: "kava1cj7njkw2g9fqx4e768zc75dp9sks8u9znxrf0w",
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "amount must not be nil",
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
