package types_test

import (
	"github.com/stretchr/testify/suite"

	"github.com/tendermint/tendermint/crypto"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3/types"
)

var (
	coinsSingle  = sdk.NewCoins(sdk.NewInt64Coin("bnb", int64(50000)))
	coinsZero    = sdk.Coins{sdk.Coin{}}
	binanceAddrs = []string{
		sdk.AccAddress(crypto.AddressHash([]byte("BinanceTest1"))).String(),
		sdk.AccAddress(crypto.AddressHash([]byte("BinanceTest2"))).String(),
	}
	kavaAddrs = []string{
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1"))).String(),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest2"))).String(),
	}
	randomNumberBytes = []byte{15}
	timestampInt64    = int64(100)
	randomNumberHash  = types.CalculateRandomHash(randomNumberBytes, timestampInt64)
)

type MsgTestSuite struct {
	suite.Suite
}

func (suite *MsgTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
}

func (suite *MsgTestSuite) TestMsgCreateAtomicSwap() {
	tests := []struct {
		description         string
		from                string
		to                  string
		recipientOtherChain string
		senderOtherChain    string
		randomNumberHash    tmbytes.HexBytes
		timestamp           int64
		amount              sdk.Coins
		heightSpan          uint64
		expectPass          bool
	}{
		{"normal cross-chain", binanceAddrs[0], kavaAddrs[0], kavaAddrs[0], binanceAddrs[0], randomNumberHash, timestampInt64, coinsSingle, 500, true},
		{"without other chain fields", binanceAddrs[0], kavaAddrs[0], "", "", randomNumberHash, timestampInt64, coinsSingle, 500, false},
		{"invalid amount", binanceAddrs[0], kavaAddrs[0], "", "", randomNumberHash, timestampInt64, coinsZero, 500, false},
	}

	for i, tc := range tests {
		msg := types.NewMsgCreateAtomicSwap(
			tc.from,
			tc.to,
			tc.recipientOtherChain,
			tc.senderOtherChain,
			tc.randomNumberHash,
			tc.timestamp,
			tc.amount,
			tc.heightSpan,
		)
		if tc.expectPass {
			suite.NoError(msg.ValidateBasic(), "test: %v", i)
		} else {
			suite.Error(msg.ValidateBasic(), "test: %v", i)
		}
	}
}

func (suite *MsgTestSuite) TestMsgClaimAtomicSwap() {
	swapID := types.CalculateSwapID(randomNumberHash, binanceAddrs[0], "")

	tests := []struct {
		description  string
		from         string
		swapID       tmbytes.HexBytes
		randomNumber tmbytes.HexBytes
		expectPass   bool
	}{
		{"normal", binanceAddrs[0], swapID, randomNumberHash, true},
	}

	for i, tc := range tests {
		msg := types.NewMsgClaimAtomicSwap(
			tc.from,
			tc.swapID,
			tc.randomNumber,
		)
		if tc.expectPass {
			suite.NoError(msg.ValidateBasic(), "test: %v", i)
		} else {
			suite.Error(msg.ValidateBasic(), "test: %v", i)
		}
	}
}

func (suite *MsgTestSuite) TestMsgRefundAtomicSwap() {
	swapID := types.CalculateSwapID(randomNumberHash, binanceAddrs[0], "")

	tests := []struct {
		description string
		from        string
		swapID      tmbytes.HexBytes
		expectPass  bool
	}{
		{"normal", binanceAddrs[0], swapID, true},
	}

	for i, tc := range tests {
		msg := types.NewMsgRefundAtomicSwap(
			tc.from,
			tc.swapID,
		)
		if tc.expectPass {
			suite.NoError(msg.ValidateBasic(), "test: %v", i)
		} else {
			suite.Error(msg.ValidateBasic(), "test: %v", i)
		}
	}
}
