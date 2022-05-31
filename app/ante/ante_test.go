package ante_test

import (
	"os"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	authz "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
	ethermint "github.com/tharsis/ethermint/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"

	"github.com/kava-labs/kava/app"
	bep3types "github.com/kava-labs/kava/x/bep3/types"
	pricefeedtypes "github.com/kava-labs/kava/x/pricefeed/types"
)

func TestMain(m *testing.M) {
	app.SetSDKConfig()
	os.Exit(m.Run())
}

func TestAppAnteHandler_AuthorizedMempool(t *testing.T) {
	testPrivKeys, testAddresses := app.GeneratePrivKeyAddressPairs(10)
	unauthed := testAddresses[0:2]
	unauthedKeys := testPrivKeys[0:2]
	deputy := testAddresses[2]
	deputyKey := testPrivKeys[2]
	oracles := testAddresses[3:6]
	oraclesKeys := testPrivKeys[3:6]
	manual := testAddresses[6:]
	manualKeys := testPrivKeys[6:]

	encodingConfig := app.MakeEncodingConfig()

	opts := app.DefaultOptions
	opts.MempoolEnableAuth = true
	opts.MempoolAuthAddresses = manual

	tApp := app.TestApp{
		App: *app.NewApp(
			log.NewNopLogger(),
			tmdb.NewMemDB(),
			app.DefaultNodeHome,
			nil,
			encodingConfig,
			opts,
		),
	}

	chainID := "kavatest_1-1"
	tApp = tApp.InitializeFromGenesisStatesWithTimeAndChainID(
		time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
		chainID,
		app.NewFundedGenStateWithSameCoins(
			tApp.AppCodec(),
			sdk.NewCoins(sdk.NewInt64Coin("ukava", 1e9)),
			testAddresses,
		),
		newBep3GenStateMulti(tApp.AppCodec(), deputy),
		newPricefeedGenStateMulti(tApp.AppCodec(), oracles),
	)

	testcases := []struct {
		name       string
		address    sdk.AccAddress
		privKey    cryptotypes.PrivKey
		expectPass bool
	}{
		{
			name:       "unauthorized",
			address:    unauthed[1],
			privKey:    unauthedKeys[1],
			expectPass: false,
		},
		{
			name:       "oracle",
			address:    oracles[1],
			privKey:    oraclesKeys[1],
			expectPass: true,
		},
		{
			name:       "deputy",
			address:    deputy,
			privKey:    deputyKey,
			expectPass: true,
		},
		{
			name:       "manual",
			address:    manual[1],
			privKey:    manualKeys[1],
			expectPass: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			stdTx, err := helpers.GenTx(
				encodingConfig.TxConfig,
				[]sdk.Msg{
					banktypes.NewMsgSend(
						tc.address,
						testAddresses[0],
						sdk.NewCoins(sdk.NewInt64Coin("ukava", 1_000_000)),
					),
				},
				sdk.NewCoins(), // no fee
				helpers.DefaultGenTxGas,
				chainID,
				[]uint64{0},
				[]uint64{0}, // fixed sequence numbers will cause tests to fail sig verification if the same address is used twice
				tc.privKey,
			)
			require.NoError(t, err)
			txBytes, err := encodingConfig.TxConfig.TxEncoder()(stdTx)
			require.NoError(t, err)

			res := tApp.CheckTx(
				abci.RequestCheckTx{
					Tx:   txBytes,
					Type: abci.CheckTxType_New,
				},
			)

			if tc.expectPass {
				require.Zero(t, res.Code, res.Log)
			} else {
				require.NotZero(t, res.Code)
			}
		})
	}
}

func newPricefeedGenStateMulti(cdc codec.JSONCodec, oracles []sdk.AccAddress) app.GenesisState {
	pfGenesis := pricefeedtypes.GenesisState{
		Params: pricefeedtypes.Params{
			Markets: []pricefeedtypes.Market{
				{MarketID: "btc:usd", BaseAsset: "btc", QuoteAsset: "usd", Oracles: oracles, Active: true},
			},
		},
	}
	return app.GenesisState{pricefeedtypes.ModuleName: cdc.MustMarshalJSON(&pfGenesis)}
}

func newBep3GenStateMulti(cdc codec.JSONCodec, deputyAddress sdk.AccAddress) app.GenesisState {
	bep3Genesis := bep3types.GenesisState{
		Params: bep3types.Params{
			AssetParams: bep3types.AssetParams{
				bep3types.AssetParam{
					Denom:  "bnb",
					CoinID: 714,
					SupplyLimit: bep3types.SupplyLimit{
						Limit:          sdk.NewInt(350000000000000),
						TimeLimited:    false,
						TimeBasedLimit: sdk.ZeroInt(),
						TimePeriod:     time.Hour,
					},
					Active:        true,
					DeputyAddress: deputyAddress,
					FixedFee:      sdk.NewInt(1000),
					MinSwapAmount: sdk.OneInt(),
					MaxSwapAmount: sdk.NewInt(1000000000000),
					MinBlockLock:  bep3types.DefaultMinBlockLock,
					MaxBlockLock:  bep3types.DefaultMaxBlockLock,
				},
			},
		},
		Supplies: bep3types.AssetSupplies{
			bep3types.NewAssetSupply(
				sdk.NewCoin("bnb", sdk.ZeroInt()),
				sdk.NewCoin("bnb", sdk.ZeroInt()),
				sdk.NewCoin("bnb", sdk.ZeroInt()),
				sdk.NewCoin("bnb", sdk.ZeroInt()),
				time.Duration(0),
			),
		},
		PreviousBlockTime: bep3types.DefaultPreviousBlockTime,
	}
	return app.GenesisState{bep3types.ModuleName: cdc.MustMarshalJSON(&bep3Genesis)}
}

func TestAppAnteHandler_RejectMsgsInAuthz(t *testing.T) {
	testPrivKeys, testAddresses := app.GeneratePrivKeyAddressPairs(10)

	newMsgGrant := func(msgTypeUrl string) *authz.MsgGrant {
		msg, err := authz.NewMsgGrant(
			testAddresses[0],
			testAddresses[1],
			authz.NewGenericAuthorization(msgTypeUrl),
			time.Date(9000, 1, 1, 0, 0, 0, 0, time.UTC),
		)
		if err != nil {
			panic(err)
		}
		return msg
	}

	chainID := "kavatest_1-1"
	encodingConfig := app.MakeEncodingConfig()

	testcases := []struct {
		name         string
		msg          sdk.Msg
		expectedCode uint32
	}{
		{
			name:         "MsgEthereumTx is blocked",
			msg:          newMsgGrant(sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{})),
			expectedCode: sdkerrors.ErrUnauthorized.ABCICode(),
		},
		{
			name:         "MsgCreateVestingAccount is blocked",
			msg:          newMsgGrant(sdk.MsgTypeURL(&vestingtypes.MsgCreateVestingAccount{})),
			expectedCode: sdkerrors.ErrUnauthorized.ABCICode(),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tApp := app.NewTestApp()

			tApp = tApp.InitializeFromGenesisStatesWithTimeAndChainID(
				time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
				chainID,
			)

			stdTx, err := helpers.GenTx(
				encodingConfig.TxConfig,
				[]sdk.Msg{tc.msg},
				sdk.NewCoins(), // no fee
				helpers.DefaultGenTxGas,
				chainID,
				[]uint64{0},
				[]uint64{0},
				testPrivKeys[0],
			)
			require.NoError(t, err)
			txBytes, err := encodingConfig.TxConfig.TxEncoder()(stdTx)
			require.NoError(t, err)

			resCheckTx := tApp.CheckTx(
				abci.RequestCheckTx{
					Tx:   txBytes,
					Type: abci.CheckTxType_New,
				},
			)
			require.Equal(t, resCheckTx.Code, tc.expectedCode, resCheckTx.Log)

			resDeliverTx := tApp.DeliverTx(
				abci.RequestDeliverTx{
					Tx: txBytes,
				},
			)
			require.Equal(t, resDeliverTx.Code, tc.expectedCode, resDeliverTx.Log)
		})
	}
}

func TestAppAnteHandler_ConvertEthAccount(t *testing.T) {
	chainID := "kavatest_1-1"
	encodingConfig := app.MakeEncodingConfig()
	testPrivKeys, testAddresses := app.GeneratePrivKeyAddressPairs(10)

	tApp := app.NewTestApp()

	tApp = tApp.InitializeFromGenesisStatesWithTimeAndChainIDAndHeight(
		time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
		chainID,
		app.FixDefaultAccountUpgradeHeight-2,
		app.NewAuthBankGenesisBuilder().
			WithAccounts(&ethermint.EthAccount{
				BaseAccount: authtypes.NewBaseAccount(
					testAddresses[0],
					nil, // no pubkey set
					0,
					0,
				),
				CodeHash: common.BytesToHash(evmtypes.EmptyCodeHash).String(), // ethermint stores the codehash with a 0x prefix
			}).
			WithBalances(banktypes.Balance{
				Address: testAddresses[0].String(),
				Coins:   sdk.NewCoins(sdk.NewInt64Coin("ukava", 1e9)),
			}).
			BuildMarshalled(encodingConfig.Marshaler),
	)

	stdTx, err := helpers.GenTx(
		encodingConfig.TxConfig,
		[]sdk.Msg{banktypes.NewMsgSend(testAddresses[0], testAddresses[1], sdk.NewCoins(sdk.NewInt64Coin("ukava", 1)))},
		sdk.NewCoins(), // no fee
		helpers.DefaultGenTxGas,
		chainID,
		[]uint64{0},
		[]uint64{0},
		testPrivKeys[0],
	)
	require.NoError(t, err)
	txBytes, err := encodingConfig.TxConfig.TxEncoder()(stdTx)
	require.NoError(t, err)

	// Check accounts not converted before upgrade height
	checkAccountIsConverted(t, testAddresses[0], tApp, txBytes, (*ethermint.EthAccount)(nil))

	// Advance to upgrade height
	tApp.EndBlock(abci.RequestEndBlock{Height: app.FixDefaultAccountUpgradeHeight - 1})
	tApp.Commit()
	tApp.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.FixDefaultAccountUpgradeHeight, ChainID: chainID}})

	stdTx, err = helpers.GenTx(
		encodingConfig.TxConfig,
		[]sdk.Msg{banktypes.NewMsgSend(testAddresses[0], testAddresses[1], sdk.NewCoins(sdk.NewInt64Coin("ukava", 1)))},
		sdk.NewCoins(), // no fee
		helpers.DefaultGenTxGas,
		chainID,
		[]uint64{0},
		[]uint64{1},
		testPrivKeys[0],
	)
	require.NoError(t, err)
	txBytes, err = encodingConfig.TxConfig.TxEncoder()(stdTx)
	require.NoError(t, err)

	// Ensure CheckTx passes
	// Check account converted at upgrade height
	checkAccountIsConverted(t, testAddresses[0], tApp, txBytes, (*authtypes.BaseAccount)(nil))
}

func checkAccountIsConverted(t *testing.T, fromAddress sdk.AccAddress, tApp app.TestApp, txBytes []byte, expectedAccountType interface{}) {
	resDeliverTx := tApp.DeliverTx(
		abci.RequestDeliverTx{
			Tx: txBytes,
		},
	)
	require.Equal(t, uint32(sdkerrors.SuccessABCICode), resDeliverTx.Code, resDeliverTx.Log)

	ctx := tApp.NewContext(false, tmproto.Header{Height: tApp.LastBlockHeight() + 1})
	acc := tApp.GetAccountKeeper().GetAccount(ctx, fromAddress)
	require.IsType(t, expectedAccountType, acc)
}
