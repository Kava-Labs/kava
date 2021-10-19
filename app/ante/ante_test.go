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
)

func TestAppAnteHandler(t *testing.T) {
	testPrivKeys, testAddresses := app.GeneratePrivKeyAddressPairs(10)
	unauthed := testAddresses[0:2]
	unathedKeys := testPrivKeys[0:2]

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
		// TODO add in tests for deputy and pricefeed addresses when those modules are added back
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
