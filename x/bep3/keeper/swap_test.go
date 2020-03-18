package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3"
	"github.com/kava-labs/kava/x/bep3/keeper"
	"github.com/kava-labs/kava/x/bep3/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	tmtime "github.com/tendermint/tendermint/types/time"
)

type AtomicSwapTestSuite struct {
	suite.Suite

	keeper             keeper.Keeper
	app                app.TestApp
	ctx                sdk.Context
	addrs              []sdk.AccAddress
	timestamps         []int64
	randomNumbers      [][]byte
	randomNumberHashes []cmn.HexBytes
}

const (
	STARING_BNB_BALANCE = int64(1000000000)
)

func (suite *AtomicSwapTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	// Initialize test app and set context
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})

	// Generate 10 timestamps and random number hashes
	var timestamps []int64
	var randomNumbers [][]byte
	var randomNumberHashes []cmn.HexBytes
	for i := 0; i < 10; i++ {
		timestamp := ts(i)
		randomNumber, _ := types.GenerateSecureRandomNumber()
		randomNumberHash := types.CalculateRandomHash(randomNumber.Bytes(), timestamp)
		timestamps = append(timestamps, timestamp)
		randomNumbers = append(randomNumbers, randomNumber.Bytes())
		randomNumberHashes = append(randomNumberHashes, randomNumberHash)
	}

	// Create and load 20 accounts with bnb tokens
	coins := []sdk.Coins{}
	for i := 0; i < 20; i++ {
		coins = append(coins, cs(c("bnb", STARING_BNB_BALANCE)))
	}
	_, addrs := app.GeneratePrivKeyAddressPairs(20)
	authGS := app.NewAuthGenState(addrs, coins)

	// Initialize genesis state
	tApp.InitializeFromGenesisStates(authGS, NewBep3GenStateMulti())

	suite.app = tApp
	suite.ctx = ctx
	suite.keeper = tApp.GetBep3Keeper()
	suite.addrs = addrs
	suite.timestamps = timestamps
	suite.randomNumbers = randomNumbers
	suite.randomNumberHashes = randomNumberHashes
	return
}

func (suite *AtomicSwapTestSuite) TestCreateAtomicSwap() {
	suite.SetupTest()
	currentTmTime := tmtime.Now()
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
			currentTmTime,
			args{
				randomNumberHash:    suite.randomNumberHashes[0],
				timestamp:           suite.timestamps[0],
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
			currentTmTime,
			args{
				randomNumberHash:    suite.randomNumberHashes[1],
				timestamp:           suite.timestamps[1],
				heightSpan:          int64(360),
				sender:              suite.addrs[1],
				recipient:           suite.addrs[2],
				senderOtherChain:    binanceAddrs[0].String(),
				recipientOtherChain: binanceAddrs[1].String(),
				coins:               cs(c("xyz", 50000)),
				expectedIncome:      "50000xyz",
			},
			false,
		},
		{
			"past timestamp",
			currentTmTime,
			args{
				randomNumberHash:    suite.randomNumberHashes[2],
				timestamp:           suite.timestamps[2] - 2000,
				heightSpan:          int64(360),
				sender:              suite.addrs[2],
				recipient:           suite.addrs[3],
				senderOtherChain:    binanceAddrs[0].String(),
				recipientOtherChain: binanceAddrs[1].String(),
				coins:               cs(c("bnb", 50000)),
				expectedIncome:      "50000bnb",
			},
			false,
		},
		{
			"future timestamp",
			currentTmTime,
			args{
				randomNumberHash:    suite.randomNumberHashes[3],
				timestamp:           suite.timestamps[3] + 5000,
				heightSpan:          int64(360),
				sender:              suite.addrs[3],
				recipient:           suite.addrs[4],
				senderOtherChain:    binanceAddrs[0].String(),
				recipientOtherChain: binanceAddrs[1].String(),
				coins:               cs(c("bnb", 50000)),
				expectedIncome:      "50000bnb",
			},
			false,
		},
		{
			"small height span",
			currentTmTime,
			args{
				randomNumberHash:    suite.randomNumberHashes[4],
				timestamp:           suite.timestamps[4],
				heightSpan:          int64(5),
				sender:              suite.addrs[4],
				recipient:           suite.addrs[5],
				senderOtherChain:    binanceAddrs[0].String(),
				recipientOtherChain: binanceAddrs[1].String(),
				coins:               cs(c("bnb", 50000)),
				expectedIncome:      "50000bnb",
			},
			false,
		},
		{
			"big height span",
			currentTmTime,
			args{
				randomNumberHash:    suite.randomNumberHashes[5],
				timestamp:           suite.timestamps[5],
				heightSpan:          int64(1000000),
				sender:              suite.addrs[5],
				recipient:           suite.addrs[6],
				senderOtherChain:    binanceAddrs[0].String(),
				recipientOtherChain: binanceAddrs[1].String(),
				coins:               cs(c("bnb", 50000)),
				expectedIncome:      "50000bnb",
			},
			false,
		},
		{
			"zero amount",
			currentTmTime,
			args{
				randomNumberHash:    suite.randomNumberHashes[6],
				timestamp:           suite.timestamps[6],
				heightSpan:          int64(360),
				sender:              suite.addrs[6],
				recipient:           suite.addrs[7],
				senderOtherChain:    binanceAddrs[0].String(),
				recipientOtherChain: binanceAddrs[1].String(),
				coins:               cs(c("bnb", 0)),
				expectedIncome:      "0bnb",
			},
			false,
		},
		{
			"duplicate swap",
			currentTmTime,
			args{
				randomNumberHash:    suite.randomNumberHashes[0],
				timestamp:           suite.timestamps[0],
				heightSpan:          int64(360),
				sender:              suite.addrs[0],
				recipient:           suite.addrs[1],
				senderOtherChain:    binanceAddrs[0].String(),
				recipientOtherChain: binanceAddrs[1].String(),
				coins:               cs(c("bnb", 50000)),
				expectedIncome:      "50000bnb",
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
	suite.SetupTest()
	invalidRandomNumber, _ := types.GenerateSecureRandomNumber()
	type args struct {
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
				swapID:       []byte{},
				randomNumber: []byte{},
			},
			true,
		},
		{
			"invalid random number",
			suite.ctx,
			args{
				swapID:       []byte{},
				randomNumber: invalidRandomNumber.Bytes(),
			},
			false,
		},
		{
			"wrong swap ID",
			suite.ctx,
			args{
				swapID:       types.CalculateSwapID(suite.randomNumberHashes[3], suite.addrs[7], binanceAddrs[2].String()),
				randomNumber: []byte{},
			},
			false,
		},
		{
			"past expiration",
			suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 2000),
			args{
				swapID:       []byte{},
				randomNumber: []byte{},
			},
			false,
		},
	}

	for i, tc := range testCases {
		// Create atomic swap
		expectedRecipient := suite.addrs[5]
		expectedClaimAmount := cs(c("bnb", 50000))

		err := suite.keeper.CreateAtomicSwap(suite.ctx, suite.randomNumberHashes[i], suite.timestamps[i],
			int64(360), suite.addrs[i], expectedRecipient, binanceAddrs[0].String(), binanceAddrs[1].String(),
			expectedClaimAmount, expectedClaimAmount.String())
		suite.NoError(err)

		realSwapID := types.CalculateSwapID(suite.randomNumberHashes[i], suite.addrs[i], binanceAddrs[0].String())

		// If args contains an invalid swap ID claim attempt will use it instead of the real swap ID
		var claimSwapID []byte
		if len(tc.args.swapID) == 0 {
			claimSwapID = realSwapID
		} else {
			claimSwapID = tc.args.swapID
		}

		// If args contains an invalid random number claim attempt will use it instead of the real random number
		var claimRandomNumber []byte
		if len(tc.args.randomNumber) == 0 {
			claimRandomNumber = suite.randomNumbers[i]
		} else {
			claimRandomNumber = tc.args.randomNumber
		}

		// Run the beginblocker before attempting claim
		bep3.BeginBlocker(tc.claimCtx, suite.keeper)

		// Attempt to claim atomic swap
		err = suite.keeper.ClaimAtomicSwap(tc.claimCtx, expectedRecipient, claimSwapID, claimRandomNumber)

		// Load sender's account at time of claim
		ak := suite.app.GetAccountKeeper()
		recipientAcc := ak.GetAccount(tc.claimCtx, expectedRecipient)

		if tc.expectPass {
			suite.NoError(err)

			// Check coins moved
			suite.Equal(
				sdk.NewInt(STARING_BNB_BALANCE).Add(expectedClaimAmount[0].Amount),
				recipientAcc.GetCoins().AmountOf("bnb"),
			)

			// Check swap not in store
			_, found := suite.keeper.GetAtomicSwap(suite.ctx, realSwapID)
			suite.False(found)
		} else {
			suite.Error(err)

			// Check coins not moved
			existingAdditionalBalance := expectedClaimAmount[0].Amount
			suite.Equal(
				sdk.NewInt(STARING_BNB_BALANCE).Add(existingAdditionalBalance),
				recipientAcc.GetCoins().AmountOf("bnb"),
			)

			// Check swap still in store
			swap, found := suite.keeper.GetAtomicSwap(suite.ctx, realSwapID)
			suite.True(found)
			suite.NotNil(swap)
		}
	}
}

func (suite *AtomicSwapTestSuite) TestRefundAtomicSwap() {
	suite.SetupTest()

	type args struct {
		swapID []byte
	}
	testCases := []struct {
		name       string
		refundCtx  sdk.Context
		args       args
		expectPass bool
	}{
		{
			"normal",
			suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 400),
			args{
				swapID: []byte{},
			},
			true,
		},
		{
			"before expiration",
			suite.ctx,
			args{
				swapID: []byte{},
			},
			false,
		},
		{
			"wrong swapID",
			suite.ctx,
			args{
				swapID: types.CalculateSwapID(suite.randomNumberHashes[6], suite.addrs[1], binanceAddrs[1].String()),
			},
			false,
		},
	}

	for i, tc := range testCases {
		// Create atomic swap
		originalSender := suite.addrs[i]
		expectedRefundAmount := cs(c("bnb", 50000))

		err := suite.keeper.CreateAtomicSwap(suite.ctx, suite.randomNumberHashes[i], suite.timestamps[i],
			int64(360), originalSender, suite.addrs[8], binanceAddrs[0].String(), binanceAddrs[1].String(),
			expectedRefundAmount, expectedRefundAmount.String())
		suite.NoError(err)

		realSwapID := types.CalculateSwapID(suite.randomNumberHashes[i], originalSender, binanceAddrs[0].String())

		// If args contains an invalid swap ID refund attempt will use it instead of the real swap ID
		var refundSwapID []byte
		if len(tc.args.swapID) == 0 {
			refundSwapID = realSwapID
		} else {
			refundSwapID = tc.args.swapID
		}

		// Run the beginblocker before attempting refund
		bep3.BeginBlocker(tc.refundCtx, suite.keeper)

		// Attempt to refund atomic swap
		err = suite.keeper.RefundAtomicSwap(tc.refundCtx, originalSender, refundSwapID)

		// Load sender's account at time of refund
		ak := suite.app.GetAccountKeeper()
		senderAcc := ak.GetAccount(tc.refundCtx, originalSender)

		if tc.expectPass {
			suite.NoError(err)

			// Check coins moved
			suite.Equal(
				sdk.NewInt(STARING_BNB_BALANCE),
				senderAcc.GetCoins().AmountOf("bnb"),
			)

			// Check swap not in store
			_, found := suite.keeper.GetAtomicSwap(suite.ctx, realSwapID)
			suite.False(found)
		} else {
			suite.Error(err)

			// Check coins not moved
			existingAdditionalBalance := expectedRefundAmount[0].Amount
			suite.Equal(
				sdk.NewInt(STARING_BNB_BALANCE).Sub(existingAdditionalBalance),
				senderAcc.GetCoins().AmountOf("bnb"),
			)

			// Check swap still in store
			swap, found := suite.keeper.GetAtomicSwap(suite.ctx, realSwapID)
			suite.True(found)
			suite.NotNil(swap)
		}
	}
}

func TestAtomicSwapTestSuite(t *testing.T) {
	suite.Run(t, new(AtomicSwapTestSuite))
}
