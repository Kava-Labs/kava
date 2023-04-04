package types_test

import (
	"testing"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/evmutil/types"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
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
			sdk.NewCoin("erc20/weth", sdkmath.NewInt(1234)),
			errArgs{
				expectPass: true,
			},
		},
		{
			"invalid - odd length hex address",
			"kava123fxg0l602etulhhcdm0vt7l57qya5wjcrwhzz",
			"0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc",
			sdk.NewCoin("erc20/weth", sdkmath.NewInt(1234)),
			errArgs{
				expectPass: false,
				contains:   "Receiver is not a valid hex address: invalid address",
			},
		},
		{
			"invalid - zero amount",
			"kava123fxg0l602etulhhcdm0vt7l57qya5wjcrwhzz",
			"0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
			sdk.NewCoin("erc20/weth", sdkmath.NewInt(0)),
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
			sdk.Coin{Denom: "erc20/weth", Amount: sdkmath.NewInt(-1234)},
			errArgs{
				expectPass: false,
				contains:   "negative coin amount",
			},
		},
		{
			"invalid - empty denom",
			"kava123fxg0l602etulhhcdm0vt7l57qya5wjcrwhzz",
			"0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
			sdk.Coin{Denom: "", Amount: sdkmath.NewInt(-1234)},
			errArgs{
				expectPass: false,
				contains:   "invalid denom",
			},
		},
		{
			"invalid - invalid denom",
			"kava123fxg0l602etulhhcdm0vt7l57qya5wjcrwhzz",
			"0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
			sdk.Coin{Denom: "h", Amount: sdkmath.NewInt(-1234)},
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
		amount       sdkmath.Int
		errArgs      errArgs
	}{
		{
			"valid",
			"kava123fxg0l602etulhhcdm0vt7l57qya5wjcrwhzz",
			"0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
			"0x404F9466d758eA33eA84CeBE9E444b06533b369e",
			sdkmath.NewInt(1234),
			errArgs{
				expectPass: true,
			},
		},
		{
			"invalid - odd length hex address",
			"kava123fxg0l602etulhhcdm0vt7l57qya5wjcrwhzz",
			"0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc",
			"0x404F9466d758eA33eA84CeBE9E444b06533b369e",
			sdkmath.NewInt(1234),
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
			sdkmath.NewInt(0),
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
			sdkmath.NewInt(-1234),
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
			sdkmath.NewInt(1234),
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
