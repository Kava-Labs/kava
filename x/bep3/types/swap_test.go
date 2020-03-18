package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3/types"
	"github.com/stretchr/testify/suite"
	cmn "github.com/tendermint/tendermint/libs/common"
)

type AtomicSwapTestSuite struct {
	suite.Suite
	addrs              []sdk.AccAddress
	timestamps         []int64
	randomNumberHashes []cmn.HexBytes
}

func (suite *AtomicSwapTestSuite) SetupTest() {
	// Generate 10 addresses
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
	_, addrs := app.GeneratePrivKeyAddressPairs(10)

	// Generate 10 timestamps and random number hashes
	var timestamps []int64
	var randomNumberHashes []cmn.HexBytes
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
		randomNumberHash    cmn.HexBytes
		expireHeight        int64
		timestamp           int64
		sender              sdk.AccAddress
		recipient           sdk.AccAddress
		recipientOtherChain string
		senderOtherChain    string
		closedBlock         int64
		status              types.SwapStatus
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
				closedBlock:         0,
				status:              types.Open,
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
				closedBlock:         0,
				status:              types.Open,
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
				closedBlock:         0,
				status:              types.Open,
			},
			false,
		},
	}

	for _, tc := range testCases {
		// Create atomic swap
		swap := types.NewAtomicSwap(tc.args.amount, tc.args.randomNumberHash, tc.args.expireHeight,
			tc.args.timestamp, tc.args.sender, tc.args.recipient, tc.args.senderOtherChain,
			tc.args.recipientOtherChain, tc.args.closedBlock, tc.args.status)

		if tc.expectPass {
			suite.Nil(swap.Validate())
			suite.Equal(tc.args.amount, swap.GetModuleAccountCoins())
			expectedSwapID := types.CalculateSwapID(tc.args.randomNumberHash, tc.args.sender, tc.args.senderOtherChain)
			suite.Equal(cmn.HexBytes(expectedSwapID), swap.GetSwapID())
		} else {
			suite.Error(swap.Validate())
		}

	}
}

func TestAtomicSwapTestSuite(t *testing.T) {
	suite.Run(t, new(AtomicSwapTestSuite))
}
