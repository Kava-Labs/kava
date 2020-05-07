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
	testCases := []struct {
		msg        string
		swap       types.AtomicSwap
		expectPass bool
	}{
		{
			"valid Swap",
			types.AtomicSwap{
				Amount:              cs(c("bnb", 50000)),
				RandomNumberHash:    suite.randomNumberHashes[0],
				ExpireHeight:        360,
				Timestamp:           suite.timestamps[0],
				Sender:              suite.addrs[0],
				Recipient:           suite.addrs[5],
				RecipientOtherChain: "bnb1urfermcg92dwq36572cx4xg84wpk3lfpksr5g7",
				SenderOtherChain:    "bnb1uky3me9ggqypmrsvxk7ur6hqkzq7zmv4ed4ng7",
				ClosedBlock:         1,
				Status:              types.Open,
				CrossChain:          true,
				Direction:           types.Incoming,
			},
			true,
		},
		{
			"invalid amount",
			types.AtomicSwap{
				Amount: sdk.Coins{sdk.Coin{Denom: "BNB", Amount: sdk.NewInt(10)}},
			},
			false,
		},
		{
			"amount not positive",
			types.AtomicSwap{
				Amount: cs(c("bnb", 0)),
			},
			false,
		},
		{
			"invalid random number hash length",
			types.AtomicSwap{
				Amount:           cs(c("bnb", 50000)),
				RandomNumberHash: suite.randomNumberHashes[1][0:20],
			},
			false,
		},
		{
			"exp height 0",
			types.AtomicSwap{
				Amount:           cs(c("bnb", 50000)),
				RandomNumberHash: suite.randomNumberHashes[0],
				ExpireHeight:     0,
			},
			false,
		},
		{
			"timestamp 0",
			types.AtomicSwap{
				Amount:           cs(c("bnb", 50000)),
				RandomNumberHash: suite.randomNumberHashes[0],
				ExpireHeight:     10,
				Timestamp:        0,
			},
			false,
		},
		{
			"empty sender",
			types.AtomicSwap{
				Amount:           cs(c("bnb", 50000)),
				RandomNumberHash: suite.randomNumberHashes[0],
				ExpireHeight:     10,
				Timestamp:        10,
				Sender:           nil,
			},
			false,
		},
		{
			"empty recipient",
			types.AtomicSwap{
				Amount:           cs(c("bnb", 50000)),
				RandomNumberHash: suite.randomNumberHashes[0],
				ExpireHeight:     10,
				Timestamp:        10,
				Sender:           suite.addrs[0],
				Recipient:        nil,
			},
			false,
		},
		{
			"invalid sender length",
			types.AtomicSwap{
				Amount:           cs(c("bnb", 50000)),
				RandomNumberHash: suite.randomNumberHashes[0],
				ExpireHeight:     10,
				Timestamp:        10,
				Sender:           suite.addrs[0][:10],
				Recipient:        suite.addrs[5],
			},
			false,
		},
		{
			"invalid recipient length",
			types.AtomicSwap{
				Amount:           cs(c("bnb", 50000)),
				RandomNumberHash: suite.randomNumberHashes[0],
				ExpireHeight:     10,
				Timestamp:        10,
				Sender:           suite.addrs[0],
				Recipient:        suite.addrs[5][:10],
			},
			false,
		},
		{
			"invalid sender other chain",
			types.AtomicSwap{
				Amount:           cs(c("bnb", 50000)),
				RandomNumberHash: suite.randomNumberHashes[0],
				ExpireHeight:     10,
				Timestamp:        10,
				Sender:           suite.addrs[0],
				Recipient:        suite.addrs[5],
				SenderOtherChain: "",
			},
			false,
		},
		{
			"invalid recipient other chain",
			types.AtomicSwap{
				Amount:              cs(c("bnb", 50000)),
				RandomNumberHash:    suite.randomNumberHashes[0],
				ExpireHeight:        10,
				Timestamp:           10,
				Sender:              suite.addrs[0],
				Recipient:           suite.addrs[5],
				SenderOtherChain:    "bnb1uky3me9ggqypmrsvxk7ur6hqkzq7zmv4ed4ng7",
				RecipientOtherChain: "",
			},
			false,
		},
		{
			"closed block 0",
			types.AtomicSwap{
				Amount:              cs(c("bnb", 50000)),
				RandomNumberHash:    suite.randomNumberHashes[0],
				ExpireHeight:        10,
				Timestamp:           10,
				Sender:              suite.addrs[0],
				Recipient:           suite.addrs[5],
				SenderOtherChain:    "bnb1uky3me9ggqypmrsvxk7ur6hqkzq7zmv4ed4ng7",
				RecipientOtherChain: "bnb1urfermcg92dwq36572cx4xg84wpk3lfpksr5g7",
				ClosedBlock:         0,
			},
			false,
		},
		{
			"invalid status 0",
			types.AtomicSwap{
				Amount:              cs(c("bnb", 50000)),
				RandomNumberHash:    suite.randomNumberHashes[0],
				ExpireHeight:        10,
				Timestamp:           10,
				Sender:              suite.addrs[0],
				Recipient:           suite.addrs[5],
				SenderOtherChain:    "bnb1uky3me9ggqypmrsvxk7ur6hqkzq7zmv4ed4ng7",
				RecipientOtherChain: "bnb1urfermcg92dwq36572cx4xg84wpk3lfpksr5g7",
				ClosedBlock:         1,
				Status:              types.NULL,
			},
			false,
		},
		{
			"invalid direction ",
			types.AtomicSwap{
				Amount:              cs(c("bnb", 50000)),
				RandomNumberHash:    suite.randomNumberHashes[0],
				ExpireHeight:        10,
				Timestamp:           10,
				Sender:              suite.addrs[0],
				Recipient:           suite.addrs[5],
				SenderOtherChain:    "bnb1uky3me9ggqypmrsvxk7ur6hqkzq7zmv4ed4ng7",
				RecipientOtherChain: "bnb1urfermcg92dwq36572cx4xg84wpk3lfpksr5g7",
				ClosedBlock:         1,
				Status:              types.Open,
				Direction:           types.INVALID,
			},
			false,
		},
	}

	for _, tc := range testCases {

		err := tc.swap.Validate()
		if tc.expectPass {
			suite.Require().NoError(err, tc.msg)
			suite.Require().Equal(tc.swap.Amount, tc.swap.GetCoins())

			expectedSwapID := types.CalculateSwapID(tc.swap.RandomNumberHash, tc.swap.Sender, tc.swap.SenderOtherChain)
			suite.Require().Equal(tmbytes.HexBytes(expectedSwapID), tc.swap.GetSwapID())
		} else {
			suite.Require().Error(err)
		}

	}
}

func TestAtomicSwapTestSuite(t *testing.T) {
	suite.Run(t, new(AtomicSwapTestSuite))
}
