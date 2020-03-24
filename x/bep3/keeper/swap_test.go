package keeper_test

import (
	"fmt"
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
	BNB_DENOM           = "bnb"
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
		coins = append(coins, cs(c(BNB_DENOM, STARING_BNB_BALANCE)))
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
	_, _ = suite.keeper.LoadAssetSupply(suite.ctx, BNB_DENOM)
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
		crossChain          bool
	}
	testCases := []struct {
		name          string
		blockTime     time.Time
		args          args
		expectPass    bool
		shouldBeFound bool
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
				coins:               cs(c(BNB_DENOM, 50000)),
				expectedIncome:      fmt.Sprintf("50000%s", BNB_DENOM),
				crossChain:          true,
			},
			true,
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
				crossChain:          true,
			},
			false,
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
				coins:               cs(c(BNB_DENOM, 50000)),
				expectedIncome:      fmt.Sprintf("50000%s", BNB_DENOM),
				crossChain:          true,
			},
			false,
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
				coins:               cs(c(BNB_DENOM, 50000)),
				expectedIncome:      fmt.Sprintf("50000%s", BNB_DENOM),
				crossChain:          true,
			},
			false,
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
				coins:               cs(c(BNB_DENOM, 50000)),
				expectedIncome:      fmt.Sprintf("50000%s", BNB_DENOM),
				crossChain:          true,
			},
			false,
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
				coins:               cs(c(BNB_DENOM, 50000)),
				expectedIncome:      fmt.Sprintf("50000%s", BNB_DENOM),
				crossChain:          true,
			},
			false,
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
				coins:               cs(c(BNB_DENOM, 0)),
				expectedIncome:      fmt.Sprintf("0%s", BNB_DENOM),
				crossChain:          true,
			},
			false,
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
				coins:               cs(c(BNB_DENOM, 50000)),
				expectedIncome:      "50000bnb",
				crossChain:          true,
			},
			false,
			true,
		},
	}

	for _, tc := range testCases {
		suite.SetupTest()
		suite.Run(tc.name, func() {
			if tc.name == "duplicate swap" {
				err := suite.keeper.CreateAtomicSwap(suite.ctx, tc.args.randomNumberHash, tc.args.timestamp,
					tc.args.heightSpan, tc.args.sender, tc.args.recipient, tc.args.senderOtherChain,
					tc.args.recipientOtherChain, tc.args.coins, tc.args.expectedIncome, tc.args.crossChain)
				suite.Nil(err)
			}

			// Load asset denom (required for zero coins test case)
			var swapAssetDenom string
			if len(tc.args.coins) == 1 {
				swapAssetDenom = tc.args.coins[0].Denom
			} else {
				swapAssetDenom = BNB_DENOM
			}

			// Load sender's account prior to swap creation
			ak := suite.app.GetAccountKeeper()
			senderAccPre := ak.GetAccount(suite.ctx, tc.args.sender)
			senderBalancePre := senderAccPre.GetCoins().AmountOf(swapAssetDenom)
			inSwapSupplyPre, _ := suite.keeper.LoadAssetSupply(suite.ctx, swapAssetDenom)

			// Create atomic swap
			err := suite.keeper.CreateAtomicSwap(suite.ctx, tc.args.randomNumberHash, tc.args.timestamp,
				tc.args.heightSpan, tc.args.sender, tc.args.recipient, tc.args.senderOtherChain,
				tc.args.recipientOtherChain, tc.args.coins, tc.args.expectedIncome, tc.args.crossChain)

			// Load sender's account after swap creation
			senderAccPost := ak.GetAccount(suite.ctx, tc.args.sender)
			senderBalancePost := senderAccPost.GetCoins().AmountOf(swapAssetDenom)
			inSwapSupplyPost, _ := suite.keeper.LoadAssetSupply(suite.ctx, swapAssetDenom)

			// Load expected swap ID
			expectedSwapID := types.CalculateSwapID(tc.args.randomNumberHash, tc.args.sender, tc.args.senderOtherChain)

			if tc.expectPass {
				suite.NoError(err)
				// Check coins moved
				suite.Equal(senderBalancePre.Sub(tc.args.coins[0].Amount), senderBalancePost)
				// Check in swap supply increased
				suite.Equal(inSwapSupplyPre.Add(tc.args.coins[0]), inSwapSupplyPost)

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
						CrossChain:          tc.args.crossChain,
					})
				suite.Equal(expectedSwap, actualSwap)
			} else {
				suite.Error(err)
				// Check coins not moved
				suite.Equal(senderBalancePre, senderBalancePost)
				// Check in swap supply not increased
				suite.Equal(inSwapSupplyPre, inSwapSupplyPost)

				// Check if swap found in store
				_, found := suite.keeper.GetAtomicSwap(suite.ctx, expectedSwapID)
				if !tc.shouldBeFound {
					suite.False(found)
				} else {
					suite.True(found)
				}
			}
		})
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
		suite.Run(tc.name, func() {
			// Create atomic swap
			expectedRecipient := suite.addrs[5]
			expectedClaimAmount := cs(c(BNB_DENOM, 50000))

			err := suite.keeper.CreateAtomicSwap(suite.ctx, suite.randomNumberHashes[i], suite.timestamps[i],
				int64(360), suite.addrs[i], expectedRecipient, binanceAddrs[0].String(), binanceAddrs[1].String(),
				expectedClaimAmount, expectedClaimAmount.String(), true)
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

			// Load expected recipient's account prior to claim attempt
			ak := suite.app.GetAccountKeeper()
			expectedRecipientAccPre := ak.GetAccount(tc.claimCtx, expectedRecipient)
			expectedRecipientBalancePre := expectedRecipientAccPre.GetCoins().AmountOf(expectedClaimAmount[0].Denom)
			// Load asset supplies prior to claim attempt
			inSwapSupplyPre, assetSupplyPre := suite.keeper.LoadAssetSupply(tc.claimCtx, expectedClaimAmount[0].Denom)

			// Attempt to claim atomic swap
			err = suite.keeper.ClaimAtomicSwap(tc.claimCtx, expectedRecipient, claimSwapID, claimRandomNumber)

			// Load expected recipient's account after the claim attempt
			expectedRecipientAccPost := ak.GetAccount(tc.claimCtx, expectedRecipient)
			expectedRecipientBalancePost := expectedRecipientAccPost.GetCoins().AmountOf(expectedClaimAmount[0].Denom)
			// Load asset supplies after the claim attempt
			inSwapSupplyPost, assetSupplyPost := suite.keeper.LoadAssetSupply(tc.claimCtx, expectedClaimAmount[0].Denom)

			if tc.expectPass {
				suite.NoError(err)
				// Check coins moved
				suite.Equal(expectedRecipientBalancePre.Add(expectedClaimAmount[0].Amount), expectedRecipientBalancePost)
				// Check in swap supply decreased
				suite.True(inSwapSupplyPre.Sub(expectedClaimAmount[0]).IsEqual(inSwapSupplyPost))
				// Check asset supply increased
				suite.True(assetSupplyPre.Add(expectedClaimAmount[0]).IsEqual(assetSupplyPost))
			} else {
				suite.Error(err)
				// Check coins not moved
				suite.Equal(expectedRecipientBalancePre, expectedRecipientBalancePost)
				// Check in swap supply not decreased
				suite.Equal(inSwapSupplyPre, inSwapSupplyPost)
				// Check asset supply not increased
				suite.Equal(assetSupplyPre, assetSupplyPost)
			}
		})
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
		suite.Run(tc.name, func() {
			// Create atomic swap
			originalSender := suite.addrs[i]
			expectedRefundAmount := cs(c(BNB_DENOM, 50000))

			err := suite.keeper.CreateAtomicSwap(suite.ctx, suite.randomNumberHashes[i], suite.timestamps[i],
				int64(360), originalSender, suite.addrs[8], binanceAddrs[0].String(), binanceAddrs[1].String(),
				expectedRefundAmount, expectedRefundAmount.String(), true)
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

			// Load sender's account prior to swap refund
			ak := suite.app.GetAccountKeeper()
			originalSenderAccPre := ak.GetAccount(tc.refundCtx, originalSender)
			originalSenderBalancePre := originalSenderAccPre.GetCoins().AmountOf(expectedRefundAmount[0].Denom)
			// Load asset supplies prior to swap refund
			inSwapSupplyPre, assetSupplyPre := suite.keeper.LoadAssetSupply(tc.refundCtx, expectedRefundAmount[0].Denom)

			// Attempt to refund atomic swap
			err = suite.keeper.RefundAtomicSwap(tc.refundCtx, originalSender, refundSwapID)

			// Load sender's account after refund
			originalSenderAccPost := ak.GetAccount(tc.refundCtx, originalSender)
			originalSenderBalancePost := originalSenderAccPost.GetCoins().AmountOf(expectedRefundAmount[0].Denom)
			// Load asset supplies after to swap refund
			inSwapSupplyPost, assetSupplyPost := suite.keeper.LoadAssetSupply(tc.refundCtx, expectedRefundAmount[0].Denom)

			if tc.expectPass {
				suite.NoError(err)
				// Check coins moved
				suite.Equal(originalSenderBalancePre.Add(expectedRefundAmount[0].Amount), originalSenderBalancePost)
				// Check in swap supply decreased
				suite.True(inSwapSupplyPre.Sub(expectedRefundAmount[0]).IsEqual(inSwapSupplyPost))
				// Check asset supply not changed
				suite.Equal(assetSupplyPre, assetSupplyPost)
			} else {
				suite.Error(err)
				// Check coins not moved
				suite.Equal(originalSenderBalancePre, originalSenderBalancePost)
				// Check in swap supply not decreased
				suite.Equal(inSwapSupplyPre, inSwapSupplyPost)
			}
		})
	}
}

func TestAtomicSwapTestSuite(t *testing.T) {
	suite.Run(t, new(AtomicSwapTestSuite))
}
