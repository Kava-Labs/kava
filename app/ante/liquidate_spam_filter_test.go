package ante_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/app/ante"
	hardtypes "github.com/kava-labs/kava/x/hard/types"
)

func TestHardLiquidateSpamFilter(t *testing.T) {
	testPrivKeys, testAddresses := app.GeneratePrivKeyAddressPairs(2)

	nonLiquidateMsg := banktypes.NewMsgSend(testAddresses[0], testAddresses[1], sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000000))))
	liquidateMsg := hardtypes.NewMsgLiquidate(testAddresses[0], testAddresses[1])

	gasLimit := uint64(1000000)
	decorator := ante.NewHardLiquidateSpamFilter(gasLimit)

	testCases := []struct {
		name        string
		msgs        []sdk.Msg
		checkTx     bool
		gas         uint64
		expectedErr error
	}{
		{
			"ignored if not checktx",
			[]sdk.Msg{&liquidateMsg},
			false,
			gasLimit * 2,
			nil,
		},
		{
			"ignored if not checktx without liquidate",
			[]sdk.Msg{nonLiquidateMsg},
			false,
			gasLimit * 2,
			nil,
		},
		{
			"errors if checktx and over limit",
			[]sdk.Msg{&liquidateMsg},
			true,
			gasLimit * 2,
			sdkerrors.ErrUnauthorized,
		},
		{
			"errors if liquidate is not the first message",
			[]sdk.Msg{nonLiquidateMsg, &liquidateMsg},
			true,
			gasLimit * 2,
			sdkerrors.ErrUnauthorized,
		},
		{
			"allows high gas without a hard liquidate message",
			[]sdk.Msg{nonLiquidateMsg},
			true,
			gasLimit * 2,
			nil,
		},
		{
			"allows liquidate message equal gas to limit with multiple messages",
			[]sdk.Msg{nonLiquidateMsg, &liquidateMsg},
			true,
			gasLimit,
			nil,
		},
		{
			"allows liquidate message with equal gas to limit with single message",
			[]sdk.Msg{&liquidateMsg},
			true,
			gasLimit,
			nil,
		},
		{
			"errors with 1 over the limit and different messages",
			[]sdk.Msg{&liquidateMsg, nonLiquidateMsg, &liquidateMsg, nonLiquidateMsg},
			true,
			gasLimit + 1,
			sdkerrors.ErrUnauthorized,
		},
	}

	txConfig := app.MakeEncodingConfig().TxConfig

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tx, err := helpers.GenTx(
				txConfig,
				tc.msgs,
				sdk.NewCoins(),
				tc.gas,
				"testing-chain-id",
				[]uint64{0},
				[]uint64{0},
				testPrivKeys[0],
			)
			require.NoError(t, err)
			mmd := MockAnteHandler{}
			ctx := sdk.Context{}.WithIsCheckTx(tc.checkTx)
			_, err = decorator.AnteHandle(ctx, tx, false, mmd.AnteHandle)
			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
