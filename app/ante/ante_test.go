package ante_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmdb "github.com/tendermint/tm-db"

	"github.com/kava-labs/kava/app"
)

func TestAppAnteHandler(t *testing.T) {
	testPrivKeys, testAddresses := app.GeneratePrivKeyAddressPairs(10)
	unauthed := testAddresses[0:2]
	unauthedKeys := testPrivKeys[0:2]
	// deputy := testAddresses[2]
	// deputyKey := testPrivKeys[2]
	oracles := testAddresses[3:6]
	oraclesKeys := testPrivKeys[3:6]
	manual := testAddresses[6:]
	manualKeys := testPrivKeys[6:]

	encodingConfig := app.MakeEncodingConfig()
	tApp := app.TestApp{
		App: *app.NewApp(
			log.NewNopLogger(),
			tmdb.NewMemDB(),
			nil,
			encodingConfig,
			app.Options{
				MempoolEnableAuth:    true,
				MempoolAuthAddresses: manual,
			},
		),
	}

	chainID := "internal-test-chain"
	tApp = tApp.InitializeFromGenesisStatesWithTimeAndChainID(
		time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
		chainID,
		NewFundedGenStateWithSameCoins(
			tApp.AppCodec(),
			sdk.NewCoins(sdk.NewInt64Coin("ukava", 1e9)),
			testAddresses,
		),
		// TODO see below
		// newBep3GenStateMulti(tApp.AppCodec(), deputy),
		// newPricefeedGenStateMulti(tApp.AppCodec(), oracles),
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
		// TODO add back when the bep3 module is reinstantiated
		// {
		// 	name:       "deputy",
		// 	address:    deputy,
		// 	privKey:    deputyKey,
		// 	expectPass: true,
		// },
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

// NewFundedGenStateWithSameCoins creates a (auth and bank) genesis state populated with accounts from the given addresses and balance.
func NewFundedGenStateWithSameCoins(cdc codec.JSONCodec, balance sdk.Coins, addresses []sdk.AccAddress) app.GenesisState {
	balances := make([]banktypes.Balance, len(addresses))
	for i, addr := range addresses {
		balances[i] = banktypes.Balance{
			Address: addr.String(),
			Coins:   balance,
		}
	}

	bankGenesis := banktypes.NewGenesisState(
		banktypes.DefaultParams(),
		balances,
		nil,
		[]banktypes.Metadata{}, // Metadata is not used in the antehandler to it is left out here
	)

	accounts := make(authtypes.GenesisAccounts, len(addresses))
	for i := range addresses {
		accounts[i] = authtypes.NewBaseAccount(addresses[i], nil, 0, 0)
	}

	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), accounts)

	return app.GenesisState{
		authtypes.ModuleName: cdc.MustMarshalJSON(authGenesis),
		banktypes.ModuleName: cdc.MustMarshalJSON(bankGenesis),
	}
}

// TODO Test pricefeed oracles and bep3 deputy txs can always get into the mempool.

// func newPricefeedGenStateMulti(cdc codec.JSONCodec, oracles []sdk.AccAddress) app.GenesisState {
// 	pfGenesis := pricefeed.GenesisState{
// 		Params: pricefeed.Params{
// 			Markets: []pricefeed.Market{
// 				{MarketID: "btc:usd", BaseAsset: "btc", QuoteAsset: "usd", Oracles: oracles, Active: true},
// 			},
// 		},
// 	}
// 	return app.GenesisState{pricefeed.ModuleName: cdc.MustMarshalJSON(pfGenesis)}
// }

// func newBep3GenStateMulti(cdc codec.JSONCodec, deputyAddress sdk.AccAddress) app.GenesisState {
// 	bep3Genesis := bep3.GenesisState{
// 		Params: bep3.Params{
// 			AssetParams: bep3.AssetParams{
// 				bep3.AssetParam{
// 					Denom:  "bnb",
// 					CoinID: 714,
// 					SupplyLimit: bep3.SupplyLimit{
// 						Limit:          sdk.NewInt(350000000000000),
// 						TimeLimited:    false,
// 						TimeBasedLimit: sdk.ZeroInt(),
// 						TimePeriod:     time.Hour,
// 					},
// 					Active:        true,
// 					DeputyAddress: deputyAddress,
// 					FixedFee:      sdk.NewInt(1000),
// 					MinSwapAmount: sdk.OneInt(),
// 					MaxSwapAmount: sdk.NewInt(1000000000000),
// 					MinBlockLock:  bep3.DefaultMinBlockLock,
// 					MaxBlockLock:  bep3.DefaultMaxBlockLock,
// 				},
// 			},
// 		},
// 		Supplies: bep3.AssetSupplies{
// 			bep3.NewAssetSupply(
// 				sdk.NewCoin("bnb", sdk.ZeroInt()),
// 				sdk.NewCoin("bnb", sdk.ZeroInt()),
// 				sdk.NewCoin("bnb", sdk.ZeroInt()),
// 				sdk.NewCoin("bnb", sdk.ZeroInt()),
// 				time.Duration(0),
// 			),
// 		},
// 		PreviousBlockTime: bep3.DefaultPreviousBlockTime,
// 	}
// 	return app.GenesisState{bep3.ModuleName: cdc.MustMarshalJSON(bep3Genesis)}
// }
