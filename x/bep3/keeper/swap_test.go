package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3/keeper"
	"github.com/kava-labs/kava/x/bep3/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

type AtomicSwapTestSuite struct {
	suite.Suite

	keeper   keeper.Keeper
	app      app.TestApp
	ctx      sdk.Context
	accounts []exported.Account
}

func (suite *AtomicSwapTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	// Set up test app with context, genesis state, and keeper
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	tApp.InitializeFromGenesisStates(
		NewBep3GenStateMulti(),
	)
	keeper := tApp.GetBep3Keeper()

	// Account set up
	_, addrs := app.GeneratePrivKeyAddressPairs(2)
	ak := tApp.GetAccountKeeper()
	acc1 := ak.NewAccountWithAddress(ctx, addrs[0])
	acc2 := ak.NewAccountWithAddress(ctx, addrs[1])

	// Fund an account with tokens and set accounts
	acc1.SetCoins(sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(1000000000))))
	ak.SetAccount(ctx, acc1)
	ak.SetAccount(ctx, acc2)

	suite.app = tApp
	suite.ctx = ctx
	suite.keeper = keeper
	suite.accounts = []exported.Account{acc1, acc2}
	return
}

func (suite *AtomicSwapTestSuite) TestCreateAtomicSwap() {
	someBlockTime := time.Now()

	timestamp := time.Now().Unix()
	randomNumber, _ := types.GenerateSecureRandomNumber()
	randomNumberHash := types.CalculateRandomHash(randomNumber.Bytes(), timestamp)

	type args struct {
		randomNumberHash    []byte
		timestamp           int64
		heightSpan          int64
		sender              sdk.AccAddress
		recipient           sdk.AccAddress
		senderOtherChain    string
		recipientOtherChain string
		coins               sdk.Coins
		expectedIncome      string
	}
	testCases := []struct {
		name       string
		blockTime  time.Time
		args       args
		expectPass bool
	}{
		{
			"normal",
			someBlockTime,
			args{
				randomNumberHash:    randomNumberHash,
				timestamp:           timestamp,
				heightSpan:          int64(360),
				sender:              suite.accounts[0].GetAddress(),
				recipient:           suite.accounts[1].GetAddress(),
				senderOtherChain:    binanceAddrs[0].String(),
				recipientOtherChain: binanceAddrs[1].String(),
				coins:               coinsSingle,
				expectedIncome:      coinsSingle.String(),
			},
			true,
		},
		{
			"unsupported asset",
			someBlockTime,
			args{
				randomNumberHash:    randomNumberHash,
				timestamp:           timestamp,
				heightSpan:          int64(360),
				sender:              suite.accounts[0].GetAddress(),
				recipient:           suite.accounts[1].GetAddress(),
				senderOtherChain:    binanceAddrs[2].String(),
				recipientOtherChain: binanceAddrs[1].String(),
				coins:               sdk.NewCoins(sdk.NewInt64Coin("xyz", int64(50000))),
				expectedIncome:      "50000xyz",
			},
			false,
		},
	}

	for _, tc := range testCases {
		// Create atomic swap
		err := suite.keeper.CreateAtomicSwap(suite.ctx, tc.args.randomNumberHash, tc.args.timestamp,
			tc.args.heightSpan, tc.args.sender, tc.args.recipient, tc.args.senderOtherChain,
			tc.args.recipientOtherChain, tc.args.coins, tc.args.expectedIncome)

		expectedSwapID := types.CalculateSwapID(tc.args.randomNumberHash, tc.args.sender, tc.args.senderOtherChain)

		if tc.expectPass {
			suite.NoError(err)

			// TODO: check coins moved
			// require.Equal(t, initialLiquidatorCoins.Sub(cs(tc.args.lot)), liquidatorCoins)

			// Check swap in store
			actualSwap, found := suite.keeper.GetAtomicSwap(suite.ctx, expectedSwapID)
			suite.True(found)
			suite.NotNil(actualSwap)

			// Confirm swap contents
			expectedSwap := types.Swap(
				types.AtomicSwap{
					Amount:              tc.args.coins,
					RandomNumberHash:    tc.args.randomNumberHash,
					ExpireHeight:        suite.ctx.BlockHeight() + tc.args.heightSpan,
					Timestamp:           tc.args.timestamp,
					Sender:              tc.args.sender,
					Recipient:           tc.args.recipient,
					SenderOtherChain:    tc.args.senderOtherChain,
					RecipientOtherChain: tc.args.recipientOtherChain,
					ClosedBlock:         0,
					Status:              types.Open,
				})
			suite.Equal(expectedSwap, actualSwap)
		} else {
			suite.Error(err)

			// TODO: check coins not moved
			// require.Equal(t, initialLiquidatorCoins, liquidatorCoins)

			// Check swap not in store
			_, found := suite.keeper.GetAtomicSwap(suite.ctx, expectedSwapID)
			suite.False(found)
		}
	}
}

// func (suite *AtomicSwapTestSuite) TestClaimAtomicSwap() {}

// func (suite *AtomicSwapTestSuite) TestRefundAtomicSwap() {}

func TestAtomicSwapTestSuite(t *testing.T) {
	suite.Run(t, new(AtomicSwapTestSuite))
}
