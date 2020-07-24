package bep3_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3"
)

const (
	TestSenderOtherChain    = "bnb1uky3me9ggqypmrsvxk7ur6hqkzq7zmv4ed4ng7"
	TestRecipientOtherChain = "bnb1urfermcg92dwq36572cx4xg84wpk3lfpksr5g7"
	TestDeputy              = "kava1xy7hrjy9r0algz9w3gzm8u6mrpq97kwta747gj"
	TestUser                = "kava1vry5lhegzlulehuutcr7nmdlmktw88awp0a39p"
)

var (
	StandardSupplyLimit = i(100000000000)
	DenomMap            = map[int]string{0: "bnb", 1: "inc"}
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

	bep3Genesis := bep3.GenesisState{
		Params: bep3.Params{
			AssetParams: bep3.AssetParams{
				bep3.AssetParam{
					Denom:  "bnb",
					CoinID: 714,
					SupplyLimit: bep3.NewAssetSupply(
						sdk.NewCoin("bnb", sdk.ZeroInt()),
						sdk.NewCoin("bnb", sdk.ZeroInt()),
						sdk.NewCoin("bnb", sdk.ZeroInt()),
						sdk.NewCoin("bnb", sdk.NewInt(350000000000000))),
					Active:               true,
					DeputyAddress:        deputy,
					IncomingSwapFixedFee: sdk.NewInt(1000),
					MinSwapAmount:        sdk.OneInt(),
					MaxSwapAmount:        sdk.NewInt(1000000000000),
					MinBlockLock:         bep3.DefaultMinBlockLock,
					MaxBlockLock:         bep3.DefaultMaxBlockLock,
				},
				bep3.AssetParam{
					Denom:  "inc",
					CoinID: 9999,
					SupplyLimit: bep3.NewAssetSupply(
						sdk.NewCoin("inc", sdk.ZeroInt()),
						sdk.NewCoin("inc", sdk.ZeroInt()),
						sdk.NewCoin("inc", sdk.ZeroInt()),
						sdk.NewCoin("inc", sdk.NewInt(100000000000))),
					Active:               true,
					DeputyAddress:        deputy,
					IncomingSwapFixedFee: sdk.NewInt(1000),
					MinSwapAmount:        sdk.OneInt(),
					MaxSwapAmount:        sdk.NewInt(1000000000000),
					MinBlockLock:         bep3.DefaultMinBlockLock,
					MaxBlockLock:         bep3.DefaultMaxBlockLock,
				},
			},
		},
	}
	return bep3Genesis
}

func loadSwap(addr sdk.AccAddress, deputy sdk.AccAddress, index int) bep3.AtomicSwap {
	coin := c(DenomMap[index], 50000)
	expireOffset := bep3.DefaultMinBlockLock // Default expire height + offet to match timestamp
	timestamp := ts(index)                   // One minute apart
	randomNumber, _ := bep3.GenerateSecureRandomNumber()
	randomNumberHash := bep3.CalculateRandomHash(randomNumber[:], timestamp)
	swap := bep3.NewAtomicSwap(cs(coin), randomNumberHash,
		expireOffset, timestamp, deputy, addr, TestSenderOtherChain,
		TestRecipientOtherChain, 1, bep3.Open, true, bep3.Incoming)

	return swap
}
