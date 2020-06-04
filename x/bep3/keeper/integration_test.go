package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/tendermint/crypto"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3"
	"github.com/kava-labs/kava/x/bep3/types"
)

const (
	TestSenderOtherChain    = "bnb1uky3me9ggqypmrsvxk7ur6hqkzq7zmv4ed4ng7"
	TestRecipientOtherChain = "bnb1urfermcg92dwq36572cx4xg84wpk3lfpksr5g7"
	TestDeputy              = "kava1xy7hrjy9r0algz9w3gzm8u6mrpq97kwta747gj"
)

var (
	StandardSupplyLimit = i(350000000000000)
	DenomMap            = map[int]string{0: "btc", 1: "eth", 2: "bnb", 3: "xrp", 4: "dai"}
	TestUser1           = sdk.AccAddress(crypto.AddressHash([]byte("KavaTestUser1")))
	TestUser2           = sdk.AccAddress(crypto.AddressHash([]byte("KavaTestUser2")))
)

func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }
func ts(minOffset int) int64                { return tmtime.Now().Add(time.Duration(minOffset) * time.Minute).Unix() }

func NewBep3GenStateMulti(deputyAddress sdk.AccAddress) app.GenesisState {
	bep3Genesis := types.GenesisState{
		Params: bep3.Params{
			BnbDeputyAddress:  deputyAddress,
			BnbDeputyFixedFee: types.DefaultBnbDeputyFixedFee, // 1000
			MinAmount:         types.DefaultMinAmount,         // 0
			MaxAmount:         types.DefaultMaxAmount,         // 10,000
			MinBlockLock:      types.DefaultMinBlockLock,      // 220
			MaxBlockLock:      types.DefaultMaxBlockLock,      // 270
			SupportedAssets: types.AssetParams{
				types.AssetParam{
					Denom:  "bnb",
					CoinID: 714,
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
	expireOffset := uint64(200) // Default expire height + offet to match timestamp
	timestamp := ts(index)      // One minute apart
	randomNumber, _ := types.GenerateSecureRandomNumber()
	randomNumberHash := types.CalculateRandomHash(randomNumber[:], timestamp)

	return types.NewAtomicSwap(cs(c("bnb", 50000)), randomNumberHash,
		uint64(ctx.BlockHeight())+expireOffset, timestamp, TestUser1, TestUser2,
		TestSenderOtherChain, TestRecipientOtherChain, 0, types.Open, true,
		types.Incoming)
}

func assetSupplies(count int) types.AssetSupplies {
	if count > 5 { // Max 5 asset supplies
		return types.AssetSupplies{}
	}

	var supplies types.AssetSupplies

	for i := 0; i < count; i++ {
		supply := assetSupply(DenomMap[i])
		supplies = append(supplies, supply)
	}
	return supplies
}

func assetSupply(denom string) types.AssetSupply {
	return types.NewAssetSupply(denom, c(denom, 0), c(denom, 0), c(denom, 0), c(denom, 10000))
}
