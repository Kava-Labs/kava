package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3"
	"github.com/kava-labs/kava/x/bep3/types"
	"github.com/tendermint/tendermint/crypto"
	tmtime "github.com/tendermint/tendermint/types/time"
)

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
)

func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }
func ts(minOffset int) int64                { return tmtime.Now().Add(time.Duration(minOffset) * time.Minute).Unix() }

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

func atomicSwaps(ctx sdk.Context, count int) types.AtomicSwaps {
	var swaps types.AtomicSwaps
	for i := 0; i < count; i++ {
		swap := atomicSwap(ctx, i)
		swaps = append(swaps, swap)
	}
	return swaps
}

func atomicSwap(ctx sdk.Context, index int) types.AtomicSwap {
	expireOffset := int64((index * 15) + 360) // Default expire height + offet to match timestamp
	timestamp := ts(index)                    // One minute apart
	randomNumber, _ := types.GenerateSecureRandomNumber()
	randomNumberHash := types.CalculateRandomHash(randomNumber.Bytes(), timestamp)

	swap := types.NewAtomicSwap(cs(c("bnb", 50000)), randomNumberHash,
		ctx.BlockHeight()+expireOffset, timestamp, kavaAddrs[0], kavaAddrs[1],
		binanceAddrs[0].String(), binanceAddrs[1].String(), 0, types.Open, true)

	return swap
}
