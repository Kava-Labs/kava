package ante_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	ethermint "github.com/tharsis/ethermint/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/app/ante"
)

func TestConvertEthAccounts(t *testing.T) {
	chainID := "kavatest_1-1"
	testPrivKeys, testAddresses := app.GeneratePrivKeyAddressPairs(5)
	encodingConfig := app.MakeEncodingConfig()

	testCases := []struct {
		name            string
		account         authtypes.GenesisAccount
		tx              sdk.Tx
		expectedErr     error
		expectedAccount authtypes.GenesisAccount
	}{
		{
			name:    "missing account fails antehandler, does not create an account",
			account: nil,
			tx: mustGenTx(
				encodingConfig.TxConfig,
				[]sdk.Msg{banktypes.NewMsgSend(testAddresses[0], testAddresses[1], sdk.NewCoins(sdk.NewInt64Coin("ukava", 1)))},
				sdk.NewCoins(),
				helpers.DefaultGenTxGas,
				chainID,
				[]uint64{0},
				[]uint64{0},
				testPrivKeys[0],
			),
			expectedErr:     sdkerrors.ErrUnknownAddress,
			expectedAccount: nil,
		},
		{
			name: "base account is left unchanged",
			account: authtypes.NewBaseAccount(
				testAddresses[0],
				nil,
				0,
				0,
			),
			tx: mustGenTx(
				encodingConfig.TxConfig,
				[]sdk.Msg{banktypes.NewMsgSend(testAddresses[0], testAddresses[1], sdk.NewCoins(sdk.NewInt64Coin("ukava", 1)))},
				sdk.NewCoins(),
				helpers.DefaultGenTxGas,
				chainID,
				[]uint64{0},
				[]uint64{0},
				testPrivKeys[0],
			),
			expectedErr: nil,
			expectedAccount: authtypes.NewBaseAccount(
				testAddresses[0],
				nil,
				0,
				0,
			),
		},
		{
			name: "eth account is converted to base account",
			account: &ethermint.EthAccount{
				BaseAccount: authtypes.NewBaseAccount(
					testAddresses[0],
					nil,
					0,
					0,
				),
				CodeHash: common.BytesToHash(evmtypes.EmptyCodeHash).String(), // code hash includes 0x prefix
			},
			tx: mustGenTx(
				encodingConfig.TxConfig,
				[]sdk.Msg{banktypes.NewMsgSend(testAddresses[0], testAddresses[1], sdk.NewCoins(sdk.NewInt64Coin("ukava", 1)))},
				sdk.NewCoins(),
				helpers.DefaultGenTxGas,
				chainID,
				[]uint64{0},
				[]uint64{0},
				testPrivKeys[0],
			),
			expectedErr: nil,
			expectedAccount: authtypes.NewBaseAccount(
				testAddresses[0],
				nil,
				0,
				0,
			),
		},
		{
			name: "contract eth account is left unchanged",
			account: &ethermint.EthAccount{
				BaseAccount: authtypes.NewBaseAccount(
					testAddresses[0],
					nil,
					0,
					0,
				),
				CodeHash: "0x6cba8c69b5f9084d8eefd5dd7cf71ed5469f5bbb9d8446533ebe4beccdfb3ce9",
			},
			tx: mustGenTx(
				encodingConfig.TxConfig,
				[]sdk.Msg{banktypes.NewMsgSend(testAddresses[0], testAddresses[1], sdk.NewCoins(sdk.NewInt64Coin("ukava", 1)))},
				sdk.NewCoins(),
				helpers.DefaultGenTxGas,
				chainID,
				[]uint64{0},
				[]uint64{0},
				testPrivKeys[0],
			),
			expectedErr: nil,
			expectedAccount: &ethermint.EthAccount{
				BaseAccount: authtypes.NewBaseAccount(
					testAddresses[0],
					nil,
					0,
					0,
				),
				CodeHash: "0x6cba8c69b5f9084d8eefd5dd7cf71ed5469f5bbb9d8446533ebe4beccdfb3ce9",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tApp := app.NewTestApp()

			authGenState := app.NewAuthBankGenesisBuilder()
			if tc.account != nil {
				authGenState = authGenState.WithAccounts(tc.account)
			}

			tApp = tApp.InitializeFromGenesisStatesWithTimeAndChainID(
				time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
				chainID,
				authGenState.BuildMarshalled(encodingConfig.Marshaler),
			)

			handler := ante.NewConvertEthAccounts(tApp.GetAccountKeeper())

			mah := MockAnteHandler{}

			ctx := tApp.NewContext(false, tmproto.Header{Height: 1})

			_, err := handler.AnteHandle(ctx, tc.tx, false, mah.AnteHandle)
			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
				require.False(t, mah.WasCalled)
			} else {
				require.NoError(t, err)
				require.True(t, mah.WasCalled)
			}

			actualAcc := tApp.GetAccountKeeper().GetAccount(ctx, testAddresses[0])
			require.Equal(t, tc.expectedAccount, actualAcc)
		})
	}
}

func mustGenTx(gen client.TxConfig, msgs []sdk.Msg, feeAmt sdk.Coins, gas uint64, chainID string, accNums, accSeqs []uint64, priv ...cryptotypes.PrivKey) sdk.Tx {
	tx, err := helpers.GenTx(gen, msgs, feeAmt, gas, chainID, accNums, accSeqs, priv...)
	if err != nil {
		panic(fmt.Sprintf("failed to generate tx: %v", err))
	}
	return tx
}
