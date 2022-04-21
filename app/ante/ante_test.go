package ante_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmdb "github.com/tendermint/tm-db"

	"github.com/kava-labs/kava/app"
	bep3types "github.com/kava-labs/kava/x/bep3/types"
	pricefeedtypes "github.com/kava-labs/kava/x/pricefeed/types"
)

func TestAppAnteHandler(t *testing.T) {
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
