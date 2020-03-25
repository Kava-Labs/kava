package bep3_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3"
	"github.com/kava-labs/kava/x/bep3/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

const (
	TestSenderOtherChain    = "bnb1uky3me9ggqypmrsvxk7ur6hqkzq7zmv4ed4ng7"
	TestRecipientOtherChain = "bnb1urfermcg92dwq36572cx4xg84wpk3lfpksr5g7"
	TestDeputy              = "kava1xy7hrjy9r0algz9w3gzm8u6mrpq97kwta747gj"
	TestUser                = "kava1vry5lhegzlulehuutcr7nmdlmktw88awp0a39p"
)

var (
	StandardSupplyLimit = i(100000000000)
	DenomMap            = map[int]string{0: "btc", 1: "eth", 2: "bnb", 3: "xrp", 4: "dai"}
)

func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func d(de int64) sdk.Dec                    { return sdk.NewDec(de) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }
func ts(minOffset int) int64                { return tmtime.Now().Add(time.Duration(minOffset) * time.Minute).Unix() }

func NewBep3GenStateMulti(deputy sdk.AccAddress) app.GenesisState {
	bep3Genesis := baseGenState(deputy)
	return app.GenesisState{bep3.ModuleName: bep3.ModuleCdc.MustMarshalJSON(bep3Genesis)}
}

func baseGenState(deputy sdk.AccAddress) bep3.GenesisState {
	bep3Genesis := types.GenesisState{
		Params: bep3.Params{
			BnbDeputyAddress: deputy,
			MinBlockLock:     types.DefaultMinBlockLock, // 80
			MaxBlockLock:     types.DefaultMaxBlockLock, // 360
			SupportedAssets: types.AssetParams{
				types.AssetParam{
					Denom:  "btc",
					CoinID: 714,
					Limit:  StandardSupplyLimit,
					Active: true,
				},
				types.AssetParam{
					Denom:  "eth",
					CoinID: 999999,
					Limit:  StandardSupplyLimit,
					Active: true,
				},
				types.AssetParam{
					Denom:  "bnb",
					CoinID: 99999,
					Limit:  StandardSupplyLimit,
					Active: true,
				},
				types.AssetParam{
					Denom:  "inc",
					CoinID: 9999,
					Limit:  i(100),
					Active: false,
				},
			},
		},
	}
	return bep3Genesis
}

func loadSwapAndSupply(addr sdk.AccAddress, index int) (types.AtomicSwap, types.AssetSupply) {
	coin := c(DenomMap[index], 50000)
	expireOffset := int64((index * 15) + 360) // Default expire height + offet to match timestamp
	timestamp := ts(index)                    // One minute apart
	randomNumber, _ := types.GenerateSecureRandomNumber()
	randomNumberHash := types.CalculateRandomHash(randomNumber.Bytes(), timestamp)
	swap := types.NewAtomicSwap(cs(coin), randomNumberHash,
		expireOffset, timestamp, addr, addr, TestSenderOtherChain,
		TestRecipientOtherChain, 0, types.Open, true, types.Incoming)

	supply := types.NewAssetSupply(coin.Denom, coin, c(coin.Denom, 0),
		c(coin.Denom, 0), c(coin.Denom, StandardSupplyLimit.Int64()))

	return swap, supply
}
