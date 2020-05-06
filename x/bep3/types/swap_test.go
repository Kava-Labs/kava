package types_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3/types"
)

type AtomicSwapTestSuite struct {
	suite.Suite
	addrs              []sdk.AccAddress
	timestamps         []int64
	randomNumberHashes []tmbytes.HexBytes
}

func (suite *AtomicSwapTestSuite) SetupTest() {
	// Generate 10 addresses
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
	_, addrs := app.GeneratePrivKeyAddressPairs(10)

	// Generate 10 timestamps and random number hashes
	var timestamps []int64
	var randomNumberHashes []tmbytes.HexBytes
	for i := 0; i < 10; i++ {
		timestamp := ts(i)
		randomNumber, _ := types.GenerateSecureRandomNumber()
		randomNumberHash := types.CalculateRandomHash(randomNumber.Bytes(), timestamp)
		timestamps = append(timestamps, timestamp)
		randomNumberHashes = append(randomNumberHashes, randomNumberHash)
	}

	suite.addrs = addrs
	suite.timestamps = timestamps
	suite.randomNumberHashes = randomNumberHashes
	return
}

func (suite *AtomicSwapTestSuite) TestNewAtomicSwap() {
	type args struct {
		amount              sdk.Coins
		randomNumberHash    tmbytes.HexBytes
		expireHeight        int64
		timestamp           int64
		sender              sdk.AccAddress
		recipient           sdk.AccAddress
		recipientOtherChain string
		senderOtherChain    string
		closedBlock         int64
		status              types.SwapStatus
		crossChain          bool
		direction           types.SwapDirection
	}
	testCases := []struct {
		description string
		args        args
		expectPass  bool
	}{
		{
			"normal",
			args{
				amount:              cs(c("bnb", 50000)),
				randomNumberHash:    suite.randomNumberHashes[0],
				expireHeight:        int64(360),
				timestamp:           suite.timestamps[0],
				sender:              suite.addrs[0],
				recipient:           suite.addrs[5],
				recipientOtherChain: "bnb1urfermcg92dwq36572cx4xg84wpk3lfpksr5g7",
				senderOtherChain:    "bnb1uky3me9ggqypmrsvxk7ur6hqkzq7zmv4ed4ng7",
				closedBlock:         1,
				status:              types.Open,
				crossChain:          true,
				direction:           types.Incoming,
			},
			true,
		},
		{
			"invalid random number hash length",
			args{
				amount:              cs(c("bnb", 50000)),
				randomNumberHash:    suite.randomNumberHashes[1][0:20],
				expireHeight:        int64(360),
				timestamp:           suite.timestamps[1],
				sender:              suite.addrs[1],
				recipient:           suite.addrs[5],
				recipientOtherChain: "bnb1urfermcg92dwq36572cx4xg84wpk3lfpksr5g7",
				senderOtherChain:    "bnb1uky3me9ggqypmrsvxk7ur6hqkzq7zmv4ed4ng7",
				closedBlock:         1,
				status:              types.Open,
				crossChain:          true,
				direction:           types.Incoming,
			},
			false,
		},
		{
			"invalid amount",
			args{
				amount:              cs(c("bnb", 0)),
				randomNumberHash:    suite.randomNumberHashes[2],
				expireHeight:        int64(360),
				timestamp:           suite.timestamps[2],
				sender:              suite.addrs[2],
				recipient:           suite.addrs[5],
				recipientOtherChain: "bnb1urfermcg92dwq36572cx4xg84wpk3lfpksr5g7",
				senderOtherChain:    "bnb1uky3me9ggqypmrsvxk7ur6hqkzq7zmv4ed4ng7",
				closedBlock:         1,
				status:              types.Open,
				crossChain:          true,
				direction:           types.Incoming,
			},
			false,
		},
	}

	for _, tc := range testCases {
		// Create atomic swap
		swap := types.NewAtomicSwap(tc.args.amount, tc.args.randomNumberHash, tc.args.expireHeight,
			tc.args.timestamp, tc.args.sender, tc.args.recipient, tc.args.senderOtherChain,
			tc.args.recipientOtherChain, tc.args.closedBlock, tc.args.status, tc.args.crossChain,
			tc.args.direction)

		if tc.expectPass {
			suite.Nil(swap.Validate())
			suite.Equal(tc.args.amount, swap.GetCoins())
			expectedSwapID := types.CalculateSwapID(tc.args.randomNumberHash, tc.args.sender, tc.args.senderOtherChain)
			suite.Equal(tmbytes.HexBytes(expectedSwapID), swap.GetSwapID())
		} else {
			suite.Error(swap.Validate())
		}

	}
}

func TestAtomicSwapTestSuite(t *testing.T) {
	suite.Run(t, new(AtomicSwapTestSuite))
}
