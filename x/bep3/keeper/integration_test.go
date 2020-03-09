package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/tendermint/tendermint/crypto"
	cmn "github.com/tendermint/tendermint/libs/common"

	// TODO: update alias
	"github.com/kava-labs/kava/x/bep3"
	"github.com/kava-labs/kava/x/bep3/types"
)

// TODO: Improve common testing variables
var (
	BNB_SUPPLY_LIMIT = i(100000000000)
	binanceAddrs     = []sdk.AccAddress{
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

func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }

func NewBep3GenStateMulti() app.GenesisState {
	deputy, _ := sdk.AccAddressFromBech32("kava1xy7hrjy9r0algz9w3gzm8u6mrpq97kwta747gj")

	bep3Genesis := types.GenesisState{
		Params: bep3.Params{
			BnbDeputyAddress: deputy,
			MinBlockLock:     types.DefaultMinBlockLock, // 80
			MaxBlockLock:     types.DefaultMaxBlockLock, // 360
			SupportedAssets: types.AssetParams{
				types.AssetParam{
					Denom:  "bnb",
					CoinID: "714",
					Limit:  BNB_SUPPLY_LIMIT,
					Active: true,
				},
				types.AssetParam{
					Denom:  "inc",
					CoinID: "9999",
					Limit:  i(100),
					Active: false,
				},
			},
		},
	}

	return app.GenesisState{bep3.ModuleName: bep3.ModuleCdc.MustMarshalJSON(bep3Genesis)}
}

func atomicSwaps(count int) types.AtomicSwaps {
	var atomicSwaps types.AtomicSwaps

	var swapIDs [][]byte
	for i := 0; i < count; i++ {
		swapID := types.CalculateSwapID(randomNumberHashes[i], binanceAddrs[i], "")
		swapIDs = append(swapIDs, swapID)
	}

	s1 := types.NewAtomicSwap(cs(c("bnb", 50000)), randomNumberHashes[0], int64(100), timestamps[0], binanceAddrs[0], kavaAddrs[0], "", "", 0, types.Open)
	s2 := types.NewAtomicSwap(cs(c("bnb", 50000)), randomNumberHashes[1], int64(275), timestamps[1], binanceAddrs[1], kavaAddrs[1], "", "", 0, types.Open)
	s3 := types.NewAtomicSwap(cs(c("bnb", 50000)), randomNumberHashes[2], int64(325), timestamps[2], binanceAddrs[2], kavaAddrs[2], "", "", 0, types.Open)
	s4 := types.NewAtomicSwap(cs(c("bnb", 50000)), randomNumberHashes[3], int64(500), timestamps[3], binanceAddrs[3], kavaAddrs[3], "", "", 0, types.Open)
	atomicSwaps = append(atomicSwaps, s1, s2, s3, s4)
	return atomicSwaps
}
