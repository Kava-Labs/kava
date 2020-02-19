package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/bep3/types"
	"github.com/tendermint/tendermint/crypto"
)

var (
	coinsSingle  = sdk.NewCoins(sdk.NewInt64Coin("bnb", int64(50000)))
	coinsZero    = sdk.Coins{sdk.Coin{}}
	binanceAddrs = []sdk.AccAddress{
		sdk.AccAddress(crypto.AddressHash([]byte("BinanceTest1"))),
		sdk.AccAddress(crypto.AddressHash([]byte("BinanceTest2"))),
		sdk.AccAddress(crypto.AddressHash([]byte("BinanceTest3"))),
		sdk.AccAddress(crypto.AddressHash([]byte("BinanceTest4"))),
	}
	kavaAddrs = []sdk.AccAddress{
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1"))),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest2"))),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest3"))),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest4"))),
	}
	timestamps         = []int64{6655443322, 7766554433, 8877665544, 9988776655}
	randomNumberHashes = [][]byte{
		types.CalculateRandomHash([]byte{15}, timestamps[0]),
		types.CalculateRandomHash([]byte{72}, timestamps[1]),
		types.CalculateRandomHash([]byte{119}, timestamps[2]),
		types.CalculateRandomHash([]byte{154}, timestamps[3]),
	}
)
