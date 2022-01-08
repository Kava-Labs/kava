package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3/types"
)

var (
	coinsSingle       = sdk.NewCoins(sdk.NewInt64Coin("bnb", 50000))
	binanceAddrs      = []sdk.AccAddress{}
	kavaAddrs         = []sdk.AccAddress{}
	randomNumberBytes = []byte{15}
	timestampInt64    = int64(100)
	randomNumberHash  = tmbytes.HexBytes(types.CalculateRandomHash(randomNumberBytes, timestampInt64))
)

func init() {
	app.SetSDKConfig()

	// Must be set after SetSDKConfig to use kava Bech32 prefix instead of cosmos
	binanceAddrs = []sdk.AccAddress{
		sdk.AccAddress(crypto.AddressHash([]byte("BinanceTest1"))),
		sdk.AccAddress(crypto.AddressHash([]byte("BinanceTest2"))),
	}
	kavaAddrs = []sdk.AccAddress{
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1"))),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest2"))),
	}
}

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
		from                sdk.AccAddress
		to                  sdk.AccAddress
		recipientOtherChain string
		senderOtherChain    string
		randomNumberHash    string
		timestamp           int64
		amount              sdk.Coins
		heightSpan          uint64
		expectPass          bool
	}{
		{"normal cross-chain", binanceAddrs[0], kavaAddrs[0], kavaAddrs[0].String(), binanceAddrs[0].String(), randomNumberHash.String(), timestampInt64, coinsSingle, 500, true},
		{"without other chain fields", binanceAddrs[0], kavaAddrs[0], "", "", randomNumberHash.String(), timestampInt64, coinsSingle, 500, false},
		{"invalid amount", binanceAddrs[0], kavaAddrs[0], kavaAddrs[0].String(), binanceAddrs[0].String(), randomNumberHash.String(), timestampInt64, nil, 500, false},
		{"invalid from address", sdk.AccAddress{}, kavaAddrs[0], kavaAddrs[0].String(), binanceAddrs[0].String(), randomNumberHash.String(), timestampInt64, coinsSingle, 500, false},
		{"invalid to address", binanceAddrs[0], sdk.AccAddress{}, kavaAddrs[0].String(), binanceAddrs[0].String(), randomNumberHash.String(), timestampInt64, coinsSingle, 500, false},
		{"invalid rand hash", binanceAddrs[0], kavaAddrs[0], kavaAddrs[0].String(), binanceAddrs[0].String(), "ff", timestampInt64, coinsSingle, 500, false},
	}

	for i, tc := range tests {
		msg := types.MsgCreateAtomicSwap{
			tc.from.String(),
			tc.to.String(),
			tc.recipientOtherChain,
			tc.senderOtherChain,
			tc.randomNumberHash,
			tc.timestamp,
			tc.amount,
			tc.heightSpan,
		}
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
		from         sdk.AccAddress
		swapID       tmbytes.HexBytes
		randomNumber tmbytes.HexBytes
		expectPass   bool
	}{
		{"normal", binanceAddrs[0], swapID, randomNumberHash, true},
		{"invalid from address", sdk.AccAddress{}, swapID, randomNumberHash, false},
	}

	for i, tc := range tests {
		msg := types.NewMsgClaimAtomicSwap(
			tc.from.String(),
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
		from        sdk.AccAddress
		swapID      tmbytes.HexBytes
		expectPass  bool
	}{
		{"normal", binanceAddrs[0], swapID, true},
		{"invalid from address", sdk.AccAddress{}, swapID, false},
	}

	for i, tc := range tests {
		msg := types.NewMsgRefundAtomicSwap(
			tc.from.String(),
			tc.swapID,
		)
		if tc.expectPass {
			suite.NoError(msg.ValidateBasic(), "test: %v", i)
		} else {
			suite.Error(msg.ValidateBasic(), "test: %v", i)
		}
	}
}

func TestMsgTestSuite(t *testing.T) {
	suite.Run(t, new(MsgTestSuite))
}
