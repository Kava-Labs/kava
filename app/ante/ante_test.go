package ante_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/log"
	tmdb "github.com/tendermint/tm-db"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3"
	"github.com/kava-labs/kava/x/pricefeed"
)

func TestAppAnteHandler(t *testing.T) {
	testPrivKeys, testAddresses := app.GeneratePrivKeyAddressPairs(10)
	unauthed := testAddresses[0:2]
	unathedKeys := testPrivKeys[0:2]
	deputy := testAddresses[2]
	deputyKey := testPrivKeys[2]
	oracles := testAddresses[3:6]
	oraclesKeys := testPrivKeys[3:6]
	manual := testAddresses[6:]
	manualKeys := testPrivKeys[6:]

	db := tmdb.NewMemDB()
	tApp := app.TestApp{
		App: *app.NewApp(log.NewNopLogger(), db, nil,
			app.AppOptions{
				MempoolEnableAuth:    true,
				MempoolAuthAddresses: manual,
			},
		),
	}

	chainID := "internal-test-chain"
	tApp = tApp.InitializeFromGenesisStatesWithTimeAndChainID(
		time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
		chainID,
		NewAuthGenStateWithSameCoins(
			sdk.NewCoins(sdk.NewInt64Coin("ukava", 1_000_000_000)),
			testAddresses,
		),
		newBep3GenStateMulti(deputy),
		newPricefeedGenStateMulti(oracles),
	)

	testcases := []struct {
		name       string
		address    sdk.AccAddress
		privKey    crypto.PrivKey
		expectPass bool
	}{
		{
			name:       "unauthorized",
			address:    unauthed[1],
			privKey:    unathedKeys[1],
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
			stdTx := helpers.GenTx(
				[]sdk.Msg{
					bank.NewMsgSend(
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
			txBytes, err := auth.DefaultTxEncoder(tApp.Codec())(stdTx)
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

func NewAuthGenStateWithSameCoins(coins sdk.Coins, addresses []sdk.AccAddress) app.GenesisState {
	coinsList := make([]sdk.Coins, len(addresses))
	for i := range addresses {
		coinsList[i] = coins
	}
	return app.NewAuthGenState(addresses, coinsList)
}

func newPricefeedGenStateMulti(oracles []sdk.AccAddress) app.GenesisState {
	pfGenesis := pricefeed.GenesisState{
		Params: pricefeed.Params{
			Markets: []pricefeed.Market{
				{MarketID: "btc:usd", BaseAsset: "btc", QuoteAsset: "usd", Oracles: oracles, Active: true},
				{MarketID: "xrp:usd", BaseAsset: "xrp", QuoteAsset: "usd", Oracles: oracles, Active: true},
			},
		},
	}
	return app.GenesisState{pricefeed.ModuleName: pricefeed.ModuleCdc.MustMarshalJSON(pfGenesis)}
}

func newBep3GenStateMulti(deputyAddress sdk.AccAddress) app.GenesisState {
	bep3Genesis := bep3.GenesisState{
		Params: bep3.Params{
			AssetParams: bep3.AssetParams{
				bep3.AssetParam{
					Denom:  "bnb",
					CoinID: 714,
					SupplyLimit: bep3.SupplyLimit{
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
					MinBlockLock:  bep3.DefaultMinBlockLock,
					MaxBlockLock:  bep3.DefaultMaxBlockLock,
				},
				bep3.AssetParam{
					Denom:  "inc",
					CoinID: 9999,
					SupplyLimit: bep3.SupplyLimit{
						Limit:          sdk.NewInt(100000000000000),
						TimeLimited:    true,
						TimeBasedLimit: sdk.NewInt(50000000000),
						TimePeriod:     time.Hour,
					},
					Active:        false,
					DeputyAddress: deputyAddress,
					FixedFee:      sdk.NewInt(1000),
					MinSwapAmount: sdk.OneInt(),
					MaxSwapAmount: sdk.NewInt(100000000000),
					MinBlockLock:  bep3.DefaultMinBlockLock,
					MaxBlockLock:  bep3.DefaultMaxBlockLock,
				},
			},
		},
		Supplies: bep3.AssetSupplies{
			bep3.NewAssetSupply(
				sdk.NewCoin("bnb", sdk.ZeroInt()),
				sdk.NewCoin("bnb", sdk.ZeroInt()),
				sdk.NewCoin("bnb", sdk.ZeroInt()),
				sdk.NewCoin("bnb", sdk.ZeroInt()),
				time.Duration(0),
			),
			bep3.NewAssetSupply(
				sdk.NewCoin("inc", sdk.ZeroInt()),
				sdk.NewCoin("inc", sdk.ZeroInt()),
				sdk.NewCoin("inc", sdk.ZeroInt()),
				sdk.NewCoin("inc", sdk.ZeroInt()),
				time.Duration(0),
			),
		},
		PreviousBlockTime: bep3.DefaultPreviousBlockTime,
	}
	return app.GenesisState{bep3.ModuleName: bep3.ModuleCdc.MustMarshalJSON(bep3Genesis)}
}
