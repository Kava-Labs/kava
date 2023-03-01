package ante_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/app/ante"
)

var _ sdk.AnteHandler = (&MockAnteHandler{}).AnteHandle

type MockAnteHandler struct {
	WasCalled bool
	CalledCtx sdk.Context
}

func (mah *MockAnteHandler) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
	mah.WasCalled = true
	mah.CalledCtx = ctx
	return ctx, nil
}

func mockAddressFetcher(addresses ...sdk.AccAddress) ante.AddressFetcher {
	return func(sdk.Context) []sdk.AccAddress { return addresses }
}

func TestAuthenticatedMempoolDecorator_AnteHandle_NotCheckTx(t *testing.T) {
	txConfig := app.MakeEncodingConfig().TxConfig

	testPrivKeys, testAddresses := app.GeneratePrivKeyAddressPairs(5)
	fetcher := mockAddressFetcher(testAddresses[1:]...)

	decorator := ante.NewAuthenticatedMempoolDecorator(fetcher)
	tx, err := helpers.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		txConfig,
		[]sdk.Msg{
			banktypes.NewMsgSend(
				testAddresses[0],
				testAddresses[1],
				sdk.NewCoins(sdk.NewInt64Coin("ukava", 100_000_000)),
			),
		},
		sdk.NewCoins(), // no fee
		helpers.DefaultGenTxGas,
		"testing-chain-id",
		[]uint64{0},
		[]uint64{0},
		testPrivKeys[0], // address is not authorized
	)
	require.NoError(t, err)
	mmd := MockAnteHandler{}
	ctx := sdk.Context{}.WithIsCheckTx(false) // run as it would be during block update ('DeliverTx'), not just checking entry to mempool

	_, err = decorator.AnteHandle(ctx, tx, false, mmd.AnteHandle)

	require.NoError(t, err)
	require.True(t, mmd.WasCalled)
}

func TestAuthenticatedMempoolDecorator_AnteHandle_Pass(t *testing.T) {
	txConfig := app.MakeEncodingConfig().TxConfig

	testPrivKeys, testAddresses := app.GeneratePrivKeyAddressPairs(5)
	fetcher := mockAddressFetcher(testAddresses[1:]...)

	decorator := ante.NewAuthenticatedMempoolDecorator(fetcher)

	tx, err := helpers.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		txConfig,
		[]sdk.Msg{
			banktypes.NewMsgSend(
				testAddresses[0],
				testAddresses[1],
				sdk.NewCoins(sdk.NewInt64Coin("ukava", 100_000_000)),
			),
			banktypes.NewMsgSend(
				testAddresses[2],
				testAddresses[1],
				sdk.NewCoins(sdk.NewInt64Coin("ukava", 100_000_000)),
			),
		},
		sdk.NewCoins(), // no fee
		helpers.DefaultGenTxGas,
		"testing-chain-id",
		[]uint64{0, 123},
		[]uint64{0, 123},
		testPrivKeys[0], // not in list of authorized addresses
		testPrivKeys[2], // in list of authorized addresses
	)
	require.NoError(t, err)
	mmd := MockAnteHandler{}
	ctx := sdk.Context{}.WithIsCheckTx(true)

	_, err = decorator.AnteHandle(ctx, tx, false, mmd.AnteHandle)

	require.NoError(t, err)
	require.True(t, mmd.WasCalled)
}

func TestAuthenticatedMempoolDecorator_AnteHandle_Reject(t *testing.T) {
	txConfig := app.MakeEncodingConfig().TxConfig

	testPrivKeys, testAddresses := app.GeneratePrivKeyAddressPairs(5)
	fetcher := mockAddressFetcher(testAddresses[1:]...)

	decorator := ante.NewAuthenticatedMempoolDecorator(fetcher)

	tx, err := helpers.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		txConfig,
		[]sdk.Msg{
			banktypes.NewMsgSend(
				testAddresses[0],
				testAddresses[1],
				sdk.NewCoins(sdk.NewInt64Coin("ukava", 100_000_000)),
			),
		},
		sdk.NewCoins(), // no fee
		helpers.DefaultGenTxGas,
		"testing-chain-id",
		[]uint64{0},
		[]uint64{0},
		testPrivKeys[0], // not in list of authorized addresses
	)
	require.NoError(t, err)
	mmd := MockAnteHandler{}
	ctx := sdk.Context{}.WithIsCheckTx(true)

	_, err = decorator.AnteHandle(ctx, tx, false, mmd.AnteHandle)

	require.Error(t, err)
	require.False(t, mmd.WasCalled)
}
