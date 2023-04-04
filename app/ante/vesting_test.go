package ante_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/app/ante"
)

func TestVestingMempoolDecorator_MsgCreateVestingAccount_Unauthorized(t *testing.T) {
	txConfig := app.MakeEncodingConfig().TxConfig

	testPrivKeys, testAddresses := app.GeneratePrivKeyAddressPairs(5)

	decorator := ante.NewVestingAccountDecorator()

	tx, err := helpers.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		txConfig,
		[]sdk.Msg{
			vesting.NewMsgCreateVestingAccount(
				testAddresses[0], testAddresses[1],
				sdk.NewCoins(sdk.NewInt64Coin("ukava", 100_000_000)),
				time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC).Unix(), false),
		},
		sdk.NewCoins(),
		helpers.DefaultGenTxGas,
		"testing-chain-id",
		[]uint64{0},
		[]uint64{0},
		testPrivKeys[0],
	)
	require.NoError(t, err)
	mmd := MockAnteHandler{}
	ctx := sdk.Context{}.WithIsCheckTx(true)
	_, err = decorator.AnteHandle(ctx, tx, false, mmd.AnteHandle)
	require.Error(t, err)
	require.Contains(t, err.Error(), "MsgCreateVestingAccount not supported")
}
