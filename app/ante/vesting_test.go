package ante_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/app/ante"
)

func TestVestingMempoolDecorator_MsgCreateVestingAccount_Unauthorized(t *testing.T) {
	txConfig := app.MakeEncodingConfig().TxConfig

	testPrivKeys, testAddresses := app.GeneratePrivKeyAddressPairs(5)

	decorator := ante.NewVestingAccountDecorator()

	tests := []struct {
		name       string
		msg        sdk.Msg
		wantHasErr bool
		wantErr    string
	}{
		{
			"MsgCreateVestingAccount",
			vesting.NewMsgCreateVestingAccount(
				testAddresses[0], testAddresses[1],
				sdk.NewCoins(sdk.NewInt64Coin("ukava", 100_000_000)),
				time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
				false,
			),
			true,
			"MsgTypeURL /cosmos.vesting.v1beta1.MsgCreateVestingAccount not supported",
		},
		{
			"MsgCreateVestingAccount",
			vesting.NewMsgCreatePermanentLockedAccount(
				testAddresses[0], testAddresses[1],
				sdk.NewCoins(sdk.NewInt64Coin("ukava", 100_000_000)),
			),
			true,
			"MsgTypeURL /cosmos.vesting.v1beta1.MsgCreatePermanentLockedAccount not supported",
		},
		{
			"MsgCreateVestingAccount",
			vesting.NewMsgCreatePeriodicVestingAccount(
				testAddresses[0], testAddresses[1],
				time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
				nil,
			),
			true,
			"MsgTypeURL /cosmos.vesting.v1beta1.MsgCreatePeriodicVestingAccount not supported",
		},
		{
			"other messages not affected",
			banktypes.NewMsgSend(
				testAddresses[0], testAddresses[1],
				sdk.NewCoins(sdk.NewInt64Coin("ukava", 100_000_000)),
			),
			false,
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx, err := sims.GenSignedMockTx(
				rand.New(rand.NewSource(time.Now().UnixNano())),
				txConfig,
				[]sdk.Msg{
					tt.msg,
				},
				sdk.NewCoins(),
				sims.DefaultGenTxGas,
				"testing-chain-id",
				[]uint64{0},
				[]uint64{0},
				testPrivKeys[0],
			)
			require.NoError(t, err)

			mmd := MockAnteHandler{}
			ctx := sdk.Context{}.WithIsCheckTx(true)

			_, err = decorator.AnteHandle(ctx, tx, false, mmd.AnteHandle)

			if tt.wantHasErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
