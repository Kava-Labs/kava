package ante

import (
	"math/rand"
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/kava-labs/kava/x/bep3"
)

var (
	_ sdk.AnteHandler = (&MockAnteHandler{}).AnteHandle
)

type MockAnteHandler struct {
	WasCalled bool
}

func (mah *MockAnteHandler) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
	mah.WasCalled = true
	return ctx, nil
}

func mockAddressFetcher(addresses ...sdk.AccAddress) AddressFetcher {
	return func(sdk.Context) []sdk.AccAddress { return addresses }
}

func TestAuthenticatedMempoolDecorator_AnteHandle_NotCheckTx(t *testing.T) {
	testPrivKeys, testAddresses := generatePrivKeyAddressPairs(5)
	fetcher := mockAddressFetcher(testAddresses[1:]...)

	decorator := NewAuthenticatedMempoolDecorator(fetcher)
	tx := helpers.GenTx(
		[]sdk.Msg{
			bep3.NewMsgClaimAtomicSwap(
				testAddresses[0],
				[]byte{},
				[]byte{},
			),
		},
		sdk.NewCoins(), // no fee
		helpers.DefaultGenTxGas,
		"testing-chain-id",
		[]uint64{0},
		[]uint64{0},
		testPrivKeys[0], // address is not authorized
	)
	mmd := MockAnteHandler{}
	ctx := sdk.Context{}.WithIsCheckTx(false) // run as it would be during block update ('DeliverTx'), not just checking entry to mempool

	_, err := decorator.AnteHandle(ctx, tx, false, mmd.AnteHandle)

	require.NoError(t, err)
	require.True(t, mmd.WasCalled)
}

func TestAuthenticatedMempoolDecorator_AnteHandle_Pass(t *testing.T) {
	testPrivKeys, testAddresses := generatePrivKeyAddressPairs(5)
	fetcher := mockAddressFetcher(testAddresses[1:]...)

	decorator := NewAuthenticatedMempoolDecorator(fetcher)

	tx := helpers.GenTx(
		[]sdk.Msg{
			bank.NewMsgSend(
				testAddresses[0],
				testAddresses[1],
				sdk.NewCoins(sdk.NewInt64Coin("ukava", 100_000_000)),
			),
			bep3.NewMsgClaimAtomicSwap(
				testAddresses[2],
				nil,
				nil,
			),
		},
		sdk.NewCoins(), // no fee
		helpers.DefaultGenTxGas,
		"testing-chain-id",
		[]uint64{0, 123},
		[]uint64{0, 123},
		testPrivKeys[0], // not in list of authorized addresses
		testPrivKeys[2],
	)
	mmd := MockAnteHandler{}
	ctx := sdk.Context{}.WithIsCheckTx(true)

	_, err := decorator.AnteHandle(ctx, tx, false, mmd.AnteHandle)

	require.NoError(t, err)
	require.True(t, mmd.WasCalled)
}

func TestAuthenticatedMempoolDecorator_AnteHandle_Reject(t *testing.T) {
	testPrivKeys, testAddresses := generatePrivKeyAddressPairs(5)
	fetcher := mockAddressFetcher(testAddresses[1:]...)

	decorator := NewAuthenticatedMempoolDecorator(fetcher)

	tx := helpers.GenTx(
		[]sdk.Msg{
			bank.NewMsgSend(
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
	mmd := MockAnteHandler{}
	ctx := sdk.Context{}.WithIsCheckTx(true)

	_, err := decorator.AnteHandle(ctx, tx, false, mmd.AnteHandle)

	require.Error(t, err)
	require.False(t, mmd.WasCalled)
}

// generatePrivKeyAddressPairsFromRand generates (deterministically) a total of n private keys and addresses.
func generatePrivKeyAddressPairs(n int) (keys []crypto.PrivKey, addrs []sdk.AccAddress) {
	r := rand.New(rand.NewSource(12345)) // make the generation deterministic
	keys = make([]crypto.PrivKey, n)
	addrs = make([]sdk.AccAddress, n)
	for i := 0; i < n; i++ {
		secret := make([]byte, 32)
		_, err := r.Read(secret)
		if err != nil {
			panic("Could not read randomness")
		}
		keys[i] = secp256k1.GenPrivKeySecp256k1(secret)
		addrs[i] = sdk.AccAddress(keys[i].PubKey().Address())
	}
	return
}
