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
	TestSenderOtherChain        = "bnb1uky3me9ggqypmrsvxk7ur6hqkzq7zmv4ed4ng7"
	TestRecipientOtherChain     = "bnb1urfermcg92dwq36572cx4xg84wpk3lfpksr5g7"
	SwapLongtermStorageDuration = 86400
)

var (
	BNB_SUPPLY_LIMIT = i(100000000000)
)

func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }
func ts(minOffset int) int64                { return tmtime.Now().Add(time.Duration(minOffset) * time.Minute).Unix() }

func NewBep3GenStateMulti() app.GenesisState {
	bep3Genesis := baseGenState()
	return app.GenesisState{bep3.ModuleName: bep3.ModuleCdc.MustMarshalJSON(bep3Genesis)}
}

func baseGenState() bep3.GenesisState {
	// TODO: Set deputy to a reasonable address
	deputy, _ := sdk.AccAddressFromBech32("kava1xy7hrjy9r0algz9w3gzm8u6mrpq97kwta747gj")

	bep3Genesis := types.GenesisState{
		Params: bep3.Params{
			BnbDeputyAddress: deputy,
			MinBlockLock:     types.DefaultMinBlockLock, // 80
			MaxBlockLock:     types.DefaultMaxBlockLock, // 360
			SupportedAssets: types.AssetParams{
				types.AssetParam{
					Denom:  "bnb",
					CoinID: "714",            // TODO: This should be a number
					Limit:  BNB_SUPPLY_LIMIT, // TODO: Change limit increment time
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
	return bep3Genesis
}

func atomicSwapsWithAssetSupply(addrs []sdk.AccAddress, denom string) (types.AtomicSwaps, sdk.Coin) {
	var swaps types.AtomicSwaps
	assetSupply := c(denom, 0)
	for i := 0; i < len(addrs); i++ {
		assetSupply.Add(c(denom, 50000))
		swap := atomicSwapFromAddress(addrs[i], i)
		swaps = append(swaps, swap)
	}
	return swaps, assetSupply
}

func atomicSwapFromAddress(addr sdk.AccAddress, index int) types.AtomicSwap {
	expireOffset := int64((index * 15) + 360) // Default expire height + offet to match timestamp
	timestamp := ts(index)                    // One minute apart
	randomNumber, _ := types.GenerateSecureRandomNumber()
	randomNumberHash := types.CalculateRandomHash(randomNumber.Bytes(), timestamp)

	swap := types.NewAtomicSwap(cs(c("bnb", 50000)), randomNumberHash,
		expireOffset, timestamp, addr, addr, TestSenderOtherChain,
		TestRecipientOtherChain, 0, types.Open, true)

	return swap
}
