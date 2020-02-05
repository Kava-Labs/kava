package types

import (
	"testing"

	"github.com/binance-chain/go-sdk/common/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
)

var (
	coinsSingle  = types.Coins{types.Coin{Denom: "bnb", Amount: 50000}}
	coinsZero    = types.Coins{types.Coin{}}
	binanceAddrs = []types.AccAddress{
		types.AccAddress(crypto.AddressHash([]byte("BinanceTest1"))),
		types.AccAddress(crypto.AddressHash([]byte("BinanceTest2"))),
	}
	kavaAddrs = []types.AccAddress{
		types.AccAddress(crypto.AddressHash([]byte("KavaTest1"))),
		types.AccAddress(crypto.AddressHash([]byte("KavaTest2"))),
	}
	randomNumberBytes = []byte{15}
	timestampInt64    = int64(9988776655)
	randomNumberHash  = CalculateRandomHash(randomNumberBytes, timestampInt64)
)

func TestMsgCreateHTLT(t *testing.T) {
	tests := []struct {
		description         string
		from                types.AccAddress
		to                  types.AccAddress
		recipientOtherChain string
		senderOtherChain    string
		randomNumberHash    types.SwapBytes
		timestamp           int64
		amount              types.Coins
		expectedIncome      string
		heightSpan          int64
		crossChain          bool
		expectPass          bool
	}{
		{"create htlt", binanceAddrs[0], kavaAddrs[0], "", "", randomNumberHash, timestampInt64, coinsSingle, "bnb50000", 80000, false, true},
		{"create htlt cross-chain", binanceAddrs[0], kavaAddrs[0], kavaAddrs[0].String(), binanceAddrs[0].String(), randomNumberHash, timestampInt64, coinsSingle, "bnb50000", 80000, true, true},
		{"create htlt with other chain fields", binanceAddrs[0], kavaAddrs[0], kavaAddrs[0].String(), binanceAddrs[0].String(), randomNumberHash, timestampInt64, coinsSingle, "bnb50000", 80000, false, false},
		{"create htlt cross-cross no other chain fields", binanceAddrs[0], kavaAddrs[0], "", "", randomNumberHash, timestampInt64, coinsSingle, "bnb50000", 80000, true, false},
		{"create htlt zero coins", binanceAddrs[0], kavaAddrs[0], "", "", randomNumberHash, timestampInt64, coinsZero, "bnb50000", 80000, true, false},
	}

	for i, tc := range tests {
		msg := NewMsgCreateHTLT(
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
