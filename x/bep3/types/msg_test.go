package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
)

var (
	coinsSingle  = sdk.NewCoins(sdk.NewInt64Coin("bnb", int64(50000)))
	coinsZero    = sdk.Coins{sdk.Coin{}}
	binanceAddrs = []sdk.AccAddress{
		sdk.AccAddress(crypto.AddressHash([]byte("BinanceTest1"))),
		sdk.AccAddress(crypto.AddressHash([]byte("BinanceTest2"))),
	}
	kavaAddrs = []sdk.AccAddress{
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1"))),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest2"))),
	}
	randomNumberBytes = []byte{15}
	timestampInt64    = int64(9988776655)
	randomNumberHash  = BytesToHexEncodedString(CalculateRandomHash(randomNumberBytes, timestampInt64))
	ethAddrs          = []common.Address{
		common.HexToAddress("0x6f456B7F0b1658Be2683375159E7f09a8831CBe5"),
		common.HexToAddress("0x3a6CEef76Fd677332Dc0bA09604bD6acB1BeF613"),
	}
)

func TestHTLTMsg(t *testing.T) {
	tests := []struct {
		description         string
		from                sdk.AccAddress
		to                  sdk.AccAddress
		recipientOtherChain string
		senderOtherChain    string
		randomNumberHash    string
		timestamp           int64
		amount              sdk.Coins
		expectedIncome      string
		heightSpan          int64
		crossChain          bool
		expectPass          bool
	}{
		{"normal", binanceAddrs[0], kavaAddrs[0], "", "", randomNumberHash, timestampInt64, coinsSingle, "bnb50000", 80000, false, true},
		{"cross-chain", binanceAddrs[0], kavaAddrs[0], kavaAddrs[0].String(), binanceAddrs[0].String(), randomNumberHash, timestampInt64, coinsSingle, "bnb50000", 80000, true, true},
		{"with other chain fields", binanceAddrs[0], kavaAddrs[0], kavaAddrs[0].String(), binanceAddrs[0].String(), randomNumberHash, timestampInt64, coinsSingle, "bnb50000", 80000, false, false},
		{"cross-cross no other chain fields", binanceAddrs[0], kavaAddrs[0], "", "", randomNumberHash, timestampInt64, coinsSingle, "bnb50000", 80000, true, false},
		{"zero coins", binanceAddrs[0], kavaAddrs[0], "", "", randomNumberHash, timestampInt64, coinsZero, "bnb50000", 80000, true, false},
	}

	for i, tc := range tests {
		msg := NewHTLTMsg(
			tc.from,
			tc.to,
			tc.recipientOtherChain,
			tc.senderOtherChain,
			tc.randomNumberHash,
			tc.timestamp,
			tc.amount,
			tc.expectedIncome,
			tc.heightSpan,
			tc.crossChain,
		)
		if tc.expectPass {
			require.NoError(t, msg.ValidateBasic(), "test: %v", i)
		} else {
			require.Error(t, msg.ValidateBasic(), "test: %v", i)
		}
	}
}

func TestMsgDepositHTLT(t *testing.T) {
	swapIDBytes, _ := CalculateSwapID(randomNumberHash, binanceAddrs[0], "")
	swapID := BytesToHexEncodedString(swapIDBytes)

	tests := []struct {
		description string
		from        sdk.AccAddress
		swapID      string
		amount      sdk.Coins
		expectPass  bool
	}{
		{"normal", binanceAddrs[0], swapID, coinsSingle, true},
	}

	for i, tc := range tests {
		msg := NewMsgDepositHTLT(
			tc.from,
			tc.swapID,
			tc.amount,
		)
		if tc.expectPass {
			require.NoError(t, msg.ValidateBasic(), "test: %v", i)
		} else {
			require.Error(t, msg.ValidateBasic(), "test: %v", i)
		}
	}
}

func TestMsgClaimHTLT(t *testing.T) {
	swapIDBytes, _ := CalculateSwapID(randomNumberHash, binanceAddrs[0], "")
	swapID := BytesToHexEncodedString(swapIDBytes)

	tests := []struct {
		description  string
		from         sdk.AccAddress
		swapID       string
		randomNumber SwapBytes
		expectPass   bool
	}{
		{"normal", binanceAddrs[0], swapID, randomNumberBytes, true},
	}

	for i, tc := range tests {
		msg := NewMsgClaimHTLT(
			tc.from,
			tc.swapID,
			tc.randomNumber,
		)
		if tc.expectPass {
			require.NoError(t, msg.ValidateBasic(), "test: %v", i)
		} else {
			require.Error(t, msg.ValidateBasic(), "test: %v", i)
		}
	}
}

func TestMsgRefundHTLT(t *testing.T) {
	swapIDBytes, _ := CalculateSwapID(randomNumberHash, binanceAddrs[0], "")
	swapID := BytesToHexEncodedString(swapIDBytes)

	tests := []struct {
		description string
		from        sdk.AccAddress
		swapID      string
		expectPass  bool
	}{
		{"normal", binanceAddrs[0], swapID, true},
	}

	for i, tc := range tests {
		msg := NewMsgRefundHTLT(
			tc.from,
			tc.swapID,
		)
		if tc.expectPass {
			require.NoError(t, msg.ValidateBasic(), "test: %v", i)
		} else {
			require.Error(t, msg.ValidateBasic(), "test: %v", i)
		}
	}
}
