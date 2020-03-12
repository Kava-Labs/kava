package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3/keeper"
	"github.com/kava-labs/kava/x/bep3/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

type AtomicSwapTestSuite struct {
	suite.Suite

	keeper keeper.Keeper
	app    app.TestApp
	ctx    sdk.Context
	addrs  []sdk.AccAddress
}

const (
	STARING_BNB_BALANCE = 1000000000
)

func (suite *AtomicSwapTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	// Initialize test app and set context
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})

	// Set up auth genesis state
	coins := []sdk.Coins{}
	for j := 0; j < 30; j++ {
		coins = append(coins, cs(c("bnb", STARING_BNB_BALANCE)))
	}
	_, addrs := app.GeneratePrivKeyAddressPairs(30)
	authGS := app.NewAuthGenState(addrs, coins)

	// Initialize genesis state
	tApp.InitializeFromGenesisStates(
		authGS,
		NewBep3GenStateMulti(),
	)
	// Load keeper
	keeper := tApp.GetBep3Keeper()

	suite.app = tApp
	suite.ctx = ctx
	suite.keeper = keeper
	suite.addrs = addrs
	return
}

func (suite *AtomicSwapTestSuite) TestCreateAtomicSwap() {
	someBlockTime := tmtime.Now()

	timestamp := tmtime.Now().Unix()
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
				sender:              suite.addrs[0],
				recipient:           suite.addrs[1],
				senderOtherChain:    binanceAddrs[0].String(),
				recipientOtherChain: binanceAddrs[1].String(),
				coins:               cs(c("bnb", 50000)),
				expectedIncome:      "50000bnb",
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
				sender:              suite.addrs[2],
				recipient:           suite.addrs[3],
				senderOtherChain:    binanceAddrs[2].String(),
				recipientOtherChain: binanceAddrs[1].String(),
				coins:               cs(c("xyz", 50000)),
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

		// Load expected swap ID
		expectedSwapID := types.CalculateSwapID(tc.args.randomNumberHash, tc.args.sender, tc.args.senderOtherChain)
		// Load sender's account
		ak := suite.app.GetAccountKeeper()
		senderAcc := ak.GetAccount(suite.ctx, tc.args.sender)

		if tc.expectPass {
			suite.NoError(err)

			// Check coins moved
			suite.Equal(
				i(STARING_BNB_BALANCE).Sub(tc.args.coins[0].Amount),
				senderAcc.GetCoins().AmountOf("bnb"),
			)

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

			// Check coins not moved
			suite.Equal(i(STARING_BNB_BALANCE), senderAcc.GetCoins().AmountOf("bnb"))

			// Check swap not in store
			_, found := suite.keeper.GetAtomicSwap(suite.ctx, expectedSwapID)
			suite.False(found)
		}
	}
}

func (suite *AtomicSwapTestSuite) TestClaimAtomicSwap() {
	// Generate timestamp and random number, calculate expected swap ID
	timestamp := tmtime.Now().Unix()
	randomNumber, _ := types.GenerateSecureRandomNumber()
	randomNumberHash := types.CalculateRandomHash(randomNumber.Bytes(), timestamp)
	expectedSwapID := types.CalculateSwapID(randomNumberHash, suite.addrs[0], binanceAddrs[0].String())

	// Generate another random number
	invalidRandomNumber, _ := types.GenerateSecureRandomNumber()

	type args struct {
		from         sdk.AccAddress
		swapID       []byte
		randomNumber []byte
	}
	testCases := []struct {
		name       string
		claimCtx   sdk.Context
		args       args
		expectPass bool
	}{
		{
			"normal",
			suite.ctx,
			args{
				from:         suite.addrs[10],
				swapID:       expectedSwapID,
				randomNumber: randomNumber.Bytes(),
			},
			true,
		},
		{
			"invalid random number",
			suite.ctx,
			args{
				from:         suite.addrs[10],
				swapID:       expectedSwapID,
				randomNumber: invalidRandomNumber.Bytes(),
			},
			false,
		},
		// {
		// 	TODO: "past expiration",
		// 	suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Duration(1) * time.Hour)),
		// 	args{
		// 		from:         suite.addrs[10],
		// 		swapID:       expectedSwapID,
		// 		randomNumber: randomNumber.Bytes(),
		// 	},
		// 	false,
		// },
	}

	for _, tc := range testCases {
		// Set up expected post-claim values
		expectedRecipient := suite.addrs[5]
		expectedClaimAmount := cs(c("bnb", 50000))

		// Create atomic swap
		err := suite.keeper.CreateAtomicSwap(suite.ctx, randomNumberHash, timestamp, int64(360), suite.addrs[0],
			expectedRecipient, binanceAddrs[0].String(), binanceAddrs[1].String(), expectedClaimAmount, "50000bnb")
		suite.NoError(err)

		// Claim atomic swap
		err = suite.keeper.ClaimAtomicSwap(tc.claimCtx, tc.args.from, tc.args.swapID, tc.args.randomNumber)

		// Load sender's account
		ak := suite.app.GetAccountKeeper()
		recipientAcc := ak.GetAccount(suite.ctx, expectedRecipient)

		if tc.expectPass {
			suite.NoError(err)

			// Check coins moved
			suite.Equal(
				i(STARING_BNB_BALANCE).Add(expectedClaimAmount[0].Amount),
				recipientAcc.GetCoins().AmountOf("bnb"),
			)

			// Check swap not in store
			_, found := suite.keeper.GetAtomicSwap(suite.ctx, expectedSwapID)
			suite.False(found)
		} else {
			suite.Error(err)

			// Check coins not moved
			existingAdditionalBalance := expectedClaimAmount[0].Amount
			suite.Equal(i(STARING_BNB_BALANCE).Add(existingAdditionalBalance), recipientAcc.GetCoins().AmountOf("bnb"))

			// Check swap still in store
			swap, found := suite.keeper.GetAtomicSwap(suite.ctx, expectedSwapID)
			suite.True(found)
			suite.NotNil(swap)
		}
	}
}

func (suite *AtomicSwapTestSuite) TestRefundAtomicSwap() {
	// Generate timestamp and random number, calculate expected swap ID
	timestamp := tmtime.Now().Unix()
	randomNumber, _ := types.GenerateSecureRandomNumber()
	randomNumberHash := types.CalculateRandomHash(randomNumber.Bytes(), timestamp)
	expectedSwapID := types.CalculateSwapID(randomNumberHash, suite.addrs[7], binanceAddrs[0].String())

	type args struct {
		from   sdk.AccAddress
		swapID []byte
	}
	testCases := []struct {
		name       string
		refundCtx  sdk.Context
		args       args
		expectPass bool
	}{
		{
			"before expiration",
			suite.ctx,
			args{
				from:   suite.addrs[10],
				swapID: expectedSwapID,
			},
			false,
		},
		// {
		// 	TODO: "normal",
		// 	TODO: should this be BlockHeight instead of BlockTime?
		// 	suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Duration(1) * time.Hour)),
		// 	args{
		// 		from:         suite.addrs[10],
		// 		swapID:       expectedSwapID,
		// 	},
		// 	true,
		// },
	}

	for _, tc := range testCases {
		// Set up expected post-refund values
		originalSender := suite.addrs[7]
		expectedRecipient := suite.addrs[8]
		expectedRefundAmount := cs(c("bnb", 50000))

		// Create atomic swap
		err := suite.keeper.CreateAtomicSwap(suite.ctx, randomNumberHash, timestamp, int64(360), originalSender,
			expectedRecipient, binanceAddrs[0].String(), binanceAddrs[1].String(), expectedRefundAmount, "50000bnb")
		suite.NoError(err)

		// Refund atomic swap
		err = suite.keeper.RefundAtomicSwap(tc.refundCtx, tc.args.from, tc.args.swapID)

		// Load sender's account
		ak := suite.app.GetAccountKeeper()
		senderAcc := ak.GetAccount(suite.ctx, originalSender)

		if tc.expectPass {
			suite.NoError(err)

			// Check coins moved
			suite.Equal(
				i(STARING_BNB_BALANCE).Add(expectedRefundAmount[0].Amount),
				senderAcc.GetCoins().AmountOf("bnb"),
			)

			// Check swap not in store
			_, found := suite.keeper.GetAtomicSwap(suite.ctx, expectedSwapID)
			suite.False(found)
		} else {
			suite.Error(err)

			// Check coins not moved
			existingAdditionalBalance := expectedRefundAmount[0].Amount
			suite.Equal(i(STARING_BNB_BALANCE).Sub(existingAdditionalBalance), senderAcc.GetCoins().AmountOf("bnb"))

			// Check swap still in store
			swap, found := suite.keeper.GetAtomicSwap(suite.ctx, expectedSwapID)
			suite.True(found)
			suite.NotNil(swap)
		}
	}
}

func TestAtomicSwapTestSuite(t *testing.T) {
	suite.Run(t, new(AtomicSwapTestSuite))
}
