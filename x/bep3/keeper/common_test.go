package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/bep3/types"
	"github.com/tendermint/tendermint/crypto"
	cmn "github.com/tendermint/tendermint/libs/common"
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
	timestamps         = []int64{100, 200, 300, 400}
	randomNumberHashes = []cmn.HexBytes{
		types.CalculateRandomHash([]byte("57047857647859512549395549808701232015920615785396990496251888108959386488324"), timestamps[0]),
		types.CalculateRandomHash([]byte("61225411119670325015452470011889923270088729812538562525047591229267896446077"), timestamps[1]),
		types.CalculateRandomHash([]byte("80338473265704256987314028010537813347969750625154662026470854854079495252215"), timestamps[2]),
		types.CalculateRandomHash([]byte("32833911140511546447866642984288167712598008093409038415228121205355103772318"), timestamps[3]),
	}
)
