package types_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/x/bep3/types"
)

func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }
func ts(minOffset int) int64                { return tmtime.Now().Add(time.Duration(minOffset) * time.Minute).Unix() }

func atomicSwaps(count int) types.AtomicSwaps {
	var swaps types.AtomicSwaps
	for i := 0; i < count; i++ {
		swap := atomicSwap(i)
		swaps = append(swaps, swap)
	}
	return swaps
}

func atomicSwap(index int) types.AtomicSwap {
	expireOffset := uint64((index * 15) + 360) // Default expire height + offet to match timestamp
	timestamp := ts(index)                     // One minute apart
	randomNumber, _ := types.GenerateSecureRandomNumber()
	randomNumberHash := types.CalculateRandomHash(randomNumber, timestamp)

	swap := types.NewAtomicSwap(cs(c("bnb", 50000)), randomNumberHash, expireOffset, timestamp, kavaAddrs[0],
		kavaAddrs[1], binanceAddrs[0].String(), binanceAddrs[1].String(), 1, types.Open, true, types.Incoming)

	return swap
}
